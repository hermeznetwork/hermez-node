package coordinator

import (
	"fmt"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree/db"
)

// PurgerCfg is the purger configuration
type PurgerCfg struct {
	// PurgeBatchDelay is the delay between batches to purge outdated transactions
	PurgeBatchDelay int64
	// InvalidateBatchDelay is the delay between batches to mark invalid transactions
	InvalidateBatchDelay int64
	// PurgeBlockDelay is the delay between blocks to purge outdated transactions
	PurgeBlockDelay int64
	// InvalidateBlockDelay is the delay between blocks to mark invalid transactions
	InvalidateBlockDelay int64
}

// Purger manages cleanup of transactions in the pool
type Purger struct {
	cfg                 PurgerCfg
	lastPurgeBlock      int64
	lastPurgeBatch      int64
	lastInvalidateBlock int64
	lastInvalidateBatch int64
}

// CanPurge returns true if it's a good time to purge according to the
// configuration
func (p *Purger) CanPurge(blockNum, batchNum int64) bool {
	if blockNum > p.lastPurgeBlock+p.cfg.PurgeBlockDelay {
		return true
	}
	if batchNum > p.lastPurgeBatch+p.cfg.PurgeBatchDelay {
		return true
	}
	return false
}

// CanInvalidate returns true if it's a good time to invalidate according to
// the configuration
func (p *Purger) CanInvalidate(blockNum, batchNum int64) bool {
	if blockNum > p.lastInvalidateBlock+p.cfg.InvalidateBlockDelay {
		return true
	}
	if batchNum > p.lastInvalidateBatch+p.cfg.InvalidateBatchDelay {
		return true
	}
	return false
}

// PurgeMaybe purges txs if it's a good time to do so
func (p *Purger) PurgeMaybe(l2DB *l2db.L2DB, blockNum, batchNum int64) (bool, error) {
	if !p.CanPurge(blockNum, batchNum) {
		return false, nil
	}
	p.lastPurgeBlock = blockNum
	p.lastPurgeBatch = batchNum
	log.Debugw("Purger: purging l2txs in pool", "block", blockNum, "batch", batchNum)
	err := l2DB.Purge(common.BatchNum(batchNum))
	return true, tracerr.Wrap(err)
}

// InvalidateMaybe invalidates txs if it's a good time to do so
func (p *Purger) InvalidateMaybe(l2DB *l2db.L2DB, stateDB *statedb.LocalStateDB,
	blockNum, batchNum int64) (bool, error) {
	if !p.CanInvalidate(blockNum, batchNum) {
		return false, nil
	}
	p.lastInvalidateBlock = blockNum
	p.lastInvalidateBatch = batchNum
	log.Debugw("Purger: invalidating l2txs in pool", "block", blockNum, "batch", batchNum)
	err := poolMarkInvalidOldNonces(l2DB, stateDB, common.BatchNum(batchNum))
	return true, tracerr.Wrap(err)
}

//nolint:unused,deadcode
func idxsNonceFromL2Txs(txs []common.L2Tx) []common.IdxNonce {
	idxNonceMap := map[common.Idx]common.Nonce{}
	for _, tx := range txs {
		if nonce, ok := idxNonceMap[tx.FromIdx]; !ok {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		} else if tx.Nonce > nonce {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		}
	}
	idxsNonce := make([]common.IdxNonce, 0, len(idxNonceMap))
	for idx, nonce := range idxNonceMap {
		idxsNonce = append(idxsNonce, common.IdxNonce{Idx: idx, Nonce: nonce})
	}
	return idxsNonce
}

func idxsNonceFromPoolL2Txs(txs []common.PoolL2Tx) []common.IdxNonce {
	idxNonceMap := map[common.Idx]common.Nonce{}
	for _, tx := range txs {
		if nonce, ok := idxNonceMap[tx.FromIdx]; !ok {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		} else if tx.Nonce > nonce {
			idxNonceMap[tx.FromIdx] = tx.Nonce
		}
	}
	idxsNonce := make([]common.IdxNonce, 0, len(idxNonceMap))
	for idx, nonce := range idxNonceMap {
		idxsNonce = append(idxsNonce, common.IdxNonce{Idx: idx, Nonce: nonce})
	}
	return idxsNonce
}

// poolMarkInvalidOldNoncesFromL2Txs marks as invalid the txs in the pool that
// contain nonces equal or older to the highest nonce used in a forged l2Tx for
// the
// corresponding sender account
func poolMarkInvalidOldNoncesFromL2Txs(l2DB *l2db.L2DB,
	idxsNonce []common.IdxNonce, batchNum common.BatchNum) error {
	return l2DB.CheckNonces(idxsNonce, batchNum)
}

// poolMarkInvalidOldNonces marks as invalid txs in the pool that contain
// nonces equal or older to the nonce of the corresponding sender account
func poolMarkInvalidOldNonces(l2DB *l2db.L2DB, stateDB *statedb.LocalStateDB,
	batchNum common.BatchNum) error {
	idxs, err := l2DB.GetPendingUniqueFromIdxs()
	if err != nil {
		return tracerr.Wrap(err)
	}
	idxsNonce := make([]common.IdxNonce, len(idxs))
	lastIdx, err := stateDB.GetIdx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	for i, idx := range idxs {
		acc, err := stateDB.GetAccount(idx)
		if err != nil {
			if tracerr.Unwrap(err) != db.ErrNotFound {
				return tracerr.Wrap(err)
			} else if idx <= lastIdx {
				return tracerr.Wrap(fmt.Errorf("account with idx %v not found: %w", idx, err))
			} else {
				return tracerr.Wrap(fmt.Errorf("unexpected stateDB error with idx %v: %w", idx, err))
			}
		}
		fmt.Printf("DBG acc: %#v\n", acc)
		idxsNonce[i].Idx = idx
		idxsNonce[i].Nonce = acc.Nonce
	}
	return l2DB.CheckNonces(idxsNonce, batchNum)
}
