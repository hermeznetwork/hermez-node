package statedb

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

func concatEthAddrBJJ(addr ethCommon.Address, pk *babyjub.PublicKey) []byte {
	pkComp := pk.Compress()
	var b []byte
	b = append(b, addr.Bytes()...)
	b = append(b[:], pkComp[:]...)
	return b
}

// setIdxByEthAddrBJJ stores the given Idx in the StateDB as follows:
// - key: Eth Address, value: idx
// - key: EthAddr & BabyJubJub PublicKey Compressed, value: idx
// If Idx already exist for the given EthAddr & BJJ, the remaining Idx will be
// always the smallest one.
func (s *StateDB) setIdxByEthAddrBJJ(idx common.Idx, addr ethCommon.Address, pk *babyjub.PublicKey) error {
	oldIdx, err := s.GetIdxByEthAddrBJJ(addr, pk)
	if err == nil {
		// EthAddr & BJJ already have an Idx
		// check which Idx is smaller
		// if new idx is smaller, store the new one
		// if new idx is bigger, don't store and return, as the used one will be the old
		if idx >= oldIdx {
			log.Debug("StateDB.setIdxByEthAddrBJJ: Idx not stored because there already exist a smaller Idx for the given EthAddr & BJJ")
			return nil
		}
	}

	// store idx for EthAddr & BJJ assuming that EthAddr & BJJ still don't
	// have an Idx stored in the DB, and if so, the already stored Idx is
	// bigger than the given one, so should be updated to the new one
	// (smaller)
	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}
	k := concatEthAddrBJJ(addr, pk)
	// store Addr&BJJ-idx
	idxBytes, err := idx.Bytes()
	if err != nil {
		return err
	}
	err = tx.Put(k, idxBytes[:])
	if err != nil {
		return err
	}
	// store Addr-idx
	err = tx.Put(addr.Bytes(), idxBytes[:])
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// GetIdxByEthAddr returns the smallest Idx in the StateDB for the given
// Ethereum Address. Will return common.Idx(0) and error in case that Idx is
// not found in the StateDB.
func (s *StateDB) GetIdxByEthAddr(addr ethCommon.Address) (common.Idx, error) {
	b, err := s.db.Get(addr.Bytes())
	if err != nil {
		return common.Idx(0), err
	}
	idx, err := common.IdxFromBytes(b)
	if err != nil {
		return common.Idx(0), err
	}
	return idx, nil
}

// GetIdxByEthAddrBJJ returns the smallest Idx in the StateDB for the given
// Ethereum Address AND the given BabyJubJub PublicKey. If `addr` is the zero
// address, it's ignored in the query.  If `pk` is nil, it's ignored in the
// query.  Will return common.Idx(0) and error in case that Idx is not found in
// the StateDB.
func (s *StateDB) GetIdxByEthAddrBJJ(addr ethCommon.Address, pk *babyjub.PublicKey) (common.Idx, error) {
	if pk == nil {
		return s.GetIdxByEthAddr(addr)
	}

	k := concatEthAddrBJJ(addr, pk)
	b, err := s.db.Get(k)
	if err != nil {
		return common.Idx(0), err
	}
	idx, err := common.IdxFromBytes(b)
	if err != nil {
		return common.Idx(0), err
	}
	return idx, nil
}

func siblingsToZKInputFormat(s []*merkletree.Hash) []*big.Int {
	b := make([]*big.Int, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[i].BigInt()
	}
	return b
}

// BJJCompressedTo256BigInts returns a [256]*big.Int array with the bit
// representation of the babyjub.PublicKeyComp
func BJJCompressedTo256BigInts(pkComp babyjub.PublicKeyComp) [256]*big.Int {
	var r [256]*big.Int
	b := pkComp[:]

	for i := 0; i < 256; i++ {
		if b[i/8]&(1<<(i%8)) == 0 {
			r[i] = big.NewInt(0)
		} else {
			r[i] = big.NewInt(1)
		}
	}

	return r
}
