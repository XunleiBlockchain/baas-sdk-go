package sdk

import (
	"os"
	"sync"

	"github.com/XunleiBlockchain/tc-libs/accounts"
	"github.com/XunleiBlockchain/tc-libs/accounts/keystore"
	"github.com/XunleiBlockchain/tc-libs/common"
)

func makeAccountManager(keydir string) (*accounts.Manager, error) {
	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP
	if err := os.MkdirAll(keydir, 0700); err != nil {
		return nil, err
	}
	backends := []accounts.Backend{
		keystore.NewKeyStore(keydir, scryptN, scryptP),
	}
	return accounts.NewManager(backends...), nil
}

type addrLocker struct {
	mu    sync.Mutex
	locks map[common.Address]*sync.Mutex
}

func (l *addrLocker) lock(address common.Address) *sync.Mutex {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.locks == nil {
		l.locks = make(map[common.Address]*sync.Mutex)
	}
	if _, ok := l.locks[address]; !ok {
		l.locks[address] = new(sync.Mutex)
	}
	return l.locks[address]
}

func (l *addrLocker) lockAddr(address common.Address) {
	l.lock(address).Lock()
}

func (l *addrLocker) unlockAddr(address common.Address) {
	l.lock(address).Unlock()
}
