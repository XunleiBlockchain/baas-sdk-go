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
	sdkConf, err := initServerConfig()
	if err != nil {
		panic(err)
	}

	// 2. init logger
	logger = log.Sugar()

	// 3. new sdk
	mySDK, err := sdk.NewSDK(sdkConf, logger)
	if err != nil {
		panic(err)
	}

	// 4. start HTTPServer
	logger.Info("sdk-server start.")
	if err := initHTTP(mySDK); err != nil {
		panic(err)
	}

	select {}
}

func initServerConfig() (*sdk.Config, error) {
	// 1. parse serverConfig
	conf = newServerConfig()
	gconf := goconf.New()
	if err := gconf.Parse(confFile); err != nil {
		return nil, err
	}
	if err := gconf.Unmarshal(conf); err != nil {
		return nil, err
	}
	// 2. new sdk.Config
	sdkConf := &sdk.Config{
		Keystore:       conf.Keystore,
		UnlockAccounts: make(map[string]string),
		XHost:          conf.XHost,
		Namespace:      conf.Namespace,
		ChainID:        conf.ChainID,
		GetGasPrice:    conf.GetGasPrice,
	}
	if authInfoFile != "" {
		authInfoJSON, err := ioutil.ReadFile(authInfoFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(authInfoJSON, &sdkConf.AuthInfo)
		if err != nil {
			return nil, err
		}
	}
	if passwdFile != "" {
		passwdsJSON, err := ioutil.ReadFile(passwdFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(passwdsJSON, &sdkConf.UnlockAccounts)
		if err != nil {
			return nil, err
		}
	}
	return sdkConf, nil
}
