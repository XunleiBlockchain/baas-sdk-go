package sdk

import (
	"fmt"
	"runtime/debug"

	"github.com/XunleiBlockchain/tc-libs/accounts"
	"github.com/XunleiBlockchain/tc-libs/accounts/keystore"
	"github.com/XunleiBlockchain/tc-libs/bal"
	"github.com/XunleiBlockchain/tc-libs/common"
)

func (sdk *SDKImpl) NewAccount(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 1 {
		return "", ErrParams
	}
	passwd := args[0].(string)
	ks := sdk.am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	acc, err := ks.NewAccount(passwd)
	if err == nil {
		err = ks.Unlock(acc, passwd)
		debug.FreeOSMemory()
		if err != nil {
			sdklog.Error("new account unlock fail", "account", acc, "passwd", passwd)
		}
		return acc.Address, nil
	}
	sdklog.Error("new account fail", "account", acc, "passwd", passwd)
	return common.Address{}, ErrNewAccount.Join(err)
}

func (sdk *SDKImpl) Accounts(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 0 {
		return "", ErrParams
	}
	addresses := make([]common.Address, 0)
	for _, wallet := range sdk.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses, nil
}

func (sdk *SDKImpl) GetBalance(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 1 {
		return "", ErrParams
	}
	addr := args[0].(string)
	account := accounts.Account{Address: common.HexToAddress(addr)}
	_, err := sdk.am.Find(account)
	if err != nil {
		return 0, ErrAccountFind.Join(err)
	}
	return sdk.c.getBalance(addr)
}

func (sdk *SDKImpl) GetTransactionCount(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 1 {
		return "", ErrParams
	}
	addr := args[0].(string)
	account := accounts.Account{Address: common.HexToAddress(addr)}
	_, err := sdk.am.Find(account)
	if err != nil {
		return 0, ErrAccountFind.Join(err)
	}
	return sdk.c.getNonce(addr)
}

func (sdk *SDKImpl) GetTransactionByHash(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 2 {
		return "", ErrParams
	}
	from := args[0].(string)
	account := accounts.Account{Address: common.HexToAddress(from)}
	_, err := sdk.am.Find(account)
	if err != nil {
		return 0, ErrAccountFind.Join(err)
	}
	hash := args[1].(string)
	return sdk.c.getTransactionByHash(from, hash)
}

func (sdk *SDKImpl) GetTransactionReceipt(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 2 {
		return "", ErrParams
	}
	from := args[0].(string)
	account := accounts.Account{Address: common.HexToAddress(from)}
	_, err := sdk.am.Find(account)
	if err != nil {
		return 0, ErrAccountFind.Join(err)
	}
	hash := args[1].(string)
	return sdk.c.getTransactionReceipt(from, hash)
}

func (sdk *SDKImpl) SendTransaction(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 1 && len(args) != 2 {
		return common.Hash{}, ErrParams
	}
	var sendTxArgs SendTxArgs
	err := sendTxArgs.parseFromArgs(args[0])
	if err != nil {
		return common.Hash{}, ErrSendTxArgs.Join(err)
	}
	if sdk.cfg.GetGasPrice {
		sendTxArgs.GasPrice = sdk.gasPrice
	}
	account := accounts.Account{Address: sendTxArgs.From}
	wallet, err := sdk.am.Find(account)
	if err != nil {
		return common.Hash{}, ErrAccountFind.Join(err)
	}
	if sendTxArgs.Nonce == nil {
		sdk.nonceLock.lockAddr(sendTxArgs.From)
		defer sdk.nonceLock.unlockAddr(sendTxArgs.From)
	}
	if err = sendTxArgs.setDefaults(sdk.c); err != nil {
		return common.Hash{}, ErrSendTxArgs.Join(err)
	}
	tx := sendTxArgs.toTransaction()
	stx, ok := tx.(accounts.SingerTx)
	if !ok {
		return nil, ErrSendTxArgs.Join(fmt.Errorf("tx is not a SignerTx type"))
	}
	// send transaction without password
	if len(args) == 1 {
		signed, err := wallet.SignTx(account, stx, sdk.signParam)
		debug.FreeOSMemory()
		if err != nil {
			return common.Hash{}, ErrSDKSignTx.Join(err)
		}
		txbal, err := bal.EncodeToBytes(signed)
		if err != nil {
			sdklog.Error("SendTransaction bal.EncodeToBytes()", "err", err)
			return common.Hash{}, ErrBalEncodeToBytes.Join(err)
		}
		res, xerr := sdk.c.sendTransaction(common.ToHex(txbal))
		if xerr.Code != 0 {
			res = common.Hash{}
		}
		sdklog.Info("SendTransaction", "res", "xerr", *xerr)
		return res, xerr
	}
	// send transaction with password
	passwd := args[1].(string)
	signed, err := wallet.SignTxWithPassphrase(account, passwd, stx, sdk.signParam)
	debug.FreeOSMemory()
	if err != nil {
		return common.Hash{}, ErrSDKSignTxWithPassphrase.Join(err)
	}
	txbal, err := bal.EncodeToBytes(signed)
	if err != nil {
		sdklog.Error("SendTransaction bal.EncodeToBytes()", "err", err)
		return common.Hash{}, ErrBalEncodeToBytes.Join(err)
	}
	res, xerr := sdk.c.sendTransaction(common.ToHex(txbal))
	if xerr.Code != 0 {
		res = common.Hash{}
	}
	sdklog.Info("SendTransaction", "res", "xerr", *xerr)
	return res, xerr
}

func (sdk *SDKImpl) SendContractTransaction(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) == 0 || len(args) > 3 {
		return common.Hash{}, ErrParams
	}
	var sendTxArgs SendTxArgs
	err := sendTxArgs.parseFromArgs(args[0])
	if err != nil {
		return common.Hash{}, ErrSendTxArgs.Join(err)
	}
	if sdk.cfg.GetGasPrice {
		sendTxArgs.GasPrice = sdk.gasPrice
	}
	account := accounts.Account{Address: sendTxArgs.From}
	wallet, err := sdk.am.Find(account)
	if err != nil {
		return common.Hash{}, ErrAccountFind.Join(err)
	}
	if sendTxArgs.Nonce == nil {
		sdk.nonceLock.lockAddr(sendTxArgs.From)
		defer sdk.nonceLock.unlockAddr(sendTxArgs.From)
	}
	if err = sendTxArgs.setDefaults(sdk.c); err != nil {
		return common.Hash{}, ErrSendTxArgs.Join(err)
	}
	tx := sendTxArgs.toContractTransaction()
	stx, ok := tx.(accounts.SingerTx)
	if !ok {
		return nil, ErrSendTxArgs.Join(fmt.Errorf("tx is not a SignerTx type"))
	}

	var contractArgs ContractExtension
	//send contract transaction without password
	switch len(args) {
	case 1:
		signed, err := wallet.SignTx(account, stx, sdk.signParam)
		debug.FreeOSMemory()
		if err != nil {
			return common.Hash{}, ErrSDKSignTx.Join(err)
		}
		txbal, err := bal.EncodeToBytes(signed)
		if err != nil {
			sdklog.Error("SendContractTransaction bal.EncodeToBytes()", "err", err)
			return common.Hash{}, ErrBalEncodeToBytes.Join(err)
		}
		res, xerr := sdk.c.sendTransaction(common.ToHex(txbal))
		if xerr.Code != 0 {
			res = common.Hash{}
		}
		return res, xerr
	case 2:
		err = contractArgs.parseFromArgs(args[1])
		if err != nil {
			return common.Hash{}, ErrContractExtension.Join(err)
		}
		signed, err := wallet.SignTx(account, stx, sdk.signParam)
		debug.FreeOSMemory()
		if err != nil {
			return common.Hash{}, ErrSDKSignTx.Join(err)
		}
		txbal, err := bal.EncodeToBytes(signed)
		if err != nil {
			sdklog.Error("SendContractTransaction bal.EncodeToBytes()", "err", err)
			return common.Hash{}, ErrBalEncodeToBytes.Join(err)
		}
		res, xerr := sdk.c.sendContractTransaction(common.ToHex(txbal), contractArgs)
		if xerr.Code != 0 {
			res = common.Hash{}
		}
		return res, xerr
	case 3:
		//send contract transaction with password
		passwd := args[1].(string)
		err = contractArgs.parseFromArgs(args[2])
		if err != nil {
			return common.Hash{}, ErrContractExtension.Join(err)
		}
		signed, err := wallet.SignTxWithPassphrase(account, passwd, stx, sdk.signParam)
		debug.FreeOSMemory()
		if err != nil {
			return common.Hash{}, ErrSDKSignTxWithPassphrase.Join(err)
		}
		txbal, err := bal.EncodeToBytes(signed)
		if err != nil {
			sdklog.Error("SendContractTransaction bal.EncodeToBytes()", "err", err)
			return common.Hash{}, ErrBalEncodeToBytes.Join(err)
		}
		res, xerr := sdk.c.sendContractTransaction(common.ToHex(txbal), contractArgs)
		if xerr.Code != 0 {
			res = common.Hash{}
		}
		return res, xerr
	default:
		return common.Hash{}, ErrParams
	}
}

func (sdk *SDKImpl) Call(params interface{}) (interface{}, *Error) {
	defer catchInterfacePanic()
	args := params.([]interface{})
	if len(args) != 1 {
		return nil, ErrParams
	}
	var callArgs CallArgs
	err := callArgs.parseFromArgs(args[0])
	if err != nil {
		return nil, ErrSendTxArgs.Join(err)
	}
	account := accounts.Account{Address: common.HexToAddress(callArgs.From)}
	_, err = sdk.am.Find(account)
	if err != nil {
		return nil, ErrAccountFind.Join(err)
	}
	return sdk.c.call(callArgs.From, callArgs.To, callArgs.Data)
}
