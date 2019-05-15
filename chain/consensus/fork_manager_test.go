package consensus

import (
	"crypto/ecdsa"
	"crypto/rand"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto/secp256k1"
	"github.com/LemoFoundationLtd/lemochain-core/store"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

var (
	testDeputies = GenerateDeputies(17)
)

type testBlockLoader map[uint32]*types.Block

func (loader testBlockLoader) GetBlockByHeight(height uint32) (*types.Block, error) {
	block, ok := loader[height]
	if !ok {
		return nil, store.ErrNotExist
	}
	return block, nil
}

// GenerateDeputies generate random deputy nodes
func GenerateDeputies(num int) types.DeputyNodes {
	var result []*types.DeputyNode
	for i := 0; i < num; i++ {
		private, _ := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
		result = append(result, &types.DeputyNode{
			MinerAddress: crypto.PubkeyToAddress(private.PublicKey),
			NodeID:       (crypto.FromECDSAPub(&private.PublicKey))[1:],
			Rank:         uint32(i),
			Votes:        big.NewInt(int64(10000000000 - i)),
		})
	}
	return result
}

// pickNodes picks some test deputy nodes by index
func pickNodes(nodeIndexList ...int) types.DeputyNodes {
	var result []*types.DeputyNode
	for i, nodeIndex := range nodeIndexList {
		newDeputy := testDeputies[nodeIndex].Copy()
		// reset rank
		newDeputy.Rank = uint32(i)
		result = append(result, newDeputy)
	}
	return result
}

// test special cases
func TestGetMinerDistance_Error(t *testing.T) {
	dm := deputynode.NewManager(5, testBlockLoader{})

	nodes0 := pickNodes(0, 1, 2)
	dm.SaveSnapshot(0, nodes0)
	nodes1 := pickNodes(1)
	dm.SaveSnapshot(params.TermDuration, nodes1)
	nodes2 := pickNodes(2, 3, 4, 5)
	dm.SaveSnapshot(params.TermDuration*2, nodes2)
	term0Height := uint32(10)
	term1RewardHeight := params.TermDuration + params.InterimDuration + 1
	term2RewardHeight := params.TermDuration*2 + params.InterimDuration + 1

	// height is 0
	_, err := GetMinerDistance(0, common.Address{}, common.Address{}, dm)
	assert.Equal(t, ErrMineGenesis, err)

	// not exist target miner
	_, err = GetMinerDistance(term0Height, common.Address{}, common.Address{}, dm)
	assert.Equal(t, ErrNotDeputy, err)
	_, err = GetMinerDistance(term0Height, common.Address{}, testDeputies[5].MinerAddress, dm)
	assert.Equal(t, ErrNotDeputy, err)

	// not exist last miner
	_, err = GetMinerDistance(term0Height, common.Address{}, testDeputies[0].MinerAddress, dm)
	assert.Equal(t, ErrNotDeputy, err)
	_, err = GetMinerDistance(term0Height, testDeputies[5].MinerAddress, testDeputies[0].MinerAddress, dm)
	assert.Equal(t, ErrNotDeputy, err)

	// only one deputy
	dis, err := GetMinerDistance(term1RewardHeight, common.Address{}, testDeputies[1].MinerAddress, dm)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), dis)

	// first block
	dis, err = GetMinerDistance(1, common.Address{}, testDeputies[0].MinerAddress, dm)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), dis)
	dis, err = GetMinerDistance(1, common.Address{}, testDeputies[2].MinerAddress, dm)
	assert.NoError(t, err)
	assert.Equal(t, uint32(3), dis)

	// reward block
	dis, err = GetMinerDistance(term2RewardHeight, common.Address{}, testDeputies[2].MinerAddress, dm)
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), dis)
	dis, err = GetMinerDistance(term2RewardHeight, common.Address{}, testDeputies[5].MinerAddress, dm)
	assert.NoError(t, err)
	assert.Equal(t, uint32(4), dis)
}

// test normal cases
func TestGetMinerDistance(t *testing.T) {
	dm := deputynode.NewManager(5, testBlockLoader{})

	nodes0 := pickNodes(0, 1, 2)
	dm.SaveSnapshot(0, nodes0)
	nodes1 := pickNodes(1)
	dm.SaveSnapshot(params.TermDuration, nodes1)
	nodes2 := pickNodes(2, 3, 4, 5)
	dm.SaveSnapshot(params.TermDuration*2, nodes2)
	term0Height := uint32(10)
	term2RewardHeight := params.TermDuration*2 + params.InterimDuration + 1
	term2Height := term2RewardHeight + 10

	type testDistanceData struct {
		CaseName          string
		TargetHeight      uint32
		LastDeputyIndex   int
		TargetDeputyIndex int
		ExpectDistance    uint32
	}
	var tests = []testDistanceData{
		{"[0,1,2] 2-0=2", term0Height, 0, 2, 2},
		{"[0,1,2] 0-2=1", term0Height, 2, 0, 1},
		{"[0,1,2] 2-2=0", term0Height, 2, 2, 0},
		{"[2,3,4,5] 3-2=1", term2Height, 2, 3, 1},
		{"[2,3,4,5] 4-2=2", term2Height, 2, 4, 2},
		{"[2,3,4,5] 4-4=0", term2Height, 4, 4, 0},
		{"[2,3,4,5] 2-5=1", term2Height, 5, 2, 1},
		{"[2,3,4,5] 2-3=3", term2Height, 3, 2, 3},
		{"[2,3,4,5] 2-2=0", term2Height, 2, 2, 0},
	}

	for _, test := range tests {
		t.Run(test.CaseName, func(t *testing.T) {
			test := test // capture range variable
			t.Parallel()

			lastBlockMiner := testDeputies[test.LastDeputyIndex].MinerAddress
			targetMiner := testDeputies[test.TargetDeputyIndex].MinerAddress
			dis, err := GetMinerDistance(test.TargetHeight, lastBlockMiner, targetMiner, dm)
			assert.NoError(t, err)
			assert.Equal(t, test.ExpectDistance, dis)
		})
	}
}
