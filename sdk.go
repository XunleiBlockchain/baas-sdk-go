package sdk

import (
	"fmt"
	"math/big"
	"runtime/debug"

	"github.com/XunleiBlockchain/tc-libs/accounts"
	"github.com/XunleiBlockchain/tc-libs/accounts/keystore"
	"github.com/XunleiBlockchain/tc-libs/common"
)

var gSDK *SDKImpl

func GetSDK(log Logger) *SDKImpl {
	setLogger(log)
	if gSDK == nil {
		newSDK()
	}
	return gSDK
}

func ReleaseSDK() {
	gSDK = nil
}

// ---------------------------------------------------------------------

type SDKImpl struct {
	signParam *big.Int
	am        *accounts.Manager
	nonceLock *addrLocker
}

func newSDK() {
	// 1. init
	initDNSCache()
	initClient()
	// 2. make account manager
	am, err := makeAccountManager(Conf.Keystore)
	if err != nil {
		panic(err)
	}
	// 2. unlock accounts
	ks := am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	for addr, passwd := range Conf.UnlockAccounts {
		acc := accounts.Account{Address: common.HexToAddress(addr)}
		_, err = am.Find(acc)
		if err != nil {
			sdklog.Error("new server account not found.", "account", addr)
			continue
		}
		err = ks.Unlock(acc, passwd)
		debug.FreeOSMemory()
		if err != nil {
			sdklog.Error("new server unlock fail.", "account", addr, "passwd", passwd)
			continue
		}
	}
	// 3. get addrLocker
	gSDK = &SDKImpl{
		signParam: big.NewInt(0).SetBytes([]byte(fmt.Sprintf("%d", Conf.ChainID))),
		am:        am,
		nonceLock: &addrLocker{},
	}
}
