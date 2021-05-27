package txselector

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

type (
	// stateDB represents the StateDB interface
	stateDB interface {
		UpdateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error)
		GetAccount(idx common.Idx) (*common.Account, error)
		GetIdxByEthAddrBJJ(addr ethCommon.Address, pk babyjub.PublicKeyComp,
			tokenID common.TokenID) (common.Idx, error)
		GetIdxByEthAddr(addr ethCommon.Address, tokenID common.TokenID) (common.Idx,
			error)
	}
	// l2DB represents the L2DB interface
	l2DB interface {
		GetAccountCreationAuth(addr ethCommon.Address) (*common.AccountCreationAuth, error)
	}
	// txProcessor represents the TxProcessor interface
	txProcessor interface {
		StateDB() *statedb.StateDB
		AccumulatedCoordFees() map[common.Idx]*big.Int
		ProcessL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx,
			*common.Account, bool, *common.Account, error)
		ProcessL2Tx(coordIdxsMap map[common.TokenID]common.Idx,
			collectedFees map[common.TokenID]*big.Int, exitTree *merkletree.MerkleTree,
			tx *common.PoolL2Tx) (*common.Idx, *common.Account, bool, error)
	}
)
