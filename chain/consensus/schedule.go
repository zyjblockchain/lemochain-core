package consensus

import (
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
)

// GetMinerDistance get miner index distance. It is always greater than 0 and not greater than deputy count
func GetMinerDistance(targetHeight uint32, parentBlockMiner, targetMiner common.Address, dm *deputynode.Manager) (uint32, error) {
	if targetHeight == 0 {
		return 0, ErrMineGenesis
	}
	deputies := dm.GetDeputiesByHeight(targetHeight)
	nodeCount := uint32(len(deputies))

	// find target block miner deputy
	targetDeputy := findDeputyByAddress(deputies, targetMiner)
	if targetDeputy == nil {
		return 0, ErrNotDeputy
	}

	// Genesis block is pre-set, not belong to any deputy node. So only blocks start with height 1 is mined by deputies
	// The reward block changes deputy nodes, so we need recompute the slot
	if targetHeight == 1 || deputynode.IsRewardBlock(targetHeight) {
		return targetDeputy.Rank + 1, nil
	}

	// if they are same miner, then return deputy count
	if targetMiner == parentBlockMiner {
		return nodeCount, nil
	}

	// find last block miner deputy
	lastDeputy := findDeputyByAddress(deputies, parentBlockMiner)
	if lastDeputy == nil {
		return 0, ErrNotDeputy
	}
	return (nodeCount + targetDeputy.Rank - lastDeputy.Rank) % nodeCount, nil
}

// getDeputyByDistance find a deputy from parent block miner by miner index distance. The distance should always greater than 0
func getDeputyByDistance(targetHeight uint32, parentBlockMiner common.Address, distance uint32, dm *deputynode.Manager) (*types.DeputyNode, error) {
	if targetHeight == 0 {
		return nil, ErrMineGenesis
	}
	deputies := dm.GetDeputiesByHeight(targetHeight)
	nodeCount := uint32(len(deputies))

	// find target block miner deputy
	parentMinerDeputy := findDeputyByAddress(deputies, parentBlockMiner)
	if parentMinerDeputy == nil {
		return nil, ErrNotDeputy
	}

	var targetRank uint32
	if targetHeight == 1 || deputynode.IsRewardBlock(targetHeight) {
		// Genesis block is pre-set, not belong to any deputy node. So only blocks start with height 1 is mined by deputies
		// The reward block changes deputy nodes, so we need recompute the slot
		targetRank = distance - 1
	} else {
		// find last block miner deputy
		targetRank = (nodeCount + parentMinerDeputy.Rank + distance) % nodeCount
	}
	return findDeputyByRank(deputies, targetRank), nil
}

func findDeputyByAddress(deputies []*types.DeputyNode, addr common.Address) *types.DeputyNode {
	for _, node := range deputies {
		if node.MinerAddress == addr {
			return node
		}
	}
	return nil
}

func findDeputyByRank(deputies []*types.DeputyNode, rank uint32) *types.DeputyNode {
	for _, node := range deputies {
		if node.Rank == rank {
			return node
		}
	}
	return nil
}

// GetNextMineWindow get next time window to mine block. The times are timestamps in millisecond
func GetNextMineWindow(nextHeight uint32, distance uint32, parentTime int64, currentTime int64, mineTimeout int64, dm *deputynode.Manager) (int64, int64) {
	nodeCount := dm.GetDeputiesCount(nextHeight)
	// 所有节点都超时所需要消耗的时间，也可以看作是下一轮出块的开始时间
	oneLoopTime := int64(nodeCount) * mineTimeout
	// 网络传输耗时，即当前时间减去父块区块头中的时间戳
	passTime := currentTime - parentTime
	if passTime < 0 {
		passTime = 0
	}
	// 从父块开始，经过的整轮数
	passLoop := passTime / oneLoopTime
	// 可以出块的时间窗口
	windowFrom := parentTime + passLoop*oneLoopTime + int64(distance-1)*mineTimeout
	windowTo := parentTime + passLoop*oneLoopTime + int64(distance)*mineTimeout
	if windowTo <= currentTime {
		windowFrom += oneLoopTime
		windowTo += oneLoopTime
	}

	log.Debug("GetNextMineWindow", "windowFrom", windowFrom, "windowTo", windowTo, "parentTime", parentTime, "passTime", passTime, "distance", distance, "passLoop", passLoop, "nodeCount", nodeCount)
	return windowFrom, windowTo
}

// GetCorrectMiner get the correct miner to mine a block after parent block
func GetCorrectMiner(parent *types.Header, mineTime int64, mineTimeout int64, dm *deputynode.Manager) (common.Address, error) {
	if mineTime < 1e10 {
		panic("mineTime should be milliseconds")
	}
	passTime := mineTime - int64(parent.Time)*1000
	if passTime < 0 {
		return common.Address{}, ErrSmallerMineTime
	}
	nodeCount := dm.GetDeputiesCount(parent.Height + 1)
	// 所有节点都超时所需要消耗的时间，也可以看作是下一轮出块的开始时间
	oneLoopTime := int64(nodeCount) * mineTimeout
	minerDistance := (passTime%oneLoopTime)/mineTimeout + 1

	deputy, err := getDeputyByDistance(parent.Height+1, parent.MinerAddress, uint32(minerDistance), dm)
	if err != nil {
		return common.Address{}, err
	}
	log.Debug("GetCorrectMiner", "correctMiner", deputy.MinerAddress, "parent", parent.MinerAddress, "mineTime", mineTime, "mineTimeout", mineTimeout, "passTime", passTime, "nodeCount", nodeCount)
	return deputy.MinerAddress, nil
}