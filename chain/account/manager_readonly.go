package account

import (
	"errors"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/LemoFoundationLtd/lemochain-core/store/protocol"
)

var (
	ErrSaveReadOnly = errors.New("can not save a read only account")
)

// ReadOnlyAccount is used to block any save action on Account
type ReadOnlyAccount struct {
	Account
}

func NewReadOnlyAccount(db protocol.ChainDB, address common.Address, data *types.AccountData) *ReadOnlyAccount {
	return &ReadOnlyAccount{Account: *NewAccount(db, address, data)}
}

func (a *ReadOnlyAccount) Finalise() error {
	return ErrSaveReadOnly
}

func (a *ReadOnlyAccount) Save() error {
	return ErrSaveReadOnly
}

// ReadOnlyManager is used to access the newest readonly account data
type ReadOnlyManager struct {
	stableOnly   bool // 是否只读稳定account
	db           protocol.ChainDB
	acctDb       *store.AccountTrieDB
	accountCache map[common.Address]*ReadOnlyAccount
}

// NewManager creates a new Manager. It is used to maintain account changes based on the block environment which specified by blockHash
func NewReadOnlyManager(db protocol.ChainDB, stableOnly bool) *ReadOnlyManager {
	if db == nil {
		panic("account.NewManager is called without a database")
	}
	manager := &ReadOnlyManager{
		stableOnly:   stableOnly,
		db:           db,
		accountCache: make(map[common.Address]*ReadOnlyAccount),
	}

	return manager
}

// Reset clears out all data and switch state to the new block environment. It is not necessary to reset if only use stable accounts data
func (am *ReadOnlyManager) Reset(blockHash common.Hash) {
	exist, err := am.db.IsExistByHash(blockHash)
	if err != nil || !exist {
		log.Errorf("Reset ReadOnlyManager to block[%#x] fail: %s", blockHash, err)
		return
	}

	am.acctDb, _ = am.db.GetActDatabase(blockHash)
	am.accountCache = make(map[common.Address]*ReadOnlyAccount)
}

// GetAccount
func (am *ReadOnlyManager) GetAccount(address common.Address) types.AccountAccessor {
	// 从缓存中读取account
	if cached, ok := am.accountCache[address]; ok {
		return cached
	}

	var data *types.AccountData
	var err error
	if am.stableOnly || am.acctDb == nil {
		data, err = am.db.GetAccount(address)
	} else {
		data, err = am.acctDb.Get(address)
	}

	if err != nil && err != store.ErrNotExist {
		panic(err)
	}
	account := NewReadOnlyAccount(am.db, address, data)
	// cache it
	am.accountCache[address] = account
	return account
}

func (am *ReadOnlyManager) RevertToSnapshot(int) {
}

func (am *ReadOnlyManager) Snapshot() int {
	return 0
}

func (am *ReadOnlyManager) AddEvent(*types.Event) {
}
