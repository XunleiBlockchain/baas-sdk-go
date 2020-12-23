package sdk

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var (
	defaultRetry    = 2
	defaultProtocal = "https"
	defaultXHost    = "rpc-baas-blockchain.xunlei.com"
	defaultNS       = "tcapi"
)

type client struct {
	retry     int
	protocal  string
	xHost     string
	nameSpace string

	auth *AuthInfo
}

func defaultClient() *client {
	return &client{
		retry:     defaultRetry,
		protocal:  defaultProtocal,
		xHost:     defaultXHost,
		nameSpace: defaultNS,
	}
}

func newClient(cfg *Config) (*client, error) {
	if cfg.AuthInfo.ChainID == "" || cfg.AuthInfo.ID == "" || cfg.AuthInfo.Key == "" {
		return nil, fmt.Errorf("newClient: AuthInfo empty")
	}
	cli := defaultClient()
	if cfg.Retry > 0 {
		cli.retry = cfg.Retry
	}
	if len(cfg.RPCProtocal) > 0 {
		cli.protocal = cfg.RPCProtocal
	}
	if len(cfg.XHost) > 0 {
		cli.xHost = cfg.XHost
	}
	if len(cfg.Namespace) > 0 {
		cli.nameSpace = cfg.Namespace
	}
	cli.auth = &cfg.AuthInfo
	return cli, nil
}

// ------------------------------- blockchain api -------------------------------
type rpcReply struct {
	ID      uint64      `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Err     Error       `json:"error"`
}

func (c *client) getNonce(addr string) (nonce uint64, xerr *Error) {
	params := []interface{}{addr, "pending"}
	reply, err := c.rpcCall(c.nameSpace+"_getTransactionCount", params)
	if err != nil {
		sdklog.Error("getTransactionCount error.", "err", err)
		return 0, ErrRpcGetNonce.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getTransactionCount.", "params", params, "reply", string(reply))
	if res.Result != nil {
		nonce, err = strconv.ParseUint(res.Result.(string), 0, 64)
		if err != nil {
			return 0, ErrRpcGetNonce.Join(err)
		}
	}
	xerr = &res.Err
	return
}

func (c *client) getBlockNumber() (nonce uint64, xerr *Error) {
	params := []interface{}{}
	reply, err := c.rpcCall(c.nameSpace+"_blockNumber", params)
	if err != nil {
		sdklog.Error("get blockNumber error.", "err", err)
		return 0, ErrRpcBlockNumber.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("get blockNumber.", "params", params, "reply", string(reply))
	if res.Result != nil {
		nonce, err = strconv.ParseUint(res.Result.(string), 0, 64)
		if err != nil {
			return 0, ErrRpcBlockNumber.Join(err)
		}
	}
	xerr = &res.Err
	return
}

func (c *client) getBalance(addr string) (balance interface{}, xerr *Error) {
	params := []interface{}{addr, "latest"}
	reply, err := c.rpcCall(c.nameSpace+"_getBalance", params)
	if err != nil {
		sdklog.Error("getBalance error.", "err", err)
		return nil, ErrRpcGetBalance.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getBalance.", "params", params, "reply", string(reply))
	if res.Result != nil {
		var suc bool
		balance, suc = new(big.Int).SetString(res.Result.(string), 0)
		if !suc {
			return nil, ErrRpcGetBalance.Join(fmt.Errorf("getBalance SetString failed."))
		}
	} else {
		balance = big.NewInt(0)
	}
	xerr = &res.Err
	return
}

func (c *client) getGasPrice() (gasPrice *big.Int, xerr *Error) {
	params := []interface{}{}
	reply, err := c.rpcCall(c.nameSpace+"_gasPrice", params)
	if err != nil {
		sdklog.Error("gasPrice error.", "err", err)
		return nil, ErrRpcGetGasPrice.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("gasPrice.", "params", params, "reply", string(reply))
	if res.Result != nil {
		if ret, ok := res.Result.(float64); ok {
			gasPrice = new(big.Int).SetInt64(int64(ret))
		} else {
			var suc bool
			gasPrice, suc = new(big.Int).SetString(res.Result.(string), 0)
			if !suc {
				return nil, ErrRpcGetGasPrice.Join(fmt.Errorf("getGasPrice SetString failed."))
			}
		}
		return gasPrice, nil
	}
	sdklog.Error("res.Result nil")
	xerr = &res.Err
	return
}

func (c *client) estimateGas(from, to, data string, value big.Int) (gas *big.Int, xerr *Error) {
	params := []interface{}{
		map[string]interface{}{
			"from":  from,
			"to":    to,
			"data":  data,
			"value": "0x" + value.Text(16),
		},
	}
	authParams := []interface{}{from, to, data}
	reply, err := c.rpcCallWithAuth(c.nameSpace+"_estimateGas", params, authParams)
	if err != nil {
		sdklog.Error("estimateGas error.", "err", err)
		return new(big.Int), ErrRpcEstimateGas.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("estimateGas.", "params", params, "reply", string(reply))
	if res.Result != nil {
		gasRes, ok := new(big.Int).SetString(res.Result.(string), 0)
		if !ok {
			return new(big.Int), ErrRpcEstimateGas.Join(err)
		}
		gas = gasRes
	}
	xerr = &res.Err
	return
}

func (c *client) getTransactionByHash(from, hash string) (receipt interface{}, xerr *Error) {
	params := []interface{}{hash}
	reply, err := c.rpcCallWithFrom(c.nameSpace+"_getTransactionByHash", params, from)
	if err != nil {
		sdklog.Error("getTransactionByHash error.", "err", err)
		return nil, ErrRpcGetTransactionByHash.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getTransactionByHash.", "params", params, "result", res)
	return res.Result, &res.Err
}

func (c *client) getTransactionReceipt(hash string) (receipt interface{}, xerr *Error) {
	params := []interface{}{hash}
	reply, err := c.rpcCall(c.nameSpace+"_getTransactionReceipt", params)
	if err != nil {
		sdklog.Error("getTransactionReceipt error.", "err", err)
		return nil, ErrRpcGetTransactionReceipt.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getTransactionReceipt.", "params", params, "result", res)
	return res.Result, &res.Err
}

func (c *client) getBlockByHash(hash string, fullTxReturn bool) (receipt interface{}, xerr *Error) {
	params := []interface{}{hash, strconv.FormatBool(fullTxReturn)}
	reply, err := c.rpcCall(c.nameSpace+"_getBlockByHash", params)
	if err != nil {
		sdklog.Error("getBlockByHash error.", "err", err)
		return nil, ErrRpcgetBlockByHash.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getBlockByHash.", "params", params, "result", res)
	return res.Result, &res.Err
}

func (c *client) getBlockByNumber(number string, fullTxReturn bool) (receipt interface{}, xerr *Error) {
	params := []interface{}{number, fullTxReturn}
	reply, err := c.rpcCall(c.nameSpace+"_getBlockByNumber", params)
	if err != nil {
		sdklog.Error("getBlockByNumber error.", "err", err)
		return nil, ErrRpcgetBlockByNumber.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getBlockByNumber.", "params", params, "result", res)
	return res.Result, &res.Err
}

func (c *client) sendTransaction(raw string) (interface{}, *Error) {
	params := []interface{}{raw}
	reply, err := c.rpcCall(c.nameSpace+"_sendRawTransaction", params)
	if err != nil {
		sdklog.Error("sendRawTransaction error.", "err", err)
		return "", ErrRpcSendTransaction.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("sendRawTransaction", "raw", raw, "reply", string(reply))
	return res.Result, &res.Err
}

func (c *client) sendContractTransaction(raw string, ext interface{}) (interface{}, *Error) {
	params := []interface{}{raw}
	reply, err := c.rpcCallWithExtension(c.nameSpace+"_sendRawTransaction", params, ext)
	if err != nil {
		sdklog.Error("sendContractTransaction error.", "err", err)
		return "", ErrRpcSendContractTransaction.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("sendContractTransaction.", "raw", raw, "ext", ext, "reply", string(reply))
	return res.Result, &res.Err
}

func (c *client) call(from, to, payload string) (interface{}, *Error) {
	params := []interface{}{
		map[string]string{
			"from": from,
			"to":   to,
			"data": payload,
		},
		"latest",
	}
	authParams := []interface{}{from, to, payload}
	reply, err := c.rpcCallWithAuth(c.nameSpace+"_call", params, authParams)
	if err != nil {
		return "", ErrCall.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	return res.Result, &res.Err
}

// ------------------------------- inner call -------------------------------
func (c *client) doRPCCallWithRetry(api string, from string, data []byte) (body []byte, err error) {
	for cnt := 0; cnt < c.retry; cnt++ {
		url := fmt.Sprintf("%s://%s/%s", c.protocal, c.xHost, api)
		if len(from) != 0 {
			url += fmt.Sprintf("?from=%s", from)
		}
		body, err = httpPostWithLongConn(url, c.xHost, "application/json", data)
		if err != nil {
			sdklog.Error("rpc call", "err", err)
			continue
		}
		break
	}
	return
}

func (c *client) rpcCall(method string, params []interface{}) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, *c.auth); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCall", "data(params)", data)
	return c.doRPCCallWithRetry(strSlice[1], "", data)
}

func (c *client) rpcCallWithExtension(method string, params []interface{}, ext interface{}) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["extension"] = ext
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, *c.auth); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithExtension", "data(params)", data)
	return c.doRPCCallWithRetry(strSlice[1], "", data)
}

func (c *client) rpcCallWithFrom(method string, params []interface{}, from string) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, *c.auth); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithFrom", "data(params)", data)
	return c.doRPCCallWithRetry(strSlice[1], from, data)
}

func (c *client) rpcCallWithAuth(method string, params interface{}, authParams []interface{}) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(authParams, *c.auth); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithAuth", "data(params)", data)
	return c.doRPCCallWithRetry(strSlice[1], "", data)
}

// ------------------------------ getChainID ------------------------------
type ChainIDData struct {
	ChainID int64 `json:"chainid"`
}

type ChainIDReply struct {
	Code uint64      `json:"code"`
	Msg  string      `json:"msg"`
	Data ChainIDData `json:"data"`
}

func (c *client) getChainID() (int64, error) {
	params := []interface{}{}
	reply, err := c.rpcCall(c.nameSpace+"_getBaasSdkConf", params)
	if err != nil {
		return 0, err
	}
	var res ChainIDReply
	json.Unmarshal(reply, &res)
	if res.Code != 0 {
		return 0, fmt.Errorf("errcode: %d errmsg: %s", res.Code, res.Msg)
	}
	return res.Data.ChainID, nil
}

// ------------------------------- auth -------------------------------
type rpcAuth struct {
	ChainID string `json:"chainid"`
	ID      string `json:"sdkid"`
	Rand    string `json:"rand"`
	Sign    string `json:"sign"`
}

func genRpcAuth(params []interface{}, authInfo AuthInfo) *rpcAuth {
	if authInfo.ChainID == "" || authInfo.ID == "" || authInfo.Key == "" {
		return nil
	}
	rand := getRandString(16)
	str := rand
	for _, v := range params {
		switch v.(type) {
		case string:
			str = str + strAND(v.(string))
		case bool:
			str = str + strAND(strconv.FormatBool(v.(bool)))
		}
	}
	str = str + strAND(authInfo.ChainID) + strAND(authInfo.ID) + strAND(authInfo.Key)
	hash := sha256.Sum256([]byte(str))
	md5sum := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%x", hash[:]))))
	sdklog.Info("rpc-auth", "origin", str, "hash", hash, "sign", md5sum)
	return &rpcAuth{
		ChainID: authInfo.ChainID,
		ID:      authInfo.ID,
		Rand:    rand,
		Sign:    md5sum,
	}
}
