package txselector

import (
	"fmt"
	"math/big"
	"sync"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
)

type (
	stateDbMock struct {
		sync.Mutex
		accounts accounts
	}
	l2DBMock struct {
		sync.Mutex
		accounts accounts
	}
	txProcessorMock struct {
		db *stateDbMock
	}
	account struct {
		idx       common.Idx
		tokenID   common.TokenID
		bjj       babyjub.PublicKeyComp
		ethAddr   ethCommon.Address
		balance   *big.Int
		nonce     common.Nonce
		signature []byte
	}
	accounts map[ethCommon.Address]account
)

var (
	_invalidEthAddr1 = ethCommon.HexToAddress("0x1e5760601923edd36f7b1c9bf5f3c36305c475a4")
	_ethAddr1        = ethCommon.HexToAddress("0x9aC7Fdc4930e7798f9a4e014AAc0544e19b8AcE0")
	// hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b
	_bjj1 = babyjub.PublicKeyComp{174, 75, 245, 119, 82, 189, 63, 219, 13, 91, 208, 49, 110, 119, 88, 47,
		180, 237, 130, 218, 171, 162, 206, 17, 194, 11, 112, 245, 236, 9, 249, 47}
	_ethAddr2 = ethCommon.HexToAddress("0xd9391B20559777E1b94954Ed84c28541E35bFEb8")
	// hez:rkv1d1K9P9sNW9AxbndYL7Ttgtqros4Rwgtw9ewJ-S_b
	_bjj2 = babyjub.PublicKeyComp{174, 75, 245, 119, 82, 189, 63, 219, 13, 91, 208, 49, 110, 119, 88, 47,
		180, 237, 130, 218, 171, 162, 206, 17, 194, 11, 112, 245, 236, 9, 249, 47}
	_ethAddr3 = ethCommon.HexToAddress("0x0186bDCc193c657fA790503A721709033686FdAA")
	// hez:CtNBupmBIq1MUs64LbATeQhAP6fA6wDXoRkvhRcwPYt7
	_bjj3 = babyjub.PublicKeyComp{10, 211, 65, 186, 153, 129, 34, 173, 76, 82, 206, 184, 45, 176, 19,
		121, 8, 64, 63, 167, 192, 235, 0, 215, 161, 25, 47, 133, 23, 48, 61, 139}
	_ethAddr4 = ethCommon.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf")
	// hez:jedt7Ort5eBN0nAsRvrDmNK068XiloHuGgc3eTUYyqZq
	_bjj4 = babyjub.PublicKeyComp{166, 202, 24, 53, 121, 55, 7, 26, 238, 129, 150, 226, 197, 235, 180,
		210, 152, 195, 250, 70, 44, 112, 210, 77, 224, 229, 237, 234, 236, 109, 231, 141}
	_accounts = accounts{
		common.FFAddr: {
			idx:       349,
			ethAddr:   _ethAddr1,
			nonce:     1,
			balance:   big.NewInt(1000),
			bjj:       _bjj1,
			signature: []byte("0x6dae7cfbb6fcc580ac6f2eae2eaf2dca91d55a4139e67865249da546458d17674fde9a480fc29584384bab1e0daf5ee1ec454ae22fa2640460fc36b9c3a969871b"),
		},
		_ethAddr2: {
			idx:       350,
			ethAddr:   _ethAddr2,
			nonce:     1,
			balance:   big.NewInt(1000),
			bjj:       _bjj2,
			signature: []byte("0x6dae7cfbb6fcc580ac6f2eae2eaf2dca91d55a4139e67865249da546458d17674fde9a480fc29584384bab1e0daf5ee1ec454ae22fa2640460fc36b9c3a969871b"),
		},
		_ethAddr3: {
			idx:       351,
			ethAddr:   _ethAddr3,
			nonce:     10,
			balance:   big.NewInt(10000),
			bjj:       _bjj3,
			signature: []byte("0x58384fddde81bcde8a8f1951a101f06b5ff2f2de61318e421df548d93a8a0c1a3092154f4e8a2bceaabb6039380d46d74cabb5390c699a5ed864d80e2071d9911c"),
		},
		_ethAddr4: {
			idx:       352,
			ethAddr:   _ethAddr4,
			nonce:     33,
			balance:   big.NewInt(1000000000000),
			bjj:       _bjj4,
			signature: []byte("0x58384fddde81bcde8a8f1951a101f06b5ff2f2de61318e421df548d93a8a0c1a3092154f4e8a2bceaabb6039380d46d74cabb5390c699a5ed864d80e2071d9911c"),
		},
	}
	_coordAccount = CoordAccount{
		Addr:                _ethAddr2,
		BJJ:                 _bjj2,
		AccountCreationAuth: []byte("0x58384fddde81bcde8a8f1951a101f06b5ff2f2de61318e421df548d93a8a0c1a3092154f4e8a2bceaabb6039380d46d74cabb5390c699a5ed864d80e2071d9911c"),
	}
	_config = txprocessor.Config{
		NLevels: 32,
		MaxTx:   5,
		MaxL1Tx: 256,
		ChainID: 1,
	}
)

var (
	txID1 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	txID2 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	txID3 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
	txID4 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}
	txID5 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5}
	txID6 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6}
	txID7 = common.TxID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7}
)

func createStateDbMock() *stateDbMock {
	accs := make(accounts)
	for key, value := range _accounts {
		accs[key] = value
	}
	return &stateDbMock{accounts: accs}
}

func createL2DBMock() *l2DBMock {
	accs := make(accounts)
	for key, value := range _accounts {
		accs[key] = value
	}
	return &l2DBMock{accounts: accs}
}

func createTxProcessorMock() *txProcessorMock {
	return &txProcessorMock{}
}

// UpdateAccount stateDb mock method
func (st *stateDbMock) UpdateAccount(idx common.Idx, newAcc *common.Account) (*merkletree.CircomProcessorProof, error) {
	st.Lock()
	defer st.Unlock()
	for _, acc := range st.accounts {
		if acc.idx == idx {
			acc.balance = newAcc.Balance
			acc.nonce = newAcc.Nonce
			st.accounts[acc.ethAddr] = acc
			return nil, nil
		}
	}
	return nil, tracerr.Wrap(statedb.ErrIdxNotFound)
}

// GetAccount stateDb mock method
func (st *stateDbMock) GetAccount(idx common.Idx) (*common.Account, error) {
	st.Lock()
	defer st.Unlock()
	for _, acc := range st.accounts {
		if acc.idx == idx {
			return &common.Account{
				Idx:     acc.idx,
				TokenID: acc.tokenID,
				BJJ:     acc.bjj,
				EthAddr: acc.ethAddr,
				Nonce:   acc.nonce,
				Balance: acc.balance,
			}, nil
		}
	}
	return nil, tracerr.Wrap(db.ErrNotFound)
}

// GetIdxByEthAddrBJJ stateDb mock method
func (st *stateDbMock) GetIdxByEthAddrBJJ(addr ethCommon.Address, pk babyjub.PublicKeyComp,
	tokenID common.TokenID) (common.Idx, error) {
	st.Lock()
	defer st.Unlock()
	if tokenID == 33 || (tokenID == 34 && addr == common.FFAddr) || (tokenID == 35 && addr == _ethAddr2) {
		return 0, tracerr.Wrap(statedb.ErrIdxNotFound)
	}
	acc, ok := st.accounts[addr]
	if !ok {
		return 0, tracerr.Wrap(fmt.Errorf("GetIdxByEthAddrBJJ: %s: ToEthAddr: %s, ToBJJ: %s, TokenID: %d",
			statedb.ErrIdxNotFound, addr.Hex(), pk, tokenID))
	}
	return acc.idx, nil
}

// GetIdxByEthAddr stateDb mock method
func (st *stateDbMock) GetIdxByEthAddr(addr ethCommon.Address, tokenID common.TokenID) (common.Idx,
	error) {
	st.Lock()
	defer st.Unlock()
	acc, ok := st.accounts[addr]
	if !ok {
		return 0, tracerr.Wrap(fmt.Errorf("GetIdxByEthAddr: %s: ToEthAddr: %s, TokenID: %d",
			statedb.ErrIdxNotFound, addr.Hex(), tokenID))
	}
	return acc.idx, nil
}

// GetAccountCreationAuth l2DB mock method
func (l2 *l2DBMock) GetAccountCreationAuth(addr ethCommon.Address) (*common.AccountCreationAuth, error) {
	l2.Lock()
	defer l2.Unlock()
	if addr == _invalidEthAddr1 {
		return &common.AccountCreationAuth{
			EthAddr:   _invalidEthAddr1,
			BJJ:       _bjj1,
			Signature: []byte("0x6dae7cfbb6fcc580ac6f2eae2eaf2dca91d55a4139e67865249da546458d17674fde9a480fc29584384bab1e0daf5ee1ec454ae22fa2640460fc36b9c3a969871b"),
		}, nil
	}
	acc, ok := l2.accounts[addr]
	if !ok {
		return nil, fmt.Errorf("sql: no rows in result set")
	}
	return &common.AccountCreationAuth{
		EthAddr:   acc.ethAddr,
		BJJ:       acc.bjj,
		Signature: acc.signature,
	}, nil
}

// ProcessL1Tx txProcessor mock method
//nolint:unused
func (tp *txProcessorMock) ProcessL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx,
	*common.Account, bool, *common.Account, error) {
	return nil, nil, false, nil, nil
}

// ProcessL2Tx txProcessor mock method
//nolint:unused
func (tp *txProcessorMock) ProcessL2Tx(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int, exitTree *merkletree.MerkleTree,
	tx *common.PoolL2Tx) (*common.Idx, *common.Account, bool, error) {
	if tp.db == nil {
		return nil, nil, false, nil
	}
	tp.db.Lock()
	defer tp.db.Unlock()
	for _, acc := range tp.db.accounts {
		if acc.idx == tx.FromIdx {
			fee, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
			if err != nil {
				return nil, nil, false, err
			}
			feeAndAmount := new(big.Int).Add(tx.Amount, fee)
			acc.balance = new(big.Int).Sub(acc.balance, feeAndAmount)
			acc.nonce++
			tp.db.accounts[acc.ethAddr] = acc
		}
		if acc.idx == tx.ToIdx {
			acc.balance = new(big.Int).Add(acc.balance, tx.Amount)
			tp.db.accounts[acc.ethAddr] = acc
		}
	}
	return nil, nil, false, nil
}

// StateDB txProcessor mock method
func (tp *txProcessorMock) StateDB() *statedb.StateDB {
	return &statedb.StateDB{}
}

// AccumulatedCoordFees txProcessor mock method
func (tp *txProcessorMock) AccumulatedCoordFees() map[common.Idx]*big.Int {
	return map[common.Idx]*big.Int{}
}
