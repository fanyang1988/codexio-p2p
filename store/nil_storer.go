package store

import (
	"encoding/hex"
	"sync"

	"github.com/fanyang1988/codexio-p2p/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NilStorer a nil BlockStorer
type NilStorer struct {
	// TODO: use a common storer base
	chainID types.Checksum256
	logger  *zap.Logger
	mutex   sync.RWMutex
	state   *BlockDBState
}

// NewNilStorer create a new NilStorer
func NewNilStorer(logger *zap.Logger, chainID string) (BlockStorer, error) {
	cID, err := hex.DecodeString(chainID)
	if err != nil {
		return nil, errors.Wrapf(err, "decode chainID error")
	}

	return &NilStorer{
		chainID: cID,
		logger:  logger,
		state:   NewBlockDBState(cID),
	}, nil
}

// ChainID get chainID
func (n *NilStorer) ChainID() types.Checksum256 {
	return n.chainID
}

// HeadBlockNum get headBlockNum
func (n *NilStorer) HeadBlockNum() uint32 {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.state.HeadBlockNum
}

// CommitBlock do nothing
func (n *NilStorer) CommitBlock(blk *types.SignedBlock) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if err := n.updateStatByBlock(blk); err != nil {
		return err
	}
	return nil
}

func (n *NilStorer) updateStatByBlock(blk *types.SignedBlock) error {
	// Just set to block state
	if n.state.HeadBlockNum >= blk.BlockNumber() {
		return nil
	}

	// n.logger.Info("up block", zap.Uint32("blockNum", blk.BlockNumber()))

	n.state.HeadBlockNum = blk.BlockNumber()
	n.state.HeadBlockID, _ = blk.BlockID()
	n.state.HeadBlockTime = blk.Timestamp.Time
	n.state.HeadBlock, _ = types.DeepCopyBlock(blk)

	if n.state.HeadBlockNum%1000 == 0 {
		n.logger.Info("on block head", zap.Uint32("blockNum", n.state.HeadBlockNum))
	}

	n.state.LastBlocks = append(n.state.LastBlocks, n.state.HeadBlock)
	if len(n.state.LastBlocks) >= maxBlocksHoldInDBStat {
		for i := 0; i < len(n.state.LastBlocks)-1; i++ {
			n.state.LastBlocks[i] = n.state.LastBlocks[i+1]
		}
		n.state.LastBlocks = n.state.LastBlocks[:len(n.state.LastBlocks)-1]
	}

	return nil
}

// State get state
func (n *NilStorer) State() BlockDBState {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return *n.state
}

// GetBlockByNum return nth
func (n *NilStorer) GetBlockByNum(blockNum uint32) (*types.SignedBlock, bool) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	if blockNum == n.state.HeadBlockNum {
		return n.state.HeadBlock, true
	}
	return nil, false
}

func (n *NilStorer) Close() {

}

func (n *NilStorer) Wait() {

}
