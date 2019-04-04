package deputynode

import (
	"bytes"
	"errors"
	"github.com/LemoFoundationLtd/lemochain-core/chain/params"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"sync"
)

const (
	// max deputy count
	TotalCount = 5
)

var (
	ErrEmptyDeputies         = errors.New("can't save empty deputy nodes")
	ErrInvalidDeputyRank     = errors.New("deputy node's rank is not right")
	ErrExistSnapshotHeight   = errors.New("exist snapshot block height")
	ErrInvalidSnapshotHeight = errors.New("invalid snapshot block height")
	ErrNoDeputies            = errors.New("can't access deputy nodes before SaveSnapshot")
	ErrNotDeputy             = errors.New("not a deputy address in specific height")
	ErrMineGenesis           = errors.New("can not mine genesis block")
)

//go:generate gencodec -type TermRecord --field-override termRecordMarshaling -out gen_term_record_json.go
type TermRecord struct {
	// 0, 100W+1K+1, 200W+1K+1, 300W+1K+1, 400W+1K+1...
	StartHeight uint32      `json:"height"`
	Nodes       DeputyNodes `json:"nodes"`
}

type termRecordMarshaling struct {
	StartHeight hexutil.Uint32
}

// Manager 代理节点管理器
type Manager struct {
	termList []*TermRecord
	lock     sync.Mutex
}

var managerInstance = &Manager{
	termList: make([]*TermRecord, 0, 1),
}

func Instance() *Manager {
	return managerInstance
}

// SaveSnapshot add deputy nodes record by snapshot block data
func (m *Manager) SaveSnapshot(snapshotHeight uint32, nodes DeputyNodes) {
	// check snapshotHeight
	if snapshotHeight%params.InterimDuration != 0 {
		log.Error("invalid snapshot block height", "height", snapshotHeight)
		panic(ErrInvalidSnapshotHeight)
	}
	// check nodes to make sure it is not empty
	if nodes == nil || len(nodes) == 0 {
		log.Error("can't save empty deputy nodes", "height", snapshotHeight)
		panic(ErrEmptyDeputies)
	}
	// check nodes' rank
	for i, node := range nodes {
		if uint32(i) != node.Rank {
			log.Error("invalid deputy rank", "rank", node.Rank, "expect", i)
			panic(ErrInvalidDeputyRank)
		}
	}

	// compute term start block height
	var termStart uint32
	if snapshotHeight == 0 {
		termStart = 0
	} else {
		termStart = snapshotHeight + params.InterimDuration + 1
	}

	// TODO sort nodes

	// save
	err := m.addDeputyRecord(&TermRecord{StartHeight: termStart, Nodes: nodes})
	if err != nil {
		if err == ErrExistSnapshotHeight {
			log.Warn("ignore exist snapshot block error")
		} else {
			panic(err)
		}
	}
}

// addDeputyRecord add a deputy nodes record
func (m *Manager) addDeputyRecord(record *TermRecord) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	// TODO may exist when fork at term snapshot height
	for _, term := range m.termList {
		if record.StartHeight <= term.StartHeight {
			log.Warn("exist snapshot block height", "new record height", record.StartHeight, "exit height", term.StartHeight)
			return ErrExistSnapshotHeight
		}
	}
	// TODO check if skip term
	m.termList = append(m.termList, record)

	log.Info("new deputy nodes", "start height", record.StartHeight, "nodes count", len(record.Nodes))
	return nil
}

// GetDeputiesByHeight 通过height获取对应的节点列表
func (m *Manager) GetDeputiesByHeight(height uint32, total bool) DeputyNodes {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.termList == nil || len(m.termList) == 0 {
		panic(ErrNoDeputies)
	}

	// find record
	var record *TermRecord
	if len(m.termList) == 1 {
		record = m.termList[0]
	} else {
		for i := 0; i < len(m.termList)-1; i++ {
			nextTermStart := m.termList[i+1].StartHeight
			// the height is after next term
			if height >= nextTermStart {
				continue
			}
			// the height is in current term
			record = m.termList[i]
			break
		}
		// the height is after last term
		if record == nil {
			record = m.termList[len(m.termList)-1]
		}
	}

	// find nodes
	nodes := record.Nodes
	// if not total, then result nodes must be less than TotalCount
	if !total && len(nodes) > TotalCount {
		nodes = nodes[:TotalCount]
	}

	return nodes[:]
}

// GetDeputiesCount 获取共识节点数量
func (m *Manager) GetDeputiesCount(height uint32) int {
	nodes := m.GetDeputiesByHeight(height, false)
	return len(nodes)
}

// GetDeputyByAddress 获取address对应的节点
func (m *Manager) GetDeputyByAddress(height uint32, addr common.Address) *DeputyNode {
	nodes := m.GetDeputiesByHeight(height, false)
	return findDeputyByAddress(nodes, addr)
}

// GetDeputyByNodeID 根据nodeID获取对应的节点
func (m *Manager) GetDeputyByNodeID(height uint32, nodeID []byte) *DeputyNode {
	nodes := m.GetDeputiesByHeight(height, false)
	for _, node := range nodes {
		if bytes.Compare(node.NodeID, nodeID) == 0 {
			return node
		}
	}
	return nil
}

func findDeputyByAddress(deputies []*DeputyNode, addr common.Address) *DeputyNode {
	for _, node := range deputies {
		if node.MinerAddress == addr {
			return node
		}
	}
	return nil
}

// GetMinerDistance get miner index distance in same term
func (m *Manager) GetMinerDistance(targetHeight uint32, lastBlockMiner, targetMiner common.Address) (uint32, error) {
	if targetHeight == 0 {
		return 0, ErrMineGenesis
	}
	termDeputies := m.GetDeputiesByHeight(targetHeight, false)

	// find target block miner deputy
	targetDeputy := findDeputyByAddress(termDeputies, targetMiner)
	if targetDeputy == nil {
		return 0, ErrNotDeputy
	}

	// only one deputy
	nodeCount := uint32(len(termDeputies))
	if nodeCount == 1 {
		return 1, nil
	}

	// Genesis block is pre-set, not belong to any deputy node. So only blocks start with height 1 is mined by deputies
	// The reward block changes deputy nodes, so we need recompute the slot
	if targetHeight == 1 || m.IsRewardBlock(targetHeight) {
		return targetDeputy.Rank + 1, nil
	}

	// find last block miner deputy
	lastDeputy := findDeputyByAddress(termDeputies, lastBlockMiner)
	if lastDeputy == nil {
		return 0, ErrNotDeputy
	}

	return (targetDeputy.Rank - lastDeputy.Rank + nodeCount) % nodeCount, nil
}

// IsRewardBlock 是否该发出块奖励了
func (m *Manager) IsRewardBlock(height uint32) bool {
	if height < params.TermDuration+params.InterimDuration+1 {
		// in genesis term
		return false
	} else if height%params.TermDuration == params.InterimDuration+1 {
		// term start block
		return true
	} else {
		// other normal block
		return false
	}
}

// IsSelfDeputyNode
func (m *Manager) IsSelfDeputyNode(height uint32) bool {
	return m.IsNodeDeputy(height, GetSelfNodeID())
}

// IsNodeDeputy
func (m *Manager) IsNodeDeputy(height uint32, nodeID []byte) bool {
	return m.GetDeputyByNodeID(height, nodeID) != nil
}

// Clear for test
func (m *Manager) Clear() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.termList = make([]*TermRecord, 0, 1)
}

// Clear for test
func (m *Manager) GetTermList() []*TermRecord {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.termList[:]
}
