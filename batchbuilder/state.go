package batchbuilder

import (
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree/db"
)

// TODO next iteration move the methods of this file into StateDB, which Synchronizer will use in the disk DB, and BatchBuilder will use with the MemoryDB

// GetBalance returns the balance for a given Idx from the DB
func (bb *BatchBuilder) GetBalance(tx db.Tx, idx common.Idx) (*common.Leaf, error) {
	idxBytes := idx.Bytes()
	vBytes, err := tx.Get(idxBytes[:])
	if err != nil {
		return nil, err
	}
	var b [32 * common.NLEAFELEMS]byte
	copy(b[:], vBytes)
	leaf, err := common.LeafFromBytes(b)
	if err != nil {
		return nil, err
	}
	return leaf, nil
}

// CreateBalance stores the Leaf into the Idx position in the MerkleTree, also adds db entry for the Leaf value
func (bb *BatchBuilder) CreateBalance(tx db.Tx, idx common.Idx, leaf common.Leaf) error {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := leaf.HashValue()
	if err != nil {
		return err
	}
	leafBytes, err := leaf.Bytes()
	if err != nil {
		return err
	}

	// store the Leaf value
	tx.Put(v.Bytes(), leafBytes[:])
	// Add k & v into the MT
	err = bb.mt.Add(idx.BigInt(), v)
	if err != nil {
		return err
	}

	return nil
}

// UpdateBalance updates the balance of the leaf of a given Idx.
// If sending==true: will substract the amount, if sending==false will add the ammount
func (bb *BatchBuilder) UpdateBalance(tx db.Tx, idx common.Idx, amount *big.Int, sending bool) error {
	leaf, err := bb.GetBalance(tx, idx)
	if err != nil {
		return err
	}

	// TODO add checks that the numbers are correct and there is no missing value neither impossible values
	if sending {
		leaf.Balance = new(big.Int).Sub(leaf.Balance, amount)
	} else {
		leaf.Balance = new(big.Int).Add(leaf.Balance, amount)
	}

	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := leaf.HashValue()
	if err != nil {
		return err
	}
	leafBytes, err := leaf.Bytes()
	if err != nil {
		return err
	}

	// store the Leaf value
	tx.Put(v.Bytes(), leafBytes[:])
	// Add k & v into the MT
	err = bb.mt.Update(idx.BigInt(), v)
	if err != nil {
		return err
	}

	return nil
}
