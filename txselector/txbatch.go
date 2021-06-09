package txselector

import (
	"fmt"
	"sort"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/hermez-node/txprocessor"
)

type (
	// TxBatch represents the future batch transactions, composed by TxGroup's
	TxBatch struct {
		l2db            l2DB
		localAccountsDB stateDB
		processor       txProcessor
		txs             []*TxGroup
		l1UserTxs       []common.L1Tx
		l1UserFutureTxs []common.L1Tx
		selectionConfig txprocessor.Config
		coordAccount    CoordAccount
	}
)

// NewTxBatch creates a new *TxBatch object
func NewTxBatch(selectionConfig txprocessor.Config, l2db l2DB, coordAccount CoordAccount,
	localAccountsDB stateDB, processor txProcessor) (*TxBatch, error) {
	return &TxBatch{
		txs:             nil,
		l1UserTxs:       nil,
		l1UserFutureTxs: nil,
		selectionConfig: selectionConfig,
		l2db:            l2db,
		localAccountsDB: localAccountsDB,
		processor:       processor,
		coordAccount:    coordAccount,
	}, nil
}

// getSelection returns the coordIdxs, auths, l1UserTxs, l1CoordTxs, poolL2Txs and discardedL2Txs
// selected or created to the next batch inside the TxGroup's
func (b *TxBatch) getSelection() ([]common.Idx, [][]byte, []common.L1Tx, []common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	coordIdxs := make([]common.Idx, 0)
	alreadyAddedCoordIdx := make(map[common.Idx]bool)
	auths := make([][]byte, 0)
	l1UserTxs := make([]common.L1Tx, 0)
	l1CoordTxs := make([]common.L1Tx, 0)
	poolL2Txs := make([]common.PoolL2Tx, 0)
	discardedL2Txs := make([]common.PoolL2Tx, 0)

	// iterate from the TxGroup's to append all returns
	l1UserTxs = append(l1UserTxs, b.l1UserTxs...)
	for _, group := range b.txs {
		// Add idx for the coordinator to get the fee, if maximum is not reached
		// and that idx is not already used
		// TODO: avoid repeating idxs of the same tokenID? (may already be the case)
		// Ideal situation: maximize the fee value by having b.selectionConfig.MaxFeeTx idxs
		// all of them linked to different tokenIDs in a way that the selected tokenIDs have the maximum fee value
		// among the different tokenIDs that can be used to get fees
		if len(coordIdxs) <= int(b.selectionConfig.MaxFeeTx) {
			for _, idx := range group.coordIdxsMap {
				if _, ok := alreadyAddedCoordIdx[idx]; !ok {
					coordIdxs = append(coordIdxs, idx)
					alreadyAddedCoordIdx[idx] = true
				}
			}
		}

		auth := group.accAuths
		auths = append(auths, auth...)

		l1UserTxs = append(l1UserTxs, group.l1UserTxs...)
		l1CoordTxs = append(l1CoordTxs, group.l1CoordTxs...)
		poolL2Txs = append(poolL2Txs, group.l2Txs...)
		discardedL2Txs = append(discardedL2Txs, group.discardedTxs...)
	}
	sort.SliceStable(coordIdxs, func(i, j int) bool {
		return coordIdxs[i] < coordIdxs[j]
	})

	metric.SelectedL1CoordinatorTxs.Set(float64(len(l1CoordTxs)))
	metric.SelectedL1UserTxs.Set(float64(len(l1UserTxs)))
	metric.SelectedL2Txs.Set(float64(len(poolL2Txs)))
	metric.DiscardedL2Txs.Set(float64(len(discardedL2Txs)))

	return coordIdxs, auths, l1UserTxs, l1CoordTxs, poolL2Txs, discardedL2Txs, nil
}

// length returns all transactions count
func (b *TxBatch) length() int {
	length := 0
	for _, tx := range b.txs {
		length += tx.length()
	}
	return length + len(b.l1UserTxs)
}

// last returns the last TxGroup
func (b *TxBatch) last() *TxGroup {
	if len(b.txs) == 0 {
		return nil
	}
	return b.txs[len(b.txs)-1]
}

// sort sort all transactions by the most profitable
func (b *TxBatch) sort() {
	sort.Slice(b.txs, func(i, j int) bool {
		txI := b.txs[i]
		txJ := b.txs[j]
		// atomic transactions always first
		if txI.atomic != txJ.atomic {
			return txI.atomic
		}
		// sort by the highest fee
		return txI.feeAverage.Cmp(txJ.feeAverage) > 0
	})
}

// prune prune the last TxGroup from the sorted array to be lower or equal then maximum
func (b *TxBatch) prune() error {
	b.sort()
	maxTx := int(b.selectionConfig.MaxTx)
	allL1Txs := append(b.l1UserTxs, b.l1UserFutureTxs...)
	// check if the batch length is greater than the maximum transaction
	i := len(b.txs) - 1
	for b.length() > maxTx && i >= 0 {
		last := b.txs[i]
		if last == nil {
			return fmt.Errorf("invalid TxGroup")
		}
		i--
		log.Debugw("TxSelector: batch pruning group",
			"l2Txs", len(last.l2Txs),
			"l1CoordTxs", len(last.l1CoordTxs),
			"l1UserTxs", len(last.l1UserTxs),
		)
		pruned := last.prune()
		if !pruned {
			continue
		}
		// re-create L1 transactions after the prune
		err := last.createL1Txs(b.processor, b.l2db, b.localAccountsDB, allL1Txs)
		if err != nil {
			return err
		}
		log.Debugw("TxSelector: batch group pruned",
			"l2Txs", len(last.l2Txs),
			"l1CoordTxs", len(last.l1CoordTxs),
			"l1UserTxs", len(last.l1UserTxs),
		)
	}
	log.Debugw("TxSelector: batch pruned", "txs", b.length())
	// if after all group prunes the transaction list stills with over limit, start to pop the last transactions
	for b.length() > maxTx {
		qty := b.length() - maxTx
		last := b.last()
		if last == nil {
			return fmt.Errorf("invalid TxGroup")
		}
		log.Debugw("TxSelector: popping tx from group",
			"l2Txs", len(last.l2Txs),
			"l1CoordTxs", len(last.l1CoordTxs),
			"l1UserTxs", len(last.l1UserTxs),
		)
		// if needed, pop transactions from the batch to be lower or equal to the batch transactions maximum
		popAll := last.popTx(qty)
		if popAll && len(b.txs) > 0 {
			b.txs = b.txs[:len(b.txs)-1]
			continue
		}
		// re-create L1 transactions after the prune
		err := last.createL1Txs(b.processor, b.l2db, b.localAccountsDB, allL1Txs)
		if err != nil {
			return err
		}
		log.Debugw("TxSelector: tx popped from group",
			"l2Txs", len(last.l2Txs),
			"l1CoordTxs", len(last.l1CoordTxs),
			"l1UserTxs", len(last.l1UserTxs),
		)
	}
	return nil
}

// createTxGroups create all transaction groups
func (b *TxBatch) createTxGroups(poolTxs []common.PoolL2Tx, l1UserTxs, l1UserFutureTxs []common.L1Tx) error {
	allL1Txs := append(l1UserTxs, l1UserFutureTxs...)
	b.l1UserTxs = l1UserTxs
	b.l1UserFutureTxs = l1UserFutureTxs
	b.txs = make([]*TxGroup, 0)

	txAtomicMapping, idxMapping := buildTxsMap(poolTxs)

	l1Position := len(l1UserTxs)
	// create the groups for the atomic transactions
	for idx, pool := range txAtomicMapping {
		group, err := NewTxGroup(true, pool, b.processor, b.l2db, b.localAccountsDB,
			l1Position, b.coordAccount, allL1Txs)
		if err != nil {
			return err
		}
		l1Position += group.l1Length()
		b.txs = append(b.txs, group)
		log.Debugw("TxSelector: atomic group created", "group", idx.String(), "txs", len(pool))
	}

	// create the groups for the regular transactions
	for idx, pool := range idxMapping {
		group, err := NewTxGroup(false, pool, b.processor, b.l2db, b.localAccountsDB,
			l1Position, b.coordAccount, allL1Txs)
		if err != nil {
			return err
		}
		l1Position += group.l1Length()
		b.txs = append(b.txs, group)
		log.Debugw("TxSelector: tx group created", "group", idx.String(), "txs", len(pool))
	}
	return nil
}

// buildTxsMap build the transaction map based in atomic transactions or idx sender
func buildTxsMap(poolTxs []common.PoolL2Tx) (map[common.TxID][]common.PoolL2Tx, map[common.Idx][]common.PoolL2Tx) {
	// build atomic transaction groups
	txAtomicMapping, discarded, usedTxs := buildAtomicTxs(poolTxs)

	// create the idx mapping
	idxMapping := make(map[common.Idx][]common.PoolL2Tx)
	for _, tx := range poolTxs {
		// check if the transaction has a atomic link or exist in a atomic group, or discarded
		// because the atomic request id not exist
		_, isUsed := usedTxs[tx.TxID]
		_, isDiscarded := discarded[tx.TxID]
		if isUsed || isDiscarded || tx.RqTxID != common.EmptyTxID {
			continue
		}
		pool, ok := idxMapping[tx.FromIdx]
		if !ok || pool == nil {
			idxMapping[tx.FromIdx] = []common.PoolL2Tx{tx}
		} else {
			idxMapping[tx.FromIdx] = append(idxMapping[tx.FromIdx], tx)
		}
	}
	return txAtomicMapping, idxMapping
}

// buildAtomicTxs build the atomic transactions groups and add into a mapping
func buildAtomicTxs(poolTxs []common.PoolL2Tx) (map[common.TxID][]common.PoolL2Tx, map[common.TxID]bool, map[common.TxID]common.TxID) {
	atomics := make(map[common.TxID][]common.PoolL2Tx)
	discarded := make(map[common.TxID]bool)
	owners := make(map[common.TxID]common.TxID)
	if len(poolTxs) == 0 {
		return atomics, discarded, owners
	}
	txMap := make(map[common.TxID]bool)
	for _, tx := range poolTxs {
		txMap[tx.TxID] = true
	}
	for _, tx := range poolTxs {
		// check if the tx rq tx exist
		_, ok := txMap[tx.RqTxID]
		if tx.RqTxID != common.EmptyTxID && !ok {
			discarded[tx.TxID] = true
			continue
		}

		// check if the tx already have a group owner
		rootTxID, ok := owners[tx.TxID]
		if !ok {
			rootTxID = tx.TxID
		}
		// check if the root tx already exist into the mapping
		txs, ok := atomics[rootTxID]
		if ok {
			// only add if exist
			atomics[rootTxID] = append(txs, tx)
		} else if tx.RqTxID != common.EmptyTxID {
			// if not exist, check if the nested atomic transaction exist
			auxTxID, ok := owners[tx.RqTxID]
			if ok {
				// set the nested atomic as a root and add the child
				rootTxID = auxTxID
				atomics[rootTxID] = append(atomics[rootTxID], tx)
			} else {
				// create a new atomic group if not exist
				atomics[rootTxID] = []common.PoolL2Tx{tx}
			}
		} else {
			// create a new atomic group if not exist
			atomics[rootTxID] = []common.PoolL2Tx{tx}
		}
		// add the tx to the owner mapping
		if tx.RqTxID != common.EmptyTxID {
			owners[tx.RqTxID] = rootTxID
		} else {
			owners[rootTxID] = tx.TxID
		}
	}
	// sanitize the atomic transaction removing the non-atomics
	for key, group := range atomics {
		if len(group) > 1 {
			continue
		}
		delete(atomics, key)
		delete(owners, key)
		tx := group[0]
		if tx.RqTxID != common.EmptyTxID {
			discarded[tx.TxID] = true
		}
	}
	return atomics, discarded, owners
}
