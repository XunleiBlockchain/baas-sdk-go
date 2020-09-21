package types

import (
	"errors"
	"math/big"
)

// SignParam is const param which used to check transaction's sign is correct or not
var (
	SignParam       = big.NewInt(30261)
	SignParamMul    = big.NewInt(30261 * 2)
	GlobalSTDSigner = MakeSTDSigner(SignParam)
)

func ResetSignParam(param *big.Int) {
	SignParam = new(big.Int).Set(param)
	SignParamMul = new(big.Int).Mul(param, big.NewInt(2))
	GlobalSTDSigner = MakeSTDSigner(SignParam)
}

const (
	TxNormal = "tx"

	TxGas                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxGasContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	TxDataNonZeroGas      uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
	ParGasPrice           int64  = 0
	ParGasLimit           uint64 = 1e5
)

var (
	ErrParams             = errors.New("invalid params")
	ErrTxEmpty            = errors.New("tx is empty")
	ErrInvalidSender      = errors.New("invalid sender")
	ErrNonceTooHigh       = errors.New("nonce too high")
	ErrNonceTooLow        = errors.New("nonce too low")
	ErrInsufficientFunds  = errors.New("insufficient funds for gas * price + value")
	ErrIntrinsicGas       = errors.New("intrinsic gas too low")
	ErrGasLimit           = errors.New("exceeds block gas limit")
	ErrOutOfGas           = errors.New("out of gas")
	ErrNegativeValue      = errors.New("negative value")
	ErrOversizedData      = errors.New("oversized data")
	ErrGasLimitOrGasPrice = errors.New("illegal gasLimit or gasPrice")
	ErrFromAddrIsRemote   = errors.New("peer recv from addr is remote")
	ErrInvalidReceiver    = errors.New("invalid receiver")
	ErrInvalidSig         = errors.New("invalid transaction v, r, s values")
)
