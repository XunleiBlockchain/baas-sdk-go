package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	sdk "github.com/XunleiBlockchain/baas-sdk-go"
)

type request struct {
	ID      int64       `json:"id"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// Server serve http
type Server struct {
	mySDK *sdk.SDKImpl
}

func newServer(mySDK *sdk.SDKImpl) *Server {
	return &Server{
		mySDK: mySDK,
	}
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	logger.Info("ServeHTTP", "url", r.URL, "params", string(body))
	w.Header().Set("content-type", "application/json")
	var req request
	resp := make(map[string]interface{})
	err := json.Unmarshal(body, &req)
	if err != nil {
		resp["id"] = req.ID
		resp["jsonrpc"] = req.Jsonrpc
		xerr := sdk.ErrJsonUnmarshal.Join(err)
		resp["errcode"] = xerr.Code
		resp["errmsg"] = xerr.Msg
		//resp["result"] = fmt.Sprintf("err: %v", err)
		respByte, _ := json.Marshal(resp)
		w.Write(respByte)
		return
	}
	var ret interface{}
	var xerr *sdk.Error
	switch req.Method {
	case "accounts":
		ret, xerr = srv.mySDK.Accounts(req.Params)
		break
	case "newAccount":
		ret, xerr = srv.mySDK.NewAccount(req.Params)
		break
	case "getBalance":
		ret, xerr = srv.mySDK.GetBalance(req.Params)
		break
	case "getTransactionCount":
		ret, xerr = srv.mySDK.GetTransactionCount(req.Params)
		break
	case "getTransactionByHash":
		ret, xerr = srv.mySDK.GetTransactionByHash(req.Params)
		break
	case "getTransactionReceipt":
		ret, xerr = srv.mySDK.GetTransactionReceipt(req.Params)
		break
	case "sendTransaction":
		ret, xerr = srv.mySDK.SendTransaction(req.Params)
		break
	case "sendContractTransaction":
		ret, xerr = srv.mySDK.SendContractTransaction(req.Params)
		break
	case "call":
		ret, xerr = srv.mySDK.Call(req.Params)
		break
	default:
		xerr = sdk.ErrMethod
	}
	resp["id"] = req.ID
	resp["jsonrpc"] = req.Jsonrpc
	resp["result"] = ret
	if xerr == nil || xerr.Code == 0 {
		xerr = sdk.ErrSuccess
	}
	resp["errcode"] = xerr.Code
	resp["errmsg"] = xerr.Msg
	//resp["result"] = fmt.Sprintf("err: %v", err)
	respByte, err := json.Marshal(resp)
	w.Write(respByte)
	return
}

func initHTTP(mySDK *sdk.SDKImpl) (err error) {
	server := newServer(mySDK)
	httpServer := &http.Server{Handler: server, ReadTimeout: conf.HTTPReadTimeout, WriteTimeout: conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	listener, err := net.Listen("tcp", conf.HTTPAddr)
	if err != nil {
		return err
	}
	err = httpServer.Serve(listener)
	return
}
