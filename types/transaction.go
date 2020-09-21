package types

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"sync/atomic"

	"github.com/XunleiBlockchain/tc-libs/bal"
	"github.com/XunleiBlockchain/tc-libs/common"
	"github.com/XunleiBlockchain/tc-libs/common/hexutil"
	"github.com/XunleiBlockchain/tc-libs/crypto"
	"github.com/XunleiBlockchain/tc-libs/types"
)

//###//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

// TxData represent a tx content
type Tx interface {
	Hash() common.Hash
	From() (common.Address, error)
	To() *common.Address
	TypeName() string
}

func RegisterTxData() {
	bal.RegisterInterface((*Tx)(nil), nil)
	bal.RegisterConcrete(&Transaction{}, TxNormal, nil)
}

var _ Tx = (*Transaction)(nil)
var _ types.SignerData = (*txdata)(nil)

type Transaction struct {
	data txdata
	// caches
	hash       atomic.Value
	size       atomic.Value
	from       atomic.Value
	fromZoneID atomic.Value
	toZoneID   atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       bal:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	ttl int8

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	fromValue atomic.Value

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" bal:"-"`
}

// Protected returns whether the transaction is protected from replay protection.
func (data txdata) Protected() bool {
	return isProtectedV(data.V)
}

func (data *txdata) From() *atomic.Value {
	return &data.fromValue
}

func (data txdata) SignFields() []interface{} {
	return []interface{}{
		data.AccountNonce,
		data.Price,
		data.GasLimit,
		data.Recipient,
		data.Amount,
		data.Payload,
	}
}

func (data txdata) Recover(hash common.Hash, signParamMul *big.Int, homestead bool) (common.Address, error) {
	if signParamMul == nil {
		return recoverPlain(hash, data.R, data.S, data.V, homestead)
	}
	V := new(big.Int).Sub(data.V, signParamMul)
	V.Sub(V, big8)
	return recoverPlain(hash, data.R, data.S, V, homestead)
}

// SignParam returns which sign param this transaction was signed with
func (data txdata) SignParam() *big.Int {
	return deriveSignParam(data.V)
}

func (tx *Transaction) Sign(signer types.STDSigner, prv crypto.PrivKey) error {
	r, s, v, err := sign(signer, prv, tx.data.SignFields())
	if err != nil {
		return err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	*tx = *cpy
	return nil
}

func (tx *Transaction) VirifySign(signer types.STDSigner, prv []byte) bool {
	sign := make([]byte, 65)
	rlen := len(tx.data.R.Bytes())
	slen := len(tx.data.S.Bytes())

	copy(sign[32-rlen:32], tx.data.R.Bytes()[:])
	copy(sign[64-slen:64], tx.data.S.Bytes()[:])

	var V byte
	if isProtectedV(tx.data.V) {
		signParam := deriveSignParam(tx.data.V).Uint64()
		V = byte(tx.data.V.Uint64() - 35 - 2*signParam)
	} else {
		V = byte(tx.data.V.Uint64() - 27)
	}
	sign[64] = V
	return VerifySign(signer, prv, tx.data.SignFields(), sign)
}
func (tx *Transaction) balHahs() {
	types.BalHash(tx.data)
}
func (tx *Transaction) GetTtl() int8 {
	return tx.data.ttl
}
func (tx *Transaction) AddTtl() {
	tx.data.ttl++
}
func (tx *Transaction) SignHash() common.Hash {
	return GlobalSTDSigner.Hash(&tx.data)
}

func (tx *Transaction) Sender(signer types.STDSigner) (common.Address, error) {
	return sender(signer, &tx.data)
}

// SignParam returns which sign param this transaction was signed with
func (tx *Transaction) SignParam() *big.Int {
	return tx.data.SignParam()
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *Transaction) Protected() bool {
	return tx.data.Protected()
}

func (tx *Transaction) TypeName() string {
	return TxNormal
}

func (tx *Transaction) From() (common.Address, error) {
	return tx.Sender(GlobalSTDSigner)
}

func (tx *Transaction) StoreFrom(addr *common.Address) {
	tx.data.From().Store(stdSigCache{signer: GlobalSTDSigner, from: *addr})
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	GasLimit     hexutil.Uint64
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
}

func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

func TestNewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	tx := newTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
	tx.data.GasLimit, tx.data.Price = gasLimit, gasPrice
	return tx
}

func TestNewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	ctx := newTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
	ctx.data.GasLimit, ctx.data.Price = gasLimit, gasPrice
	return ctx
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	// Need to set fixed gasPrice = 1e11
	// The non-contract gasLimit would be equal to gasUsed, so gasLimit * gasPrice = 0.01 (eth)

	// TODO @binacs 此处对于最新的底层链GasPrice是0 此处注释应取消
	// gasPrice = big.NewInt(ParGasPrice)
	if !IsContract(data) {
		d.GasLimit = ParGasLimit
	}

	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &Transaction{data: d}
}

func IsContract(data []byte) bool {
	if len(data) > 0 {
		var extra map[string]interface{}
		err := json.Unmarshal(data, &extra)
		if err == nil {
			// third party json payload
			return false
		}
		return true
	}
	return false
}

func (tx *Transaction) IllegalGasLimitOrGasPrice() bool {
	if tx.GasPrice().Sign() < 0 {
		fmt.Printf("tx.GasPrice().Sign() < 0:  sign = %+v\n", tx.GasPrice().Sign())
		return true
	}

	if tx.GasPrice().Cmp(big.NewInt(ParGasPrice)) != 0 {
		fmt.Printf("tx.GasPrice().Cmp(big.NewInt(ParGasPrice)) != 0:  price = %+v\n", tx.GasPrice())
		return true
	}

	if !IsContract(tx.Data()) && tx.Gas() != ParGasLimit {
		fmt.Printf("!IsContract(tx.Data()) && tx.Gas() != ParGasLimit:  !IsContract(tx.Data()) = %+v\n", !IsContract(tx.Data()), "tx.Gas()", tx.Gas(), "ParGasLimit", ParGasLimit)
		return true
	}
	fmt.Printf("false!\n")
	return false
}

// EncodeBAL implements bal.Encoder
func (tx *Transaction) EncodeBAL(w io.Writer) error {
	return bal.Encode(w, &tx.data)
}

// DecodeBAL implements bal.Decoder
func (tx *Transaction) DecodeBAL(s *bal.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(bal.ListSize(size)))
	}

	return err
}

// Encodebal implements bal.Encoder
func (tx *Transaction) Encodebal(w io.Writer) error {
	return bal.Encode(w, &tx.data)
}

// Decodebal implements bal.Decoder
func (tx *Transaction) Decodebal(s *bal.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(bal.ListSize(size)))
	}

	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		signParam := deriveSignParam(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*signParam)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return ErrInvalidSig
	}
	*tx = Transaction{data: dec}
	return nil
}

func (tx Transaction) GetTxData() txdata             { return tx.data }
func (tx *Transaction) TokenAddress() common.Address { return common.EmptyAddress }
func (tx *Transaction) Data() []byte                 { return common.CopyBytes(tx.data.Payload) }
func (tx *Transaction) Gas() uint64                  { return tx.data.GasLimit }
func (tx *Transaction) GasPrice() *big.Int           { return new(big.Int).Set(tx.data.Price) }
func (tx *Transaction) Value() *big.Int              { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Nonce() uint64                { return tx.data.AccountNonce }
func (tx *Transaction) CheckNonce() bool             { return true }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the bal encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := types.BalHash(tx)
	tx.hash.Store(v)
	return v
}

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}

// Size returns the true bal encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	bal.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *Transaction) AsMessage() (Message, error) {
	msg := Message{
		nonce:        tx.data.AccountNonce,
		gasLimit:     tx.data.GasLimit,
		gasPrice:     new(big.Int).Set(tx.data.Price),
		to:           tx.data.Recipient,
		amount:       tx.data.Amount,
		data:         tx.data.Payload,
		checkNonce:   true,
		tokenAddr:    common.Address{},
		contractZone: -1,
		txType:       TxNormal,
	}

	var err error
	msg.from, err = tx.From()
	return msg, err
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer types.STDSigner, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(total, tx.data.Amount)
	return total
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}

func (tx *Transaction) RawString() string {
	if tx == nil {
		return ""
	}
	enc, _ := bal.EncodeToBytes(&tx.data)
	return fmt.Sprintf("0x%x", enc)
}

func (tx *Transaction) String() string {
	var from, to string
	if tx.data.V != nil {
		if f, err := tx.From(); err != nil { // derive but don't cache
			from = "[invalid sender: invalid sig]"
		} else {
			from = fmt.Sprintf("%x", f[:])
		}
	} else {
		from = "[invalid sender: nil V field]"
	}

	if tx.data.Recipient == nil {
		to = "[contract creation]"
	} else {
		to = fmt.Sprintf("%x", tx.data.Recipient[:])
	}

	enc, _ := bal.EncodeToBytes(&tx.data)
	return fmt.Sprintf(`
	TX(0x%x)
	Contract: %v
	From:     0x%s
	To:       0x%s
	Nonce:    %v
	GasPrice: %#x
	GasLimit  %#x
	Value:    %#x
	Data:     0x%x
	V:        %#x
	R:        %#x
	S:        %#x
	Hex:      0x%x
`,
		tx.Hash(),
		tx.data.Recipient == nil,
		from,
		to,
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.GasLimit,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.V,
		tx.data.R,
		tx.data.S,
		enc,
	)
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func intrinsicGas(data []byte, contractCreation, homestead bool) (gas uint64, err error) {
	// Set the starting gas for the raw transaction
	if contractCreation && homestead {
		gas = TxGasContractCreation
	} else {
		gas = TxGas
	}
	// Bump the required gas by the amount of transactional data
	if len(data) <= 0 {
		return
	}
	// Zero and non-zero bytes are priced differently
	var nz uint64
	for _, byt := range data {
		if byt != 0 {
			nz++
		}
	}
	// Make sure we don't exceed uint64 for all data combinations
	if (math.MaxUint64-gas)/TxDataNonZeroGas < nz {
		return 0, ErrOutOfGas
	}
	gas += nz * TxDataNonZeroGas

	z := uint64(len(data)) - nz
	if (math.MaxUint64-gas)/TxDataZeroGas < z {
		return 0, ErrOutOfGas
	}
	gas += z * TxDataZeroGas
	return
}

/*
// Transactions is a Transaction slice type for basic sorting.
type Transactions []RegularTx

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Getbal implements balable and returns the i'th element of s in bal.
func (s Transactions) Getbal(i int) []byte {
	enc, _ := bal.EncodeToBytes(s[i])
	return enc
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].Nonce() < s[j].Nonce() }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
*/

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to         *common.Address
	from       common.Address
	tokenAddr  common.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool

	//for ContractCreateTransaction
	contractZone int
	contractAddr common.Address

	txType string
}

func NewMessage(from common.Address, to *common.Address, tokenAddr common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:       from,
		to:         to,
		tokenAddr:  tokenAddr,
		nonce:      nonce,
		amount:     amount,
		gasLimit:   gasLimit,
		gasPrice:   gasPrice,
		data:       data,
		checkNonce: checkNonce,
	}
}

func (m Message) MsgFrom() common.Address { return m.from }
func (m Message) To() *common.Address     { return m.to }
func (m Message) GasPrice() *big.Int      { return m.gasPrice }

// GasCost returns gasprice * gaslimit.
func (m Message) GasCost() *big.Int {
	return new(big.Int).Mul(m.gasPrice, new(big.Int).SetUint64(m.gasLimit))
}
func (m Message) Value() *big.Int              { return m.amount }
func (m Message) Gas() uint64                  { return m.gasLimit }
func (m Message) Nonce() uint64                { return m.nonce }
func (m Message) Data() []byte                 { return m.data }
func (m Message) CheckNonce() bool             { return m.checkNonce }
func (m Message) TokenAddress() common.Address { return m.tokenAddr }
func (m Message) AsMessage() (Message, error)  { return m, nil }
func (m Message) ContractZone() int            { return m.contractZone }
func (m Message) ContractAddr() common.Address { return m.contractAddr }
func (m Message) TxType() string               { return m.txType }

func (m *Message) SetTxType(txType string) { m.txType = txType }
func (m *Message) SetCheckNonce(yn bool)   { m.checkNonce = yn }
