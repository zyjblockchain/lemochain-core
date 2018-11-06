package protocol

import (
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/store"
)

type ChainDB interface {
	// 设置区块
	SetBlock(hash common.Hash, block *types.Block) error

	// 获取区块 优先根据hash与height同时获取，若hash为空则根据Height获取 获取不到返回：nil,原因
	GetBlock(hash common.Hash, height uint32) (*types.Block, error)
	GetBlockByHeight(height uint32) (*types.Block, error)
	GetBlockByHash(hash common.Hash) (*types.Block, error)
	IsExistByHash(hash common.Hash) (bool, error)

	// 设置区块的确认信息 每次收到一个
	SetConfirmInfo(hash common.Hash, signData types.SignData) error
	AppendConfirmInfo(hash common.Hash, signData types.SignData) error
	SetConfirmPackage(hash common.Hash, pack []types.SignData) error
	AppendConfirmPackage(hash common.Hash, pack []types.SignData) error

	// 获取区块的确认包 获取不到返回：nil,原因
	GetConfirmPackage(hash common.Hash) ([]types.SignData, error)

	// 区块得到共识
	SetStableBlock(hash common.Hash) error

	// GetAccount loads account from cache or db
	GetAccount(blockHash common.Hash, address common.Address) (*types.AccountData, error)
	// SetAccounts saves dirty accounts generated by a block
	SetAccounts(blockHash common.Hash, accounts []*types.AccountData) error
	GetCanonicalAccount(address common.Address) (*types.AccountData, error)
	DelAccount(address common.Address) error

	// GetTrieDatabase returns the db required by storage trie.
	GetTrieDatabase() *store.TrieDatabase
	// GetContractCode loads contract's code from db.
	GetContractCode(codeHash common.Hash) (types.Code, error)
	// SetContractCode saves contract's code
	SetContractCode(codeHash common.Hash, code types.Code) error

	// LoadLatestBlock 程序启动时加载本地最新块
	LoadLatestBlock() (*types.Block, error)
	// Close 关闭数据库
	Close() error
}
