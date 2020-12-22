package sdk

import (
	"errors"
	"math/big"
	"strconv"

	"github.com/XunleiBlockchain/baas-sdk-go/types"
	"github.com/XunleiBlockchain/tc-libs/common"
)

type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *big.Int        `json:"gas"`
	GasPrice *big.Int        `json:"gasPrice"`
	Value    *big.Int        `json:"value"`
	Data     []byte          `json:"data"`
	Nonce    *uint64         `json:"nonce"`
}

func (args *SendTxArgs) parseFromArgs(params interface{}) (err error) {
	txArgs := params.(map[string]interface{})
	var succ bool
	if from, ok := txArgs["from"]; ok {
		args.From = common.HexToAddress(from.(string))
	}
	if to, ok := txArgs["to"]; ok {
		if !common.IsHexAddress(to.(string)) {
			return errors.New("invalid to address")
		}
		toAddr := common.HexToAddress(to.(string))
		args.To = &toAddr
	}
	if gas, ok := txArgs["gas"]; ok {
		args.Gas, succ = new(big.Int).SetString(gas.(string), 0)
		if !succ {
			return errors.New("gas err.")
		}
	}
	if gasPrice, ok := txArgs["gasPrice"]; ok {
		args.GasPrice, succ = new(big.Int).SetString(gasPrice.(string), 0)
		if !succ {
			return errors.New("gasPrice err.")
		}
	}
	if value, ok := txArgs["value"]; ok {
		args.Value, succ = new(big.Int).SetString(value.(string), 0)
		if !succ {
			return errors.New("value err.")
		}
	}
	if data, ok := txArgs["data"]; ok {
		args.Data = common.FromHex(data.(string))
	}
	if nonce, ok := txArgs["nonce"]; ok {
		nonceUint64, err := strconv.ParseUint(nonce.(string), 0, 64)
		if err != nil {
			return errors.New("nonce err.")
		}
		args.Nonce = &nonceUint64
	}
	return nil
}

func (args *SendTxArgs) setDefaults(c *client) error {
	if args.GasPrice == nil {
		args.GasPrice = big.NewInt(1e11)
	}
	if args.Value == nil {
		args.Value = new(big.Int)
	}
	if args.Gas == nil {
		gas, xerr := c.estimateGas(args.From.String(), args.To.String(), common.ToHex(args.Data), *args.Value)
		if xerr != nil && xerr.Code != 0 {
			return errors.New(xerr.Msg)
		}
		args.Gas = gas
	}
	if args.Nonce == nil {
		nonce, xerr := c.getNonce(args.From.String())
		if xerr != nil && xerr.Code != 0 {
			return errors.New(xerr.Msg)
		}
		args.Nonce = &nonce
	}
	return nil
}

func (args *SendTxArgs) setData(data []byte) error {
	args.Data = data
	return nil
}

func (args *SendTxArgs) toTransaction() types.Tx {
	// nonce toAddress amout gasLimit gasPrice data
	return types.NewTransaction(uint64(*args.Nonce), *args.To, args.Value, args.Gas.Uint64(), args.GasPrice, args.Data)
}

func (args *SendTxArgs) toContractTransaction() types.Tx {
	return types.NewTransaction(uint64(*args.Nonce), *args.To, args.Value, args.Gas.Uint64(), args.GasPrice, args.Data)
}

type ContractExtension struct {
	Callback  string `json:"callback"`
	PrepayID  string `json:"prepay_id"`
	ServiceID string `json:"service_id"`
	TxType    string `json:"tx_type"`
	Sign      string `json:"sign"`
	Title     string `json:"title"`
	Desc      string `json:"desc"`
}

func (args *ContractExtension) parseFromArgs(params interface{}) (err error) {
	extArgs := params.(map[string]interface{})
	args.Callback = checkStrRes(extArgs["callback"], "")
	args.PrepayID = checkStrRes(extArgs["prepay_id"], "")
	args.ServiceID = checkStrRes(extArgs["service_id"], "")
	args.TxType = checkStrRes(extArgs["tx_type"], "contract")
	args.Sign = checkStrRes(extArgs["sign"], "")
	args.Title = checkStrRes(extArgs["title"], "")
	args.Desc = checkStrRes(extArgs["desc"], "")
	return nil
}

type CallArgs struct {
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
}

func (args *CallArgs) parseFromArgs(params interface{}) (err error) {
	txArgs := params.(map[string]interface{})
	args.From = checkStrRes(txArgs["from"], "")
	args.To = checkStrRes(txArgs["to"], "")
	args.Data = checkStrRes(txArgs["data"], "")
	return nil
}
