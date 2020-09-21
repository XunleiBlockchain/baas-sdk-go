package types

import (
	"math/big"
	"sync/atomic"

	"github.com/XunleiBlockchain/tc-libs/common"
	"github.com/XunleiBlockchain/tc-libs/crypto"
	"github.com/XunleiBlockchain/tc-libs/types"
)

// stdSigCache is used to cache the derived sender and contains
// the signer used to derive it.
type stdSigCache struct {
	signer types.STDSigner
	from   common.Address
}

// MakeSTDSigner returns a STDSigner based on the given signparam or default param.
func MakeSTDSigner(signParam *big.Int) types.STDSigner {
	return types.NewSTDEIP155Signer(signParam)
}

var _ types.SignerData = (*signdata)(nil)
var big8 = big.NewInt(8)

type signdata struct {
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	fromValue atomic.Value
	// lock           sync.Mutex
	signFieldsFunc func() []interface{}
}

// Sign signs the signdata using the given signer and private key
func sign(signer types.STDSigner, prv crypto.PrivKey, data []interface{}) (*big.Int, *big.Int, *big.Int, error) {
	fields := append(data, signer.SignParam(), uint(0), uint(0))
	h := types.BalHash(fields)

	sig, err := prv.Sign(h[:])
	if err != nil {
		return nil, nil, nil, err
	}

	return signer.SignatureValues(sig.Raw())
}

// VerifySign The signature should be in [R || S || V] format.
func VerifySign(signer types.STDSigner, prv []byte, data []interface{}, sign []byte) bool {
	fields := append(data, signer.SignParam(), uint(0), uint(0))
	h := types.BalHash(fields)
	return crypto.VerifySignature(prv, h[:], sign)
}

// Sender returns the address derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// Sender may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func sender(signer types.STDSigner, data types.SignerData) (common.Address, error) {
	if sc := data.From().Load(); sc != nil {
		sigCache := sc.(stdSigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	addr, err := signer.Sender(data)
	if err != nil {
		return common.EmptyAddress, err
	}
	data.From().Store(stdSigCache{signer: signer, from: addr})
	return addr, nil
}

func senders(signer types.STDSigner, signatures []*signdata, signFields func() []interface{}) ([]common.Address, error) {
	addrs := make([]common.Address, 0, len(signatures))
	for _, s := range signatures {
		s.setSignFieldsFunc(signFields)
		addr, err := sender(signer, s)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func isProtectedV(V *big.Int) bool {
	if V != nil && V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

// Protected returns whether the transaction is protected from replay protection.
func (data signdata) Protected() bool {
	return isProtectedV(data.V)
}

func (data *signdata) From() *atomic.Value {
	return &data.fromValue
}

func (data *signdata) setSignFieldsFunc(signFields func() []interface{}) {
	// data.lock.Lock()
	data.signFieldsFunc = signFields
	// data.lock.Unlock()
}

func (data signdata) SignFields() []interface{} {
	return data.signFieldsFunc()
}

func (data signdata) Recover(hash common.Hash, signParamMul *big.Int, homestead bool) (common.Address, error) {
	if signParamMul == nil {
		return recoverPlain(hash, data.R, data.S, data.V, homestead)
	}
	V := new(big.Int).Sub(data.V, signParamMul)
	V.Sub(V, big8)
	return recoverPlain(hash, data.R, data.S, V, homestead)
}

// SignParam returns which sign param this transaction was signed with
func (data signdata) SignParam() *big.Int {
	return deriveSignParam(data.V)
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
	if Vb.BitLen() > 8 {
		return common.EmptyAddress, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return common.EmptyAddress, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V

	return crypto.Sender(sighash[:], sig)
}

// deriveSignParam derives the sign param from the given v parameter
func deriveSignParam(v *big.Int) *big.Int {
	if v == nil {
		return big.NewInt(0)
	}
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
