package miner

import (
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var (
	// The first deputy's private is set to "selfNodeKey" which means my miner private
	testDeputies = generateDeputies(17)
)

// GenerateDeputies generate random deputy nodes
func generateDeputies(num int) types.DeputyNodes {
	var result []*types.DeputyNode
	for i := 0; i < num; i++ {
		private, _ := crypto.GenerateKey()
		result = append(result, &types.DeputyNode{
			MinerAddress: crypto.PubkeyToAddress(private.PublicKey),
			NodeID:       crypto.PrivateKeyToNodeID(private),
			Rank:         uint32(i),
			Votes:        big.NewInt(int64(10000000000 - i)),
		})
		// let me to be the first deputy
		if i == 0 {
			deputynode.SetSelfNodeKey(private)
		}
	}
	return result
}

type testChain struct {
	currentBlock *types.Block
}

func (tc *testChain) CurrentBlock() *types.Block {
	return tc.currentBlock
}

func (tc *testChain) MineBlock(int64) {
}

func (tc *testChain) GetBlockByHeight(height uint32) (*types.Block, error) {
	return nil, store.ErrNotExist
}

func TestMiner_GetSleepTime(t *testing.T) {
	deputyCount := 3
	dm := deputynode.NewManager(deputyCount, &testChain{})
	dm.SaveSnapshot(0, testDeputies[:deputyCount])
	type testInfo struct {
		distance        uint32
		timeDistance    int64
		output          int64
		endOfMineWindow int64
	}

	var blockInterval int64 = 1000
	var mineTimeout int64 = 2000
	oneLoopTime := mineTimeout * int64(deputyCount)
	parentBlockTime := int64(time.Now().Unix()) * 1000
	miner := New(MineConfig{SleepTime: blockInterval, Timeout: mineTimeout}, nil, dm, nil)
	tests := []testInfo{
		// no network delay
		{1, 0, blockInterval, mineTimeout*1 + parentBlockTime},
		{2, 0, mineTimeout, mineTimeout*2 + parentBlockTime},
		{3, 0, mineTimeout * 2, mineTimeout*3 + parentBlockTime},

		// next miner
		{1, 10, blockInterval - 10, mineTimeout*1 + parentBlockTime},
		{1, blockInterval, 0, mineTimeout*1 + parentBlockTime},
		{1, blockInterval + 10, 0, mineTimeout*1 + parentBlockTime},
		{1, mineTimeout, oneLoopTime - (mineTimeout), mineTimeout*1 + oneLoopTime + parentBlockTime},
		{1, mineTimeout + 10, oneLoopTime - (mineTimeout + 10), mineTimeout*1 + parentBlockTime + oneLoopTime},
		{1, mineTimeout*2 + 10, oneLoopTime - (mineTimeout*2 + 10), mineTimeout*1 + parentBlockTime + oneLoopTime},
		{1, oneLoopTime, 0, mineTimeout*1 + parentBlockTime + oneLoopTime},
		{1, oneLoopTime + 10, 0, mineTimeout*1 + parentBlockTime + oneLoopTime},

		// second miner
		{2, 10, mineTimeout - 10, mineTimeout*2 + parentBlockTime},
		{2, blockInterval, mineTimeout - blockInterval, mineTimeout*2 + parentBlockTime},
		{2, blockInterval + 10, mineTimeout - (blockInterval + 10), mineTimeout*2 + parentBlockTime},
		{2, mineTimeout, 0, mineTimeout*2 + parentBlockTime},
		{2, mineTimeout + 10, 0, mineTimeout*2 + parentBlockTime},
		{2, mineTimeout*2 + 10, oneLoopTime - (mineTimeout + 10), mineTimeout*2 + parentBlockTime + oneLoopTime},
		{2, oneLoopTime, mineTimeout, mineTimeout*2 + parentBlockTime + oneLoopTime},
		{2, oneLoopTime + 10, mineTimeout - 10, mineTimeout*2 + parentBlockTime + oneLoopTime},

		// self miner
		{3, 10, mineTimeout*2 - 10, mineTimeout*3 + parentBlockTime},
		{3, blockInterval, mineTimeout*2 - blockInterval, mineTimeout*3 + parentBlockTime},
		{3, blockInterval + 10, mineTimeout*2 - (blockInterval + 10), mineTimeout*3 + parentBlockTime},
		{3, mineTimeout, mineTimeout, mineTimeout*3 + parentBlockTime},
		{3, mineTimeout + 10, mineTimeout - 10, mineTimeout*3 + parentBlockTime},
		{3, mineTimeout*2 + 10, 0, mineTimeout*3 + parentBlockTime},
		{3, oneLoopTime, mineTimeout * 2, mineTimeout*3 + parentBlockTime + oneLoopTime},
		{3, oneLoopTime + 10, mineTimeout*2 - 10, mineTimeout*3 + parentBlockTime + oneLoopTime},

		// parent block is future block
		{1, -10, blockInterval - (-10), mineTimeout*1 + parentBlockTime},
		{1, -10000, blockInterval - (-10000), mineTimeout*1 + parentBlockTime},
		{2, -10, mineTimeout - (-10), mineTimeout*2 + parentBlockTime},
		{2, -10000, mineTimeout - (-10000), mineTimeout*2 + parentBlockTime},
		{3, -10, mineTimeout*2 - (-10), mineTimeout*3 + parentBlockTime},
		{3, -10000, mineTimeout*2 - (-10000), mineTimeout*3 + parentBlockTime},
	}
	for _, test := range tests {
		caseName := fmt.Sprintf("distance=%d,timeDistance=%d", test.distance, test.timeDistance)
		t.Run(caseName, func(t *testing.T) {
			test := test // capture range variable
			t.Parallel()
			waitTime, endOfMineWindow := miner.getSleepTime(1, test.distance, parentBlockTime, parentBlockTime+test.timeDistance)
			assert.Equal(t, test.output, waitTime)
			assert.Equal(t, test.endOfMineWindow, endOfMineWindow)
		})
	}
}

func TestMiner_waitCanPackageTx(t *testing.T) {
	var blockInterval int64 = 3000
	var mineTimeout int64 = 10000
	pool := &testTxpool{false}
	miner := New(MineConfig{SleepTime: blockInterval, Timeout: mineTimeout}, nil, nil, pool)

	// 1. 交易池中一开始就有可以打包的交易的情况
	pool.IsExist = true
	start := time.Now()
	minerTimeoutStamp := time.Now().UnixNano()/1e6 + 100000000
	miner.waitCanPackageTx(minerTimeoutStamp)
	// 不用等待直接返回
	assert.Equal(t, int64(0), time.Since(start).Nanoseconds()/1e9)

	// 2. 交易池中始终没有可以打包的交易情况
	pool.IsExist = false
	start = time.Now()
	minerTimeoutStamp = time.Now().UnixNano()/1e6 + 2000
	miner.waitCanPackageTx(minerTimeoutStamp)
	// 等待超时
	assert.Equal(t, int64(2), time.Since(start).Nanoseconds()/1e9)

	// 3. 交易池中等一会儿就有交易的情况
	pool.IsExist = false
	start = time.Now()
	setTime := time.Millisecond * 3000 // 3000毫秒之后交易池中有可以打包的交易了
	go func() {
		time.AfterFunc(setTime, func() {
			// 3s之后设置交易池中有交易
			pool.IsExist = true
		})
	}()
	minerTimeoutStamp = time.Now().UnixNano()/1e6 + 5000
	miner.waitCanPackageTx(minerTimeoutStamp)
	// 3000毫秒之后发现交易池中存在可以打包的交易然后就退出
	assert.Equal(t, int64(3), time.Since(start).Nanoseconds()/1e9)
}

type testTxpool struct {
	IsExist bool
}

func (p *testTxpool) ExistCanPackageTx(time uint32) bool {
	return p.IsExist
}
