package txpool

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTxPool_RecvTx(t *testing.T) {
	curTime := time.Now().Unix()
	pool := NewTxPool()

	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	pool.RecvTx(tx1)
	pool.RecvTx(tx2)
	pool.RecvTx(tx3)

	result := pool.Get(uint32(curTime), 10)
	assert.Equal(t, 3, len(result))

	result = pool.Get(uint32(curTime), 10)
	assert.Equal(t, 3, len(result))
}

func TestTxPool_DelInvalidTxs(t *testing.T) {
	curTime := time.Now().Unix()
	pool := NewTxPool()

	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	tx4 := makeTxRandom(common.HexToAddress("0x04"))
	pool.RecvTx(tx1)
	pool.RecvTx(tx2)
	pool.RecvTx(tx3)
	pool.RecvTx(tx4)

	delTxs := make([]*types.Transaction, 0, 3)
	delTxs = append(delTxs, tx1)
	delTxs = append(delTxs, tx2)
	delTxs = append(delTxs, tx3)
	pool.DelInvalidTxs(delTxs)

	result := pool.Get(uint32(curTime), 10)
	assert.Equal(t, 1, len(result))
}

func TestTxPool_RecvBlock(t *testing.T) {
	curTime := time.Now().Unix()

	pool := NewTxPool()
	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	tx4 := makeTxRandom(common.HexToAddress("0x04"))
	tx5 := makeTxRandom(common.HexToAddress("0x05"))
	tx6 := makeTxRandom(common.HexToAddress("0x06"))
	tx7 := makeTxRandom(common.HexToAddress("0x07"))
	tx8 := makeTxRandom(common.HexToAddress("0x08"))
	tx9 := makeTxRandom(common.HexToAddress("0x09"))
	pool.RecvTx(tx1)
	pool.RecvTx(tx2)
	pool.RecvTx(tx3)
	pool.RecvTx(tx4)
	pool.RecvTx(tx5)
	pool.RecvTx(tx6)
	pool.RecvTx(tx7)
	pool.RecvTx(tx8)
	pool.RecvTx(tx9)

	block1 := store.GetBlock1()
	block1.Header.Time = uint32(curTime)
	block1.Txs = append(block1.Txs, tx1)
	block1.Txs = append(block1.Txs, tx2)
	block1.Txs = append(block1.Txs, tx3)
	block1.Txs = append(block1.Txs, tx4)
	block1.Txs = append(block1.Txs, tx5)
	block1.Txs = append(block1.Txs, tx6)
	block1.Txs = append(block1.Txs, tx7)
	pool.RecvBlock(block1)

	result := pool.Get(uint32(curTime), 10)
	assert.Equal(t, 2, len(result))
}

func TestTxPool_PruneBlock(t *testing.T) {
	curTime := time.Now().Unix()

	pool := NewTxPool()
	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	tx4 := makeTxRandom(common.HexToAddress("0x04"))
	tx5 := makeTxRandom(common.HexToAddress("0x05"))
	tx6 := makeTxRandom(common.HexToAddress("0x06"))
	tx7 := makeTxRandom(common.HexToAddress("0x07"))
	tx8 := makeTxRandom(common.HexToAddress("0x08"))
	tx9 := makeTxRandom(common.HexToAddress("0x09"))
	pool.RecvTx(tx8)
	pool.RecvTx(tx9)

	block1 := store.GetBlock1()
	block1.Header.Time = uint32(curTime)
	block1.Txs = append(block1.Txs, tx1)
	block1.Txs = append(block1.Txs, tx2)
	block1.Txs = append(block1.Txs, tx3)
	block1.Txs = append(block1.Txs, tx4)
	block1.Txs = append(block1.Txs, tx5)
	block1.Txs = append(block1.Txs, tx6)
	block1.Txs = append(block1.Txs, tx7)
	pool.PruneBlock(block1)

	result := pool.Get(uint32(curTime), 10)
	assert.Equal(t, 9, len(result))
}

func TestTxPool_VerifyTxInBlock1(t *testing.T) {
	curTime := time.Now().Unix()

	pool := NewTxPool()
	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	tx4 := makeTxRandom(common.HexToAddress("0x04"))
	tx5 := makeTxRandom(common.HexToAddress("0x05"))
	tx6 := makeTxRandom(common.HexToAddress("0x06"))
	tx7 := makeTxRandom(common.HexToAddress("0x07"))
	tx8 := makeTxRandom(common.HexToAddress("0x08"))
	tx9 := makeTxRandom(common.HexToAddress("0x09"))

	block1 := store.GetBlock1()
	block1.Header.Time = uint32(curTime)
	block1.Txs = append(block1.Txs, tx1)
	block1.Txs = append(block1.Txs, tx2)
	block1.Txs = append(block1.Txs, tx3)
	pool.RecvBlock(block1)

	block2 := store.GetBlock2()
	block2.Header.ParentHash = block1.Hash()
	block2.Header.Time = uint32(curTime)
	block2.Txs = append(block2.Txs, tx4)
	block2.Txs = append(block2.Txs, tx5)
	block2.Txs = append(block2.Txs, tx6)
	block2.Txs = append(block2.Txs, tx7)
	pool.RecvBlock(block2)

	block3 := store.GetBlock3()
	block3.Header.ParentHash = block2.Hash()
	block3.Header.Time = uint32(curTime)
	block3.Txs = append(block3.Txs, tx7)
	block3.Txs = append(block3.Txs, tx8)
	block3.Txs = append(block3.Txs, tx9)

	isValid := pool.VerifyTxInBlock(block3)
	assert.Equal(t, false, isValid)

	tx10 := makeTxRandom(common.HexToAddress("0x10"))
	block4 := store.GetBlock4()
	block4.Header.ParentHash = block3.Hash()
	block4.Header.Time = uint32(curTime)
	block4.Txs = append(block4.Txs, tx10)

	isValid = pool.VerifyTxInBlock(block4)
	assert.Equal(t, true, isValid)
}

func TestTxPool_VerifyTxInBlock2(t *testing.T) {
	curTime := time.Now().Unix()

	pool := NewTxPool()
	tx1 := makeTxRandom(common.HexToAddress("0x01"))
	tx2 := makeTxRandom(common.HexToAddress("0x02"))
	tx3 := makeTxRandom(common.HexToAddress("0x03"))
	tx4 := makeTxRandom(common.HexToAddress("0x04"))
	tx5 := makeTxRandom(common.HexToAddress("0x05"))
	tx6 := makeTxRandom(common.HexToAddress("0x06"))
	tx7 := makeTxRandom(common.HexToAddress("0x07"))
	tx8 := makeTxRandom(common.HexToAddress("0x08"))

	block1 := store.GetBlock1()
	block1.Header.Time = uint32(curTime)
	block1.Txs = append(block1.Txs, tx1)
	block1.Txs = append(block1.Txs, tx2)
	pool.RecvBlock(block1)

	block21 := store.GetBlock2()
	block21.Header.ParentHash = block1.Hash()
	block21.Header.Time = uint32(curTime)
	block21.Txs = append(block21.Txs, tx3)
	block21.Txs = append(block21.Txs, tx4)
	pool.RecvBlock(block21)

	block22 := store.GetBlock2()
	block22.Header.ParentHash = block1.Hash()
	block22.Header.Time = uint32(curTime + 1)
	block22.Txs = append(block22.Txs, tx5)
	block22.Txs = append(block22.Txs, tx6)
	pool.RecvBlock(block22)

	block3 := store.GetBlock3()
	block3.Header.ParentHash = block21.Hash()
	block3.Header.Time = uint32(curTime)
	block3.Txs = append(block3.Txs, tx7)
	block3.Txs = append(block3.Txs, tx8)

	isValid := pool.VerifyTxInBlock(block3)
	assert.Equal(t, true, isValid)

	block3 = store.GetBlock3()
	block3.Header.ParentHash = block21.Hash()
	block3.Header.Time = uint32(curTime)
	block3.Txs = append(block3.Txs, tx5)
	block3.Txs = append(block3.Txs, tx8)

	isValid = pool.VerifyTxInBlock(block3)
	assert.Equal(t, true, isValid)

	block3 = store.GetBlock3()
	block3.Header.ParentHash = block22.Hash()
	block3.Header.Time = uint32(curTime)
	block3.Txs = append(block3.Txs, tx5)
	block3.Txs = append(block3.Txs, tx8)

	isValid = pool.VerifyTxInBlock(block3)
	assert.Equal(t, false, isValid)
}

func TestTxPool_Box(t *testing.T) {
	log.Setup(log.LevelDebug, false, false)

	curTime := time.Now().Unix()
	tx := createBoxTxRandom(common.HexToAddress("0xabcde"), 5, uint64(curTime))

	pool := NewTxPool()
	pool.RecvTx(tx)

	result := pool.Get(uint32(curTime), 10)
	assert.Equal(t, 1, len(result))

	pool.RecvTx(tx)
	result = pool.Get(uint32(curTime), 10)
	assert.Equal(t, 1, len(result))
}
