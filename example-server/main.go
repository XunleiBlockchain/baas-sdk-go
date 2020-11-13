package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"github.com/Terry-Mao/goconf"
	"github.com/binacsgo/log"

	sdk "github.com/XunleiBlockchain/baas-sdk-go"
)

var (
	conf         *serverConfig
	logger       log.Logger
	confFile     string
	passwdFile   string
	authInfoFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./sdk-server.conf", " set sdk-server config file path")
	flag.StringVar(&passwdFile, "p", "", " set password json file path")
	flag.StringVar(&authInfoFile, "a", "", " set auth info json file path")
}

func main() {
	flag.Parse()
	// 1. init serverConfig and sdk.Config
	if err := initServerConfig(); err != nil {
		panic(err)
	}

	// 2. init logger
	logger = log.Sugar()

	// 3. get sdk instance
	ins := sdk.GetSDK(logger)

	// 4. start HTTPServer
	logger.Info("sdk-server start.")
	if err := initHTTP(ins); err != nil {
		panic(err)
	}

	select {}
}

func initServerConfig() (err error) {
	// 1. parse serverConfig
	conf = newServerConfig()
	gconf := goconf.New()
	if err = gconf.Parse(confFile); err != nil {
		return err
	}
	if err := gconf.Unmarshal(conf); err != nil {
		return err
	}
	// 2. init sdk.Config
	sdkConf := &sdk.Config{
		Keystore:               conf.Keystore,
		UnlockAccounts:         make(map[string]string),
		DNSCacheUpdateInterval: conf.DNSCacheUpdateInterval,
		RPCProtocal:            conf.RPCProtocal,
		XHost:                  conf.XHost,
		Namespace:              conf.Namespace,
		ChainID:                conf.ChainID,
		GetGasPrice:            conf.GetGasPrice,
	}
	if authInfoFile != "" {
		authInfoJSON, err := ioutil.ReadFile(authInfoFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(authInfoJSON, &sdkConf.AuthInfo)
		if err != nil {
			return err
		}
	}
	if passwdFile != "" {
		passwdsJSON, err := ioutil.ReadFile(passwdFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(passwdsJSON, &sdkConf.UnlockAccounts)
		if err != nil {
			return err
		}
	}
	sdk.Conf = sdkConf
	return nil
}
