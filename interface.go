package sdk

// SDK interface defination
type SDK interface {
	NewAccount(params interface{}) (interface{}, *Error)
	Accounts(params interface{}) (interface{}, *Error)
	GetBalance(params interface{}) (interface{}, *Error)
	GetTransactionCount(params interface{}) (interface{}, *Error)
	BlockNumber() (interface{}, *Error)
	GetTransactionByHash(params interface{}) (interface{}, *Error)
	GetTransactionReceipt(params interface{}) (interface{}, *Error)
	GetBlockByNumber(params interface{}) (interface{}, *Error)
	GetBlockByHash(params interface{}) (interface{}, *Error)
	SendTransaction(params interface{}) (interface{}, *Error)
	SendContractTransaction(params interface{}) (interface{}, *Error)
	Call(params interface{}) (interface{}, *Error)
	SignTx(params interface{}) (interface{}, *Error)
	SendRawTransaction(params interface{}) (interface{}, *Error)
}

var _ SDK = &SDKImpl{}
