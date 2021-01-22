package sdk

import (
	"fmt"
)

type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

func (e *Error) Join(err error) *Error {
	if err == nil {
		return e
	}
	ne := *e
	ne.Msg = fmt.Sprintf("%s (%s)", e.Msg, err.Error())
	return &ne
}

func (e *Error) Error() string {
	return fmt.Sprintf("erro code: %d, msg: %v", e.Code, e.Msg)
}

var (
	ErrSuccess = &Error{
		Code: 0,
		Msg:  "success",
	}

	ErrMethod = &Error{
		Code: -1000,
		Msg:  "invalid method",
	}

	ErrParams = &Error{
		Code: -1001,
		Msg:  "params err",
	}

	ErrAccountFind = &Error{
		Code: -1002,
		Msg:  "find account err",
	}

	ErrRpcGetBalance = &Error{
		Code: -1003,
		Msg:  "rpc getBalance err",
	}

	ErrNewAccount = &Error{
		Code: -1004,
		Msg:  "ks.NewAccount err",
	}

	ErrRpcGetNonce = &Error{
		Code: -1005,
		Msg:  "rpc getNonce err",
	}

	ErrSDKSignTx = &Error{
		Code: -1006,
		Msg:  "SDK SignTx err",
	}

	ErrBalEncodeToBytes = &Error{
		Code: -1007,
		Msg:  "bal EncodeToBytes err",
	}

	ErrRpcSendTransaction = &Error{
		Code: -1008,
		Msg:  "rpc sendTransaction err",
	}

	ErrSDKSignTxWithPassphrase = &Error{
		Code: -1009,
		Msg:  "SDK SignTxWithPassphrase err",
	}

	ErrContractExtension = &Error{
		Code: -1012,
		Msg:  "ContractExtension err",
	}

	ErrRpcSendContractTransaction = &Error{
		Code: -1013,
		Msg:  "rpc sendContractTransaction err",
	}

	ErrJsonUnmarshal = &Error{
		Code: -1018,
		Msg:  "json unmarshal err",
	}
	ErrRpcGetTransactionByHash = &Error{
		Code: -1019,
		Msg:  "rpc getTransactionByHash err",
	}
	ErrRpcGetTransactionReceipt = &Error{
		Code: -1020,
		Msg:  "rpc getTransactionReceipt err",
	}

	ErrSendTxArgs = &Error{
		Code: -1021,
		Msg:  "SendTxArgs err",
	}
	ErrRpcGetGasPrice = &Error{
		Code: -1022,
		Msg:  "rpc getGasPrice err",
	}
	ErrCall = &Error{
		Code: -1023,
		Msg:  "rpc call err",
	}
	ErrRpcEstimateGas = &Error{
		Code: -1024,
		Msg:  "rpc EstimateGas err",
	}

	ErrRpcBlockNumber = &Error{
		Code: -1025,
		Msg:  "rpc BlockNumber err",
	}

	ErrRpcgetBlockByHash = &Error{
		Code: -1026,
		Msg:  "rpc getBlockByHash err",
	}

	ErrRpcgetBlockByNumber = &Error{
		Code: -1027,
		Msg:  "rpc getBlockByNumber err",
	}

	ErrSignTxArgs = &Error{
		Code: -1028,
		Msg:  "SignTx args err",
	}

	ErrSendRawTransaction = &Error{
		Code: -1029,
		Msg:  "SendRawTransaction error",
	}
)
