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
	ins *sdk.SDKImpl
}

func newServer(ins *sdk.SDKImpl) *Server {
	return &Server{
		ins: ins,
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
		ret, xerr = srv.ins.Accounts(req.Params)
		break
	case "newAccount":
		ret, xerr = srv.ins.NewAccount(req.Params)
		break
	case "getBalance":
		ret, xerr = srv.ins.GetBalance(req.Params)
		break
	case "getTransactionCount":
		ret, xerr = srv.ins.GetTransactionCount(req.Params)
		break
	case "getTransactionByHash":
		ret, xerr = srv.ins.GetTransactionByHash(req.Params)
		break
	case "getTransactionReceipt":
		ret, xerr = srv.ins.GetTransactionReceipt(req.Params)
		break
	case "sendTransaction":
		ret, xerr = srv.ins.SendTransaction(req.Params)
		break
	case "sendContractTransaction":
		ret, xerr = srv.ins.SendContractTransaction(req.Params)
		break
	case "call":
		ret, xerr = srv.ins.Call(req.Params)
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

func initHTTP(ins *sdk.SDKImpl) (err error) {
	server := newServer(ins)
	httpServer := &http.Server{Handler: server, ReadTimeout: conf.HTTPReadTimeout, WriteTimeout: conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	listener, err := net.Listen("tcp", conf.HTTPAddr)
	if err != nil {
		return err
	}
	err = httpServer.Serve(listener)
	return
}
