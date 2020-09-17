package test

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// WARNING: the generators in this file doesn't necessary follow the protocol
// they are intended to check that the parsers between struct <==> DB are correct

// GenBlocks generates block from, to block numbers. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenBlocks(from, to int64) []common.Block {
	var blocks []common.Block
	for i := from; i < to; i++ {
		blocks = append(blocks, common.Block{
			EthBlockNum: i,
			//nolint:gomnd
			Timestamp: time.Now().Add(time.Second * 13).UTC(),
			Hash:      ethCommon.BigToHash(big.NewInt(int64(i))),
		})
	}
	return blocks
}

// GenTokens generates tokens. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenTokens(nTokens int, blocks []common.Block) []common.Token {
	tokens := []common.Token{}
	for i := 0; i < nTokens; i++ {
		token := common.Token{
			TokenID:     common.TokenID(i),
			Name:        fmt.Sprint(i),
			Symbol:      fmt.Sprint(i),
			Decimals:    uint64(i),
			EthBlockNum: blocks[i%len(blocks)].EthBlockNum,
			EthAddr:     ethCommon.BigToAddress(big.NewInt(int64(i))),
		}
		if i%2 == 0 {
			token.USD = 3
			token.USDUpdate = time.Now()
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// GenBatches generates batches. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenBatches(nBatches int, blocks []common.Block) []common.Batch {
	batches := []common.Batch{}
	collectedFees := make(map[common.TokenID]*big.Int)
	for i := 0; i < 64; i++ {
		collectedFees[common.TokenID(i)] = big.NewInt(int64(i))
	}
	for i := 0; i < nBatches; i++ {
		batch := common.Batch{
			BatchNum:    common.BatchNum(i + 1),
			EthBlockNum: blocks[i%len(blocks)].EthBlockNum,
			//nolint:gomnd
			ForgerAddr:    ethCommon.BigToAddress(big.NewInt(6886723)),
			CollectedFees: collectedFees,
			StateRoot:     common.Hash([]byte("duhdqlwiucgwqeiu")),
			//nolint:gomnd
			NumAccounts: 30,
			ExitRoot:    common.Hash([]byte("tykertheuhtgenuer3iuw3b")),
			SlotNum:     common.SlotNum(i),
		}
		if i%2 == 0 {
			batch.ForgeL1TxsNum = uint32(i)
		}
		batches = append(batches, batch)
	}
	return batches
}

// GenAccounts generates accounts. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenAccounts(totalAccounts, userAccounts int, tokens []common.Token, userAddr *ethCommon.Address, batches []common.Batch) []common.Account {
	if totalAccounts < userAccounts {
		panic("totalAccounts must be greater than userAccounts")
	}
	privK := babyjub.NewRandPrivKey()
	pubK := privK.Public()
	accs := []common.Account{}
	for i := 0; i < totalAccounts; i++ {
		var addr ethCommon.Address
		if i < userAccounts {
			addr = *userAddr
		} else {
			addr = ethCommon.BigToAddress(big.NewInt(int64(i)))
		}
		accs = append(accs, common.Account{
			Idx:       common.Idx(i),
			TokenID:   tokens[i%len(tokens)].TokenID,
			EthAddr:   addr,
			BatchNum:  batches[i%len(batches)].BatchNum,
			PublicKey: pubK,
		})
	}
	return accs
}

// GenL1Txs generates L1 txs. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenL1Txs(
	fromIdx int,
	totalTxs, nUserTxs int,
	userAddr *ethCommon.Address,
	accounts []common.Account,
	tokens []common.Token,
	blocks []common.Block,
	batches []common.Batch,
) ([]common.L1Tx, []common.L1Tx) {
	if totalTxs < nUserTxs {
		panic("totalTxs must be greater than userTxs")
	}
	userTxs := []common.L1Tx{}
	othersTxs := []common.L1Tx{}
	for i := 0; i < totalTxs; i++ {
		var tx common.L1Tx
		if batches[i%len(batches)].ForgeL1TxsNum != 0 {
			tx = common.L1Tx{
				TxID:            common.TxID(common.Hash([]byte("L1_" + strconv.Itoa(fromIdx+i)))),
				ToForgeL1TxsNum: batches[i%len(batches)].ForgeL1TxsNum,
				Position:        i,
				UserOrigin:      i%2 == 0,
				TokenID:         tokens[i%len(tokens)].TokenID,
				Amount:          big.NewInt(int64(i + 1)),
				LoadAmount:      big.NewInt(int64(i + 1)),
				EthBlockNum:     blocks[i%len(blocks)].EthBlockNum,
				Type:            randomTxType(i),
			}
			if i%4 == 0 {
				tx.BatchNum = batches[i%len(batches)].BatchNum
			}
		} else {
			continue
		}
		if i < nUserTxs {
			var from, to common.Account
			var err error
			if i%2 == 0 {
				from, err = randomAccount(i, true, userAddr, accounts)
				if err != nil {
					panic(err)
				}
				to, err = randomAccount(i, false, userAddr, accounts)
				if err != nil {
					panic(err)
				}
			} else {
				from, err = randomAccount(i, false, userAddr, accounts)
				if err != nil {
					panic(err)
				}
				to, err = randomAccount(i, true, userAddr, accounts)
				if err != nil {
					panic(err)
				}
			}
			tx.FromIdx = from.Idx
			tx.FromEthAddr = from.EthAddr
			tx.FromBJJ = from.PublicKey
			tx.ToIdx = to.Idx
			userTxs = append(userTxs, tx)
		} else {
			from, err := randomAccount(i, false, userAddr, accounts)
			if err != nil {
				panic(err)
			}
			to, err := randomAccount(i, false, userAddr, accounts)
			if err != nil {
				panic(err)
			}
			tx.FromIdx = from.Idx
			tx.FromEthAddr = from.EthAddr
			tx.FromBJJ = from.PublicKey
			tx.ToIdx = to.Idx
			othersTxs = append(othersTxs, tx)
		}
	}
	return userTxs, othersTxs
}

// GenL2Txs generates L2 txs. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenL2Txs(
	fromIdx int,
	totalTxs, nUserTxs int,
	userAddr *ethCommon.Address,
	accounts []common.Account,
	tokens []common.Token,
	blocks []common.Block,
	batches []common.Batch,
) ([]common.L2Tx, []common.L2Tx) {
	if totalTxs < nUserTxs {
		panic("totalTxs must be greater than userTxs")
	}
	userTxs := []common.L2Tx{}
	othersTxs := []common.L2Tx{}
	for i := 0; i < totalTxs; i++ {
		tx := common.L2Tx{
			TxID:     common.TxID(common.Hash([]byte("L2_" + strconv.Itoa(fromIdx+i)))),
			BatchNum: batches[i%len(batches)].BatchNum,
			Position: i,
			//nolint:gomnd
			Amount: big.NewInt(int64(i + 1)),
			//nolint:gomnd
			Fee:         common.FeeSelector(i % 256),
			Nonce:       common.Nonce(i + 1),
			EthBlockNum: blocks[i%len(blocks)].EthBlockNum,
			Type:        randomTxType(i),
		}
		if i < nUserTxs {
			var from, to common.Account
			var err error
			if i%2 == 0 {
				from, err = randomAccount(i, true, userAddr, accounts)
				if err != nil {
					panic(err)
				}
				to, err = randomAccount(i, false, userAddr, accounts)
				if err != nil {
					panic(err)
				}
			} else {
				from, err = randomAccount(i, false, userAddr, accounts)
				if err != nil {
					panic(err)
				}
				to, err = randomAccount(i, true, userAddr, accounts)
				if err != nil {
					panic(err)
				}
			}
			tx.FromIdx = from.Idx
			tx.ToIdx = to.Idx
			userTxs = append(userTxs, tx)
		} else {
			from, err := randomAccount(i, false, userAddr, accounts)
			if err != nil {
				panic(err)
			}
			to, err := randomAccount(i, false, userAddr, accounts)
			if err != nil {
				panic(err)
			}
			tx.FromIdx = from.Idx
			tx.ToIdx = to.Idx
			othersTxs = append(othersTxs, tx)
		}
	}
	return userTxs, othersTxs
}

// GenCoordinators generates coordinators. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenCoordinators(nCoords int, blocks []common.Block) []common.Coordinator {
	coords := []common.Coordinator{}
	for i := 0; i < nCoords; i++ {
		coords = append(coords, common.Coordinator{
			EthBlockNum: blocks[i%len(blocks)].EthBlockNum,
			Forger:      ethCommon.BigToAddress(big.NewInt(int64(i))),
			Withdraw:    ethCommon.BigToAddress(big.NewInt(int64(i))),
			URL:         "https://foo.bar",
		})
	}
	return coords
}

// GenBids generates bids. WARNING: This is meant for DB/API testing, and may not be fully consistent with the protocol.
func GenBids(nBids int, blocks []common.Block, coords []common.Coordinator) []common.Bid {
	bids := []common.Bid{}
	for i := 0; i < nBids; i++ {
		bids = append(bids, common.Bid{
			SlotNum:     common.SlotNum(i),
			BidValue:    big.NewInt(int64(i)),
			EthBlockNum: blocks[i%len(blocks)].EthBlockNum,
			ForgerAddr:  coords[i%len(blocks)].Forger,
		})
	}
	return bids
}

func randomAccount(seed int, userAccount bool, userAddr *ethCommon.Address, accs []common.Account) (common.Account, error) {
	i := seed % len(accs)
	firstI := i
	for {
		acc := accs[i]
		if userAccount && *userAddr == acc.EthAddr {
			return acc, nil
		}
		if !userAccount && (userAddr == nil || *userAddr != acc.EthAddr) {
			return acc, nil
		}
		i++
		i = i % len(accs)
		if i == firstI {
			return acc, errors.New("Didnt found any account matchinng the criteria")
		}
	}
}

func randomTxType(seed int) common.TxType {
	//nolint:gomnd
	switch seed % 11 {
	case 0:
		return common.TxTypeExit
	//nolint:gomnd
	case 1:
		return common.TxTypeWithdrawn
	//nolint:gomnd
	case 2:
		return common.TxTypeTransfer
	//nolint:gomnd
	case 3:
		return common.TxTypeDeposit
	//nolint:gomnd
	case 4:
		return common.TxTypeCreateAccountDeposit
	//nolint:gomnd
	case 5:
		return common.TxTypeCreateAccountDepositTransfer
	//nolint:gomnd
	case 6:
		return common.TxTypeDepositTransfer
	//nolint:gomnd
	case 7:
		return common.TxTypeForceTransfer
	//nolint:gomnd
	case 8:
		return common.TxTypeForceExit
	//nolint:gomnd
	case 9:
		return common.TxTypeTransferToEthAddr
	//nolint:gomnd
	case 10:
		return common.TxTypeTransferToBJJ
	default:
		return common.TxTypeTransfer
	}
}
