// Package txsets contains Til sets of transactions & Transactions generation
// that are used at tests of other packages of hermez-node
//nolint:gomnd
package txsets

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The methods from this file are used at txprocessor package to test the
// ZKInputs generation & at tests of the test/zkproof to test the integration
// of the ZKInputs generation with the proof server

// GenerateJsUsers generates the same values than in the js test
func GenerateJsUsers(t *testing.T) []til.User {
	// same values than in the js test
	// skJsHex is equivalent to the 0000...000i js private key in commonjs
	skJsHex := []string{"7eb258e61862aae75c6c1d1f7efae5006ffc9e4d5596a6ff95f3df4ea209ea7f",
		"c005700f76f4b4cec710805c21595688648524df0a9d467afae537b7a7118819",
		"b373d14c67fb2a517bf4ac831c93341eec8e1b38dbc14e7d725b292a7cf84707",
		"2064b68d04a7aaae0ac3b36bf6f1850b380f1423be94a506c531940bd4a48b76"}
	addrHex := []string{"0x7e5f4552091a69125d5dfcb7b8c2659029395bdf",
		"0x2b5ad5c4795c026514f8317c7a215e218dccd6cf",
		"0x6813eb9362372eef6200f3b1dbc3f819671cba69",
		"0x1eff47bc3a10a45d4b230b5d10e37751fe6aa718"}
	var users []til.User
	for i := 0; i < len(skJsHex); i++ {
		skJs, err := hex.DecodeString(skJsHex[i])
		require.NoError(t, err)
		var sk babyjub.PrivateKey
		copy(sk[:], skJs)
		// bjj := sk.Public()
		user := til.User{
			Name: strconv.Itoa(i),
			BJJ:  &sk,
			Addr: ethCommon.HexToAddress(addrHex[i]),
		}
		users = append(users, user)
	}
	assert.Equal(t, "d746824f7d0ac5044a573f51b278acb56d823bec39551d1d7bf7378b68a1b021",
		users[0].BJJ.Public().String())
	assert.Equal(t, "4d05c307400c65795f02db96b1b81c60386fd53e947d9d3f749f3d99b1853909",
		users[1].BJJ.Public().String())
	assert.Equal(t, "38ffa002724562eb2a952a2503e206248962406cf16392ff32759b6f2a41fe11",
		users[2].BJJ.Public().String())
	assert.Equal(t, "c719e6401190be7fa7fbfcd3448fe2755233c01575341a3b09edadf5454f760b",
		users[3].BJJ.Public().String())

	return users
}

func signL2Tx(t *testing.T, chainID uint16, user til.User, l2Tx common.PoolL2Tx) common.PoolL2Tx {
	toSign, err := l2Tx.HashToSign(chainID)
	require.NoError(t, err)
	sig := user.BJJ.SignPoseidon(toSign)
	l2Tx.Signature = sig.Compress()
	return l2Tx
}

// GenerateTxsZKInputsHash0 generates the transactions for the TestZKInputsHash0
func GenerateTxsZKInputsHash0(t *testing.T, chainID uint16) (users []til.User,
	coordIdxs []common.Idx, l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx,
	l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])

	return users, []common.Idx{}, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputsHash1 generates the transactions for the TestZKInputsHash1
func GenerateTxsZKInputsHash1(t *testing.T, chainID uint16) (users []til.User,
	coordIdxs []common.Idx, l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx,
	l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)
	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 257,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     137,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])

	return users, []common.Idx{}, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs0 generates the transactions for the TestZKInputs0
func GenerateTxsZKInputs0(t *testing.T, chainID uint16) (users []til.User,
	coordIdxs []common.Idx, l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx,
	l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	depositAmount, err := common.Float40(10400).BigInt()
	require.Nil(t, err)
	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: depositAmount,
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     0,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])

	return users, []common.Idx{}, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs1 generates the transactions for the TestZKInputs1
func GenerateTxsZKInputs1(t *testing.T, chainID uint16) (users []til.User, coordIdxs []common.Idx,
	l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])

	coordIdxs = []common.Idx{257}
	return users, coordIdxs, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs2 generates the transactions for the TestZKInputs2
func GenerateTxsZKInputs2(t *testing.T, chainID uint16) (users []til.User, coordIdxs []common.Idx,
	l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[2].BJJ.Public().Compress(),
			FromEthAddr:   users[2].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[3].BJJ.Public().Compress(),
			FromEthAddr:   users[3].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   258,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   259,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, chainID, users[0], l2Txs[1])

	coordIdxs = []common.Idx{257}
	return users, coordIdxs, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs3 generates the transactions for the TestZKInputs3
func GenerateTxsZKInputs3(t *testing.T, chainID uint16) (users []til.User, coordIdxs []common.Idx,
	l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[2].BJJ.Public().Compress(),
			FromEthAddr:   users[2].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(1000),
			TokenID:       1,
			FromBJJ:       users[3].BJJ.Public().Compress(),
			FromEthAddr:   users[3].Addr,
			ToIdx:         258,
			Type:          common.TxTypeCreateAccountDepositTransfer,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   258,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   259,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, chainID, users[0], l2Txs[1])

	coordIdxs = []common.Idx{257}
	return users, coordIdxs, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs4 generates the transactions for the TestZKInputs4
func GenerateTxsZKInputs4(t *testing.T, chainID uint16) (users []til.User, coordIdxs []common.Idx,
	l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[2].BJJ.Public().Compress(),
			FromEthAddr:   users[2].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(1000),
			TokenID:       1,
			FromBJJ:       users[3].BJJ.Public().Compress(),
			FromEthAddr:   users[3].Addr,
			ToIdx:         258,
			Type:          common.TxTypeCreateAccountDepositTransfer,
			UserOrigin:    true,
		},
		{
			FromIdx:       258,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(1000),
			TokenID:       1,
			FromEthAddr:   users[2].Addr,
			ToIdx:         259,
			Type:          common.TxTypeDepositTransfer,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   258,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   259,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, chainID, users[0], l2Txs[1])

	coordIdxs = []common.Idx{257}
	return users, coordIdxs, l1UserTxs, []common.L1Tx{}, l2Txs
}

// GenerateTxsZKInputs5 generates the transactions for the TestZKInputs5
func GenerateTxsZKInputs5(t *testing.T, chainID uint16) (users []til.User, coordIdxs []common.Idx,
	l1UserTxs []common.L1Tx, l1CoordTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// same values than in the js test
	users = GenerateJsUsers(t)

	l1UserTxs = []common.L1Tx{
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[0].BJJ.Public().Compress(),
			FromEthAddr:   users[0].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       0,
			DepositAmount: big.NewInt(16000000),
			Amount:        big.NewInt(0),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         0,
			Type:          common.TxTypeCreateAccountDeposit,
			UserOrigin:    true,
		},
		{
			FromIdx:       257,
			DepositAmount: big.NewInt(0),
			Amount:        big.NewInt(1000),
			TokenID:       1,
			FromBJJ:       users[1].BJJ.Public().Compress(),
			FromEthAddr:   users[1].Addr,
			ToIdx:         1,
			Type:          common.TxTypeForceExit,
			UserOrigin:    true,
		},
	}
	l2Txs = []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   257,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
		{
			FromIdx: 256,
			ToIdx:   1,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   1,
			Fee:     126,
			Type:    common.TxTypeExit,
		},
	}

	l2Txs[0] = signL2Tx(t, chainID, users[0], l2Txs[0])
	l2Txs[1] = signL2Tx(t, chainID, users[0], l2Txs[1])

	coordIdxs = []common.Idx{257}
	return users, coordIdxs, l1UserTxs, []common.L1Tx{}, l2Txs
}
