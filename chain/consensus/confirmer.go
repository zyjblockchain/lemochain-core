package consensus

import (
	"bytes"
	"github.com/LemoFoundationLtd/lemochain-core/chain/deputynode"
	"github.com/LemoFoundationLtd/lemochain-core/chain/types"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/network"
)

type confirmWriter interface {
	SetConfirms(hash common.Hash, pack []types.SignData) (*types.Block, error)
}

// Confirmer process the confirm logic
type Confirmer struct {
	blockLoader  BlockLoader
	stableLoader StableBlockStore
	confirmStore confirmWriter
	dm           *deputynode.Manager
	lastSig      blockSignRecord
}

type blockSignRecord struct {
	Height uint32
	Hash   common.Hash
}

func NewConfirmer(dm *deputynode.Manager, blockLoader BlockLoader, confirmStore confirmWriter, stableLoader StableBlockStore) *Confirmer {
	confirmer := &Confirmer{
		blockLoader:  blockLoader,
		stableLoader: stableLoader,
		confirmStore: confirmStore,
		dm:           dm,
	}
	stable, _ := stableLoader.LoadLatestBlock()
	confirmer.lastSig.Height = stable.Height()
	confirmer.lastSig.Hash = stable.Hash()
	return confirmer
}

// TryConfirm try to sign and save a confirm into a received block
func (c *Confirmer) TryConfirm(block *types.Block) (types.SignData, bool) {
	if !c.needConfirm(block) {
		return types.SignData{}, false
	}

	sig, err := c.confirmBlock(block)
	if err != nil {
		return types.SignData{}, false
	}

	if block.IsConfirmExist(sig) {
		return types.SignData{}, false
	}

	block.Confirms = append(block.Confirms, sig)

	return sig, true
}

func (c *Confirmer) needConfirm(block *types.Block) bool {
	// test if we are deputy node
	if !c.dm.IsSelfDeputyNode(block.Height()) {
		return false
	}
	// test if it contains enough confirms
	if IsConfirmEnough(block, c.dm) {
		return false
	}
	// It's not necessary to test if the block was mined or been confirmed by myself. Because confirmed blocks must be in database. So they will be dropped by network module at the beginning

	// load last confirmed block
	lastConfirmHeight := c.lastSig.Height
	lastConfirmHash := c.lastSig.Hash
	stable, _ := c.stableLoader.LoadLatestBlock()
	if lastConfirmHeight < stable.Height() {
		lastConfirmHeight = stable.Height()
		lastConfirmHash = stable.Hash()
	}

	// the block is at same fork with last signed block
	if block.ParentHash() == lastConfirmHash {
		return true
	}
	// the block is deputyCount*2/3 far from signed block
	signDistance := c.dm.TwoThirdDeputyCount(block.Height())
	// not ">=" so that we would never need to confirm a new block after switch fork
	if block.Height() > lastConfirmHeight+signDistance {
		return true
	}

	log.Debug("can't confirm the block", "lastConfirm", lastConfirmHeight, "height", block.Height(), "minDistance", signDistance)
	return false
}

// BatchConfirmStable confirm and broadcast unsigned stable blocks one by one
func (c *Confirmer) BatchConfirmStable(startHeight, endHeight uint32) []*network.BlockConfirmData {
	if endHeight < startHeight {
		return nil
	}

	result := make([]*network.BlockConfirmData, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		block, err := c.blockLoader.GetBlockByHeight(i)
		if err != nil {
			log.Error("Load block fail, can't confirm it", "height", i)
			continue
		}
		if sig := c.tryConfirmStable(block); sig != nil {
			result = append(result, &network.BlockConfirmData{
				Hash:     block.Hash(),
				Height:   block.Height(),
				SignInfo: *sig,
			})
		}
	}

	return result
}

// NeedConfirmList
func (c *Confirmer) NeedConfirmList(startHeight, endHeight uint32) []network.GetConfirmInfo {
	if startHeight > endHeight {
		return nil
	}
	fetchList := make([]network.GetConfirmInfo, 0, endHeight-startHeight+1)
	for i := startHeight; i <= endHeight; i++ {
		block, err := c.blockLoader.GetBlockByHeight(i)
		if err != nil {
			log.Errorf("Load block fail, can't fetch it's confirms, height: %d", i)
			continue
		}
		if IsConfirmEnough(block, c.dm) {
			continue
		}
		info := network.GetConfirmInfo{
			Height: block.Height(),
			Hash:   block.Hash(),
		}
		fetchList = append(fetchList, info)
	}
	return fetchList
}

// SetLastSig
func (c *Confirmer) SetLastSig(block *types.Block) {
	if block.Height() > c.lastSig.Height {
		c.lastSig.Height = block.Height()
		c.lastSig.Hash = block.Hash()
	}
}

func IsMinedByself(block *types.Block) bool {
	nodeID, err := block.SignerNodeID()
	if err != nil {
		return false
	}
	return bytes.Compare(nodeID, deputynode.GetSelfNodeID()) == 0
}

// TryConfirmStable try to sign and save a confirm into a stable block
func (c *Confirmer) tryConfirmStable(block *types.Block) *types.SignData {
	// test if we are deputy node
	if !c.dm.IsSelfDeputyNode(block.Height()) {
		return nil
	}
	// test if it contains enough confirms
	if IsConfirmEnough(block, c.dm) {
		return nil
	}

	sig, err := c.confirmBlock(block)
	if err != nil {
		return nil
	}

	if block.IsConfirmExist(sig) {
		return nil
	}

	_, _ = c.SaveConfirm(block, []types.SignData{sig})
	return &sig
}

// SaveConfirm save a confirm to store, then return a new block
func (c *Confirmer) SaveConfirm(block *types.Block, sigList []types.SignData) (*types.Block, error) {
	newBlock, err := c.confirmStore.SetConfirms(block.Hash(), sigList)
	if err != nil {
		log.Errorf("SetConfirm failed: %v", err)
		return nil, err
	}
	log.Debugf("Now block %s contains %d confirms", newBlock.ShortString(), len(newBlock.Confirms))
	return newBlock, nil
}

// confirmBlock sign a block and return signData
func (c *Confirmer) confirmBlock(block *types.Block) (types.SignData, error) {
	sig, err := SignBlock(block.Hash())
	if err != nil {
		log.Error("sign for confirm data error", "err", err)
		return types.SignData{}, err
	}
	c.SetLastSig(block)
	return types.BytesToSignData(sig), nil
}
