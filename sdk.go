package sdk

import (
	"fmt"
	"math/big"
	"runtime/debug"
	"time"

	"github.com/XunleiBlockchain/tc-libs/accounts"
	"github.com/XunleiBlockchain/tc-libs/accounts/keystore"
	"github.com/XunleiBlockchain/tc-libs/common"
)

// ------------------------------- SDK Impl -------------------------------
type SDKImpl struct {
	cfg       *Config
	am        *accounts.Manager
	signParam *big.Int
	gasPrice  *big.Int
	nonceLock *addrLocker
	c         *client
}

// NewSDK return a pointer to SDKImpl
func NewSDK(cfg *Config, log Logger) (*SDKImpl, error) {
	sdklog = log
	// 1. account manager
	am, err := makeAccountManager(cfg.Keystore)
	if err != nil {
		return nil, fmt.Errorf("New: makeAccountManager error: %v", err)
	}
	// 2. keystore
	ks := am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	for addr, passwd := range cfg.UnlockAccounts {
		acc := accounts.Account{Address: common.HexToAddress(addr)}
		_, err = am.Find(acc)
		if err != nil {
			continue
		}
		err = ks.Unlock(acc, passwd)
		debug.FreeOSMemory()
		if err != nil {
			continue
		}
	}
	// 3. get client
	cli, err := newClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("New: newClient error: %v", err)
	}
	// 3. get chain id
	chainID, err := cli.getChainID()
	if err != nil {
		return nil, fmt.Errorf("New: getChainID error: %v", err)
	}
	// 4. get SDKImpl
	sdk := &SDKImpl{
		cfg:       cfg,
		signParam: big.NewInt(0).SetBytes([]byte(fmt.Sprintf("%d", chainID))),
		am:        am,
		nonceLock: &addrLocker{},
		c:         cli,
	}
	go sdk.getLoop()
	return sdk, nil
}

func (sdk *SDKImpl) getLoop() {
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
			if gasPrice, err := sdk.c.getGasPrice(); err == nil {
				sdk.gasPrice = gasPrice
			}
		}
	}
}
