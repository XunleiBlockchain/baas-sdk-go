package sdk

// SDK interface defination
type SDK interface {
	NewAccount(params interface{}) (interface{}, *Error)
	Accounts(params interface{}) (interface{}, *Error)
	GetBalance(params interface{}) (interface{}, *Error)
	GetTransactionCount(params interface{}) (interface{}, *Error)
	GetTransactionByHash(params interface{}) (interface{}, *Error)
	GetTransactionReceipt(params interface{}) (interface{}, *Error)
	SendTransaction(params interface{}) (interface{}, *Error)
	SendContractTransaction(params interface{}) (interface{}, *Error)
	Call(params interface{}) (interface{}, *Error)
}

var _ SDK = gSDK
