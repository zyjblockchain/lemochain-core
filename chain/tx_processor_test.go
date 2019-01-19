package chain

import (
	"encoding/json"
	"github.com/LemoFoundationLtd/lemochain-go/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-go/chain/params"
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/chain/vm"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-go/common/flag"
	"github.com/LemoFoundationLtd/lemochain-go/common/rlp"
	"github.com/LemoFoundationLtd/lemochain-go/store"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestNewTxProcessor(t *testing.T) {
	chain := newChain()
	p := NewTxProcessor(chain)
	assert.NotEqual(t, (*vm.Config)(nil), p.cfg)
	assert.Equal(t, false, p.cfg.Debug)

	flags := flag.CmdFlags{}
	flags.Set(common.Debug, "1")
	chain, _ = NewBlockChain(chainID, NewDpovp(10*1000, chain.db), chain.db, flags)
	p = NewTxProcessor(chain)
}

// test valid block processing
func TestTxProcessor_Process(t *testing.T) {
	store.ClearData()
	p := NewTxProcessor(newChain())

	sender := p.am.GetAccount(testAddr)
	// last not stable block
	block := defaultBlocks[2]
	newHeader, err := p.Process(block)
	assert.NoError(t, err)
	assert.Equal(t, block.Header.Bloom, newHeader.Bloom)
	assert.Equal(t, block.Header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, block.Header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, block.Header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, block.Header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, block.Header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, block.Hash(), newHeader.Hash())
	sender = p.am.GetAccount(testAddr)
	assert.Equal(t, 3, len(sender.GetTxHashList()))
	assert.Equal(t, block.Txs[0].Hash(), sender.GetTxHashList()[2])

	// block not in db
	block = defaultBlocks[3]
	newHeader, err = p.Process(block)
	assert.NoError(t, err)
	assert.Equal(t, block.Header.Bloom, newHeader.Bloom)
	assert.Equal(t, block.Header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, block.Header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, block.Header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, block.Header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, block.Header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, block.Hash(), newHeader.Hash())
	sender = p.am.GetAccount(testAddr)
	assert.Equal(t, 5, len(sender.GetTxHashList()))
	assert.Equal(t, block.Txs[0].Hash(), sender.GetTxHashList()[3])

	// genesis block
	block = defaultBlocks[0]
	newHeader, err = p.Process(block)
	assert.NoError(t, err)
	assert.Equal(t, block.Header.Bloom, newHeader.Bloom)
	assert.Equal(t, block.Header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, block.Header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, block.Header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, block.Header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, block.Header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, block.Hash(), newHeader.Hash())

	// block on fork branch
	block = createNewBlock()
	newHeader, err = p.Process(block)
	assert.NoError(t, err)
	assert.Equal(t, block.Header.Bloom, newHeader.Bloom)
	assert.Equal(t, block.Header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, block.Header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, block.Header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, block.Header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, block.Header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, block.Hash(), newHeader.Hash())
}

// test invalid block processing
func TestTxProcessor_Process2(t *testing.T) {
	store.ClearData()
	p := NewTxProcessor(newChain())

	// tamper with amount
	block := createNewBlock()
	rawTx, _ := rlp.EncodeToBytes(block.Txs[0])
	rawTx[29]++ // amount++
	cpy := new(types.Transaction)
	err := rlp.DecodeBytes(rawTx, cpy)
	assert.NoError(t, err)
	assert.Equal(t, new(big.Int).Add(block.Txs[0].Amount(), big.NewInt(1)), cpy.Amount())
	block.Txs[0] = cpy
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// invalid signature
	block = createNewBlock()
	rawTx, _ = rlp.EncodeToBytes(block.Txs[0])
	rawTx[43] = 0 // invalid S
	cpy = new(types.Transaction)
	err = rlp.DecodeBytes(rawTx, cpy)
	assert.NoError(t, err)
	block.Txs[0] = cpy
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// not enough gas (resign by another address)
	block = createNewBlock()
	private, _ := crypto.GenerateKey()
	origFrom, _ := block.Txs[0].From()
	block.Txs[0] = signTransaction(block.Txs[0], private)
	newFrom, _ := block.Txs[0].From()
	assert.NotEqual(t, origFrom, newFrom)
	block.Header.TxRoot = types.DeriveTxsSha(block.Txs)
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// exceed block gas limit
	block = createNewBlock()
	block.Header.GasLimit = 1
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// used gas reach limit in some tx
	block = createNewBlock()
	block.Txs[0] = makeTransaction(testPrivate, defaultAccounts[1], params.OrdinaryTx, big.NewInt(100), common.Big1, 0, 1)
	block.Header.TxRoot = types.DeriveTxsSha(block.Txs)
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// balance not enough
	block = createNewBlock()
	balance := p.am.GetAccount(testAddr).GetBalance()
	block.Txs[0] = makeTx(testPrivate, defaultAccounts[1], params.OrdinaryTx, new(big.Int).Add(balance, big.NewInt(1)))
	block.Header.TxRoot = types.DeriveTxsSha(block.Txs)
	_, err = p.Process(block)
	assert.Equal(t, ErrInvalidTxInBlock, err)

	// TODO test create or call contract fail
}

func createNewBlock() *types.Block {
	db := store.NewChainDataBase(store.GetStorePath(), store.DRIVER_MYSQL, store.DNS_MYSQL)
	return makeBlock(db, blockInfo{
		height:     2,
		parentHash: defaultBlocks[1].Hash(),
		author:     testAddr,
		txList: []*types.Transaction{
			makeTx(testPrivate, defaultAccounts[1], params.OrdinaryTx, big.NewInt(100)),
		}}, false)
}

// test tx picking logic
func TestTxProcessor_ApplyTxs(t *testing.T) {
	store.ClearData()
	p := NewTxProcessor(newChain())

	// 1 txs
	header := defaultBlocks[2].Header
	txs := defaultBlocks[2].Txs
	emptyHeader := &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	newHeader, selectedTxs, invalidTxs, err := p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, header.Bloom, newHeader.Bloom)
	assert.Equal(t, header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, header.LogRoot, newHeader.LogRoot)
	assert.Empty(t, newHeader.DeputyRoot)
	assert.Equal(t, header.Hash(), newHeader.Hash())
	assert.Equal(t, len(txs), len(selectedTxs))
	assert.Equal(t, 0, len(invalidTxs))

	// 2 txs
	header = defaultBlocks[3].Header
	txs = defaultBlocks[3].Txs
	emptyHeader = &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	newHeader, selectedTxs, invalidTxs, err = p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, header.Bloom, newHeader.Bloom)
	assert.Equal(t, header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, header.Hash(), newHeader.Hash())
	assert.Equal(t, len(txs), len(selectedTxs))
	assert.Equal(t, 0, len(invalidTxs))

	// 0 txs
	header = defaultBlocks[3].Header
	emptyHeader = &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	p.am.Reset(emptyHeader.ParentHash)
	author := p.am.GetAccount(header.MinerAddress)
	origBalance := author.GetBalance()
	newHeader, selectedTxs, invalidTxs, err = p.ApplyTxs(emptyHeader, nil)
	assert.NoError(t, err)
	assert.Equal(t, types.Bloom{}, newHeader.Bloom)
	emptyTrieHash := common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	assert.Equal(t, emptyTrieHash, newHeader.EventRoot)
	assert.Equal(t, emptyTrieHash, newHeader.TxRoot)
	assert.Equal(t, emptyTrieHash, newHeader.LogRoot)
	assert.Equal(t, 0, len(selectedTxs))
	assert.Equal(t, *origBalance, *author.GetBalance())
	assert.Equal(t, 0, len(p.am.GetChangeLogs()))

	// too many txs
	header = defaultBlocks[3].Header
	txs = defaultBlocks[3].Txs
	emptyHeader = &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     45000, // Every transaction's gasLimit is 30000. So the block only contains one transaction.
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	newHeader, selectedTxs, invalidTxs, err = p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, header.Bloom, newHeader.Bloom)
	assert.Equal(t, header.EventRoot, newHeader.EventRoot)
	assert.NotEqual(t, header.GasUsed, newHeader.GasUsed)
	assert.NotEqual(t, header.TxRoot, newHeader.TxRoot)
	assert.NotEqual(t, header.VersionRoot, newHeader.VersionRoot)
	assert.NotEqual(t, header.LogRoot, newHeader.LogRoot)
	assert.NotEqual(t, header.Hash(), newHeader.Hash())
	assert.NotEqual(t, len(txs), len(selectedTxs))
	assert.Equal(t, 0, len(invalidTxs))

	// balance not enough
	header = defaultBlocks[3].Header
	txs = defaultBlocks[3].Txs
	emptyHeader = &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	p.am.Reset(emptyHeader.ParentHash)
	balance := p.am.GetAccount(testAddr).GetBalance()
	txs = types.Transactions{
		txs[0],
		makeTx(testPrivate, defaultAccounts[1], params.OrdinaryTx, new(big.Int).Add(balance, big.NewInt(1))),
		txs[1],
	}
	newHeader, selectedTxs, invalidTxs, err = p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, header.Bloom, newHeader.Bloom)
	assert.Equal(t, header.EventRoot, newHeader.EventRoot)
	assert.Equal(t, header.GasUsed, newHeader.GasUsed)
	assert.Equal(t, header.TxRoot, newHeader.TxRoot)
	assert.Equal(t, header.VersionRoot, newHeader.VersionRoot)
	assert.Equal(t, header.LogRoot, newHeader.LogRoot)
	assert.Equal(t, header.Hash(), newHeader.Hash())
	assert.Equal(t, len(txs)-1, len(selectedTxs))
	assert.Equal(t, 1, len(invalidTxs))
}

// TODO move these cases to evm
// test different transactions
func TestTxProcessor_ApplyTxs2(t *testing.T) {
	store.ClearData()
	p := NewTxProcessor(newChain())

	// transfer to other
	header := defaultBlocks[3].Header
	emptyHeader := &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	p.am.Reset(emptyHeader.ParentHash)
	senderBalance := p.am.GetAccount(testAddr).GetBalance()
	minerBalance := p.am.GetAccount(defaultAccounts[0]).GetBalance()
	recipientBalance := p.am.GetAccount(defaultAccounts[1]).GetBalance()
	txs := types.Transactions{
		makeTx(testPrivate, defaultAccounts[1], params.OrdinaryTx, common.Big1),
	}
	newHeader, _, _, err := p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, params.TxGas, newHeader.GasUsed)
	newSenderBalance := p.am.GetAccount(testAddr).GetBalance()
	newMinerBalance := p.am.GetAccount(defaultAccounts[0]).GetBalance()
	newRecipientBalance := p.am.GetAccount(defaultAccounts[1]).GetBalance()
	cost := txs[0].GasPrice().Mul(txs[0].GasPrice(), big.NewInt(int64(params.TxGas)))
	senderBalance.Sub(senderBalance, cost)
	senderBalance.Sub(senderBalance, common.Big1)
	assert.Equal(t, senderBalance, newSenderBalance)
	assert.Equal(t, minerBalance.Add(minerBalance, cost), newMinerBalance)
	assert.Equal(t, recipientBalance.Add(recipientBalance, common.Big1), newRecipientBalance)

	// transfer to self, only cost gas
	header = defaultBlocks[3].Header
	emptyHeader = &types.Header{
		ParentHash:   header.ParentHash,
		MinerAddress: header.MinerAddress,
		Height:       header.Height,
		GasLimit:     header.GasLimit,
		GasUsed:      header.GasUsed,
		Time:         header.Time,
	}
	p.am.Reset(emptyHeader.ParentHash)
	senderBalance = p.am.GetAccount(testAddr).GetBalance()
	txs = types.Transactions{
		makeTx(testPrivate, testAddr, params.OrdinaryTx, common.Big1),
	}
	newHeader, _, _, err = p.ApplyTxs(emptyHeader, txs)
	assert.NoError(t, err)
	assert.Equal(t, params.TxGas, newHeader.GasUsed)
	newSenderBalance = p.am.GetAccount(testAddr).GetBalance()
	cost = txs[0].GasPrice().Mul(txs[0].GasPrice(), big.NewInt(int64(params.TxGas)))
	assert.Equal(t, senderBalance.Sub(senderBalance, cost), newSenderBalance)
}

// Test_applyTx 测试执行交易的函数
func Test_CandidateVoteTx(t *testing.T) {
	p := NewTxProcessor(newChain())
	header := defaultBlocks[3].Header
	p.am.Reset(header.ParentHash)
	account00 := p.am.GetAccount(defaultAccounts[0])
	// // 初始化defaultAccounts[0]为候选节点
	// profile := account00.GetCandidateProfile()
	// _, ok := profile[types.CandidateKeyIsCandidate]
	// if !ok {
	// 	profile = make(types.CandidateProfile, 5)
	// 	profile[types.CandidateKeyIsCandidate] = "true"
	// 	account00.SetCandidateProfile(profile)
	// }

	// 为投票者初始化balance
	testAccount := p.am.GetAccount(testAddr)
	testAccount.SetBalance(big.NewInt(1000000000000))
	voteTx := makeTx(testPrivate, defaultAccounts[0], params.VoteTx, big.NewInt(0))
	// 为申请者初始化balance
	account03 := p.am.GetAccount(defaultAccounts[3])
	account03.SetBalance(new(big.Int).Mul(params.RegisterCandidateNodeFees, big.NewInt(2))) // 2000LEMO
	// 申请候选节点交易
	candidata := &deputynode.CandidateNode{
		MinerAddress: common.HexToAddress("0x20000"),
		NodeID:       common.FromHex("0x34f0df789b46e9bc09f23d5315b951bc77bbfeda653ae6f5aab564c9b4619322fddb3b1f28d1c434250e9d4dd8f51aa8334573d7281e4d63baba913e9fa6908f"),
		Host:         "0.0.0.1",
		Port:         8080,
	}
	data, _ := json.Marshal(candidata)

	tx := types.NewTransaction(defaultAccounts[3], params.RegisterCandidateNodeFees, 220000, common.Big1, data, params.RegisterTx, chainID, uint64(time.Now().Unix()+300), "", "")
	sign_RegisterTx := signTransaction(tx, testPrivate)

	txs := types.Transactions{sign_RegisterTx, voteTx}

	p.ApplyTxs(header, txs)

	votes := account00.GetVotes()
	log.Println("vote=", votes)
	pro := account03.GetCandidateProfile()
	log.Printf("nodeid = %s,ip = %s,port = %s,miner = %s\n", pro[types.CandidateKeyNodeID], pro[types.CandidateKeyHost], pro[types.CandidateKeyPort], pro[types.CandidateKeyMinerAddress])
}
