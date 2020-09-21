package sdk

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	gGasPrice    *big.Int
	maxFee       *big.Int
	minFee       *big.Int
	feeRate      int
	defaultRetry = 2
	xhost        = "rpc-baas-blockchain.xunlei.com"
)

func initClient() {
	// 1. init client xhost
	if len(Conf.XHost) > 0 {
		xhost = Conf.XHost
	}
	// 2. init sdk chainID by client query
	if Conf.AuthInfo.ChainID == "" || Conf.AuthInfo.ID == "" || Conf.AuthInfo.Key == "" {
		return
	}
	chainID, err := getChainID()
	if err != nil {
		panic(err)
	}
	Conf.ChainID = chainID
	// 3. init loop
	go getLoop()
}

type rpcReply struct {
	Id      uint64      `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Err     Error       `json:"error"`
}

type feeParms struct {
	Min  string `json:"min"`
	Max  string `json:"max"`
	Rate int    `json:"rate"`
}

type getFeeReply struct {
	Id      uint64   `json:"id"`
	Jsonrpc string   `json:"jsonrpc"`
	Result  feeParms `json:"result"`
	Err     Error    `json:"error"`
}

func getNonce(addr string) (nonce uint64, xerr *Error) {
	params := []string{addr, "pending"}
	reply, err := rpcCall(Conf.Namespace+"_getTransactionCount", params)
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

func getBalance(addr string) (balance interface{}, xerr *Error) {
	params := []string{addr, "latest"}
	reply, err := rpcCall(Conf.Namespace+"_getBalance", params)
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

func getGasPrice() *Error {
	params := []string{}
	reply, err := rpcCall(Conf.Namespace+"_gasPrice", params)
	if err != nil {
		sdklog.Error("gasPrice error.", "err", err)
		return ErrRpcGetGasPrice.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("gasPrice.", "params", params, "reply", string(reply))
	if res.Result != nil {
		if ret, ok := res.Result.(float64); ok {
			gGasPrice = new(big.Int).SetInt64(int64(ret))
		} else {
			var suc bool
			gGasPrice, suc = new(big.Int).SetString(res.Result.(string), 0)
			if !suc {
				return ErrRpcGetGasPrice.Join(fmt.Errorf("getGasPrice SetString failed."))
			}
		}
	} else {
		sdklog.Error("res.Result nil")
	}
	return nil
}

func getFee() error {
	params := []string{}
	reply, err := rpcCall(Conf.Namespace+"_getFee", params)
	if err != nil {
		sdklog.Error("getFee error.", "err", err)
		return err
	}
	var res getFeeReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getFee.", "params", params, "reply", string(reply))
	rlt := res.Result
	minFee, _ = new(big.Int).SetString(rlt.Min, 0)
	minFee.Div(minFee, big.NewInt(1e11))
	maxFee, _ = new(big.Int).SetString(rlt.Max, 0)
	maxFee.Div(maxFee, big.NewInt(1e11))
	feeRate = rlt.Rate
	return nil
}

func getTransactionByHash(from, hash string) (receipt interface{}, xerr *Error) {
	params := []string{hash}
	reply, err := rpcCallWithFrom(Conf.Namespace+"_getTransactionByHash", params, from)
	if err != nil {
		sdklog.Error("getTransactionByHash error.", "err", err)
		return nil, ErrRpcGetTransactionByHash.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getTransactionByHash.", "params", params, "result", res)
	return res.Result, &res.Err
}

func getTransactionReceipt(from, hash string) (receipt interface{}, xerr *Error) {
	params := []string{hash}
	reply, err := rpcCallWithFrom(Conf.Namespace+"_getTransactionReceipt", params, from)
	if err != nil {
		sdklog.Error("getTransactionReceipt error.", "err", err)
		return nil, ErrRpcGetTransactionReceipt.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("getTransactionReceipt.", "params", params, "result", res)
	return res.Result, &res.Err
}

func sendTransaction(raw string) (interface{}, *Error) {
	params := []string{raw}
	reply, err := rpcCall(Conf.Namespace+"_sendRawTransaction", params)
	if err != nil {
		sdklog.Error("sendRawTransaction error.", "err", err)
		return "", ErrRpcSendTransaction.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("sendRawTransaction", "raw", raw, "reply", string(reply))
	return res.Result, &res.Err
}

func sendContractTransaction(raw string, ext interface{}) (interface{}, *Error) {
	params := []string{raw}
	reply, err := rpcCallWithExtension(Conf.Namespace+"_sendRawTransaction", params, ext)
	if err != nil {
		sdklog.Error("sendContractTransaction error.", "err", err)
		return "", ErrRpcSendContractTransaction.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	sdklog.Info("sendContractTransaction.", "raw", raw, "ext", ext, "reply", string(reply))
	return res.Result, &res.Err
}

// ------------------------------- inner call -------------------------------
func doRpcCallWithRetry(retry int, api string, from string, data []byte) (body []byte, err error) {
	for cnt := 0; cnt < retry; cnt++ {
		host := getDNSHost()
		url := fmt.Sprintf("%s/%s", host, api)
		if len(from) != 0 {
			url += fmt.Sprintf("?from=%s", from)
		}
		body, err = httpPostWithLongConn(url, xhost, "application/json", data)
		if err != nil {
			sdklog.Error("rpc call", "err", err)
			updateDNSHost()
			continue
		}
		break
	}
	return
}

func rpcCall(method string, params []string) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, Conf.AuthInfo); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCall", "data(params)", data)
	return doRpcCallWithRetry(defaultRetry, strSlice[1], "", data)
}

func rpcCallWithExtension(method string, params []string, ext interface{}) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["extension"] = ext
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, Conf.AuthInfo); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithExtension", "data(params)", data)
	return doRpcCallWithRetry(defaultRetry, strSlice[1], "", data)
}

func rpcCallWithFrom(method string, params []string, from string) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(params, Conf.AuthInfo); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithFrom", "data(params)", data)
	return doRpcCallWithRetry(defaultRetry, strSlice[1], from, data)
}

// ------------------------------- auth -------------------------------
type rpcAuth struct {
	ChainID string `json:"chainid"`
	ID      string `json:"sdkid"`
	Rand    string `json:"rand"`
	Sign    string `json:"sign"`
}

func genRpcAuth(params []string, authInfo AuthInfo) *rpcAuth {
	if authInfo.ChainID == "" || authInfo.ID == "" || authInfo.Key == "" {
		return nil
	}
	rand := getRandString(16)
	str := rand
	for _, v := range params {
		str = str + strAND(v)
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

func call(from, to, payload string) (interface{}, *Error) {
	params := []interface{}{
		map[string]string{
			"from": from,
			"to":   to,
			"data": payload,
		},
		"latest",
	}
	authParams := []string{from, to, payload}
	reply, err := rpcCallWithAuth(Conf.Namespace+"_call", params, authParams)
	if err != nil {
		return "", ErrCall.Join(err)
	}
	var res rpcReply
	json.Unmarshal(reply, &res)
	return res.Result, &res.Err
}

func rpcCallWithAuth(method string, params interface{}, authParams []string) (body []byte, err error) {
	rpcParams := make(map[string]interface{})
	rpcParams["jsonrpc"] = "2.0"
	rpcParams["method"] = method
	rpcParams["params"] = params
	rpcParams["id"] = 1
	if auth := genRpcAuth(authParams, Conf.AuthInfo); auth != nil {
		rpcParams["auth"] = auth
	}
	data, err := json.Marshal(rpcParams)
	if err != nil {
		return nil, err
	}
	strSlice := strings.Split(method, "_")
	sdklog.Info("rpcCallWithAuth", "data(params)", data)
	return doRpcCallWithRetry(defaultRetry, strSlice[1], "", data)
}

// ------------------------------ inner ------------------------------
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second*15)
}

func calcFee(Amount *big.Int) *big.Int {
	sdklog.Error("calcFee start", "max", maxFee, "min", minFee, "rate", feeRate)
	var fee big.Int
	var amount big.Int
	amount.Set(Amount)
	mod := new(big.Int)
	price := big.NewInt(1e18)
	z, m := amount.DivMod(&amount, price, mod)
	if m.Cmp(big.NewInt(0)) != 0 {
		fee.Set(z.Add(z, big.NewInt(1)).Mul(z, big.NewInt(int64(feeRate))).Mul(z, big.NewInt(1e2)))
	} else {
		fee.Set(z.Mul(z, big.NewInt(int64(feeRate))).Mul(z, big.NewInt(1e2)))
	}
	sdklog.Error("calcFee", "get-fee1", fee)
	if fee.Cmp(minFee) < 0 {
		fee.Set(minFee)
		sdklog.Error("calcFee get less then", "min", minFee, "set-fee", fee)
	}
	if maxFee.Cmp(big.NewInt(0)) != 0 && fee.Cmp(maxFee) > 0 {
		fee.Set(maxFee)
		sdklog.Error("calcFee get more then", "max", maxFee, "set-fee", fee)
	}
	sdklog.Error("calcFee end", "fee", fee)
	return &fee
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

func getChainID() (int64, error) {
	params := []string{}
	reply, err := rpcCall(Conf.Namespace+"_getBaasSdkConf", params)
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

// ------------------------------ loop ------------------------------

func getLoop() {
	interval := 30 * time.Second
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		// timer may be not active, and fired
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(interval)
		select {
		case <-timer.C:
			if Conf.GetFee {
				getFee()
			}
			if Conf.GetGasPrice {
				getGasPrice()
			}
		}
	}
}
