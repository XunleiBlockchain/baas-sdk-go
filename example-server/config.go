package main

import (
	"time"
)

type serverConfig struct {
	// for server:
	HTTPAddr         string        `goconf:"base:http.addr"`
	HTTPReadTimeout  time.Duration `goconf:"base:http.read.timeout:time"`
	HTTPWriteTimeout time.Duration `goconf:"base:http.write.timeout:time"`
	// for sdk:
	Keystore               string `goconf:"base:keystore"`
	DNSCacheUpdateInterval int    `goconf:"base:dnscache.updateinterval"`
	RPCProtocal            string `goconf:"base:rpc.protocal"`
	XHost                  string `goconf:"base:xhost"`
	ChainID                int64  `goconf:"base:chain_id"`
	GetFee                 bool   `goconf:"base:getfee"`
	GetGasPrice            bool   `goconf:"base:getgasprice"`
	Namespace              string `goconf:"base:namespace"`
}

func newServerConfig() *serverConfig {
	return &serverConfig{
		HTTPAddr: "8080",
		Keystore: "./keystore",
	}
}
