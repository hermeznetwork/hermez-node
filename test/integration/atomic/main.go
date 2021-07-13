package main

/*
This integration test will perform atomic transactions.

Config requirements:
- Two EthPrivKeys that have HEZ accounts with funds on Goerli
- HermezNodeURL operating on Goerli, should win the auction too, oterwise the test will run forever
*/

import (
	"math/big"
	"time"

	"github.com/hermeznetwork/hermez-go-sdk/transaction"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/integration"
)

func main() {
	// Load integration testing
	it, err := integration.NewIntegrationTest()
	if err != nil {
		log.Error(err)
		panic("Failed initializing integration testing framework")
	}
	if len(it.Wallets) < 2 {
		panic("To run this test at least two wallets with HEZ deposited must be provided.")
	}
	// Test happy path
	if err := caseHappyPath(it); err != nil {
		log.Error(err)
		panic("Failed testing case: happy path")
	}
	// Test reject and follow
	if err := caseRejectAndFollow(it); err != nil {
		log.Error(err)
		panic("Failed testing case: reject and follow")
	}
}

// caseHappyPath send a valid (nonce, balance, ...) atomic group of two linked txs
func caseHappyPath(it integration.IntegrationTest) error {
	log.Info("Start atomic test case: happy path")
	log.Info("Sending atomic group of two linked txs")
	// Define txs
	tx1 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[0],
		RecipientAddress:      it.Wallets[1].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              1, //+1
	}

	tx2 := transaction.AtomicTxItem{
		SenderBjjWallet:       it.Wallets[1],
		RecipientAddress:      it.Wallets[0].HezEthAddress,
		TokenSymbolToTransfer: "HEZ",
		Amount:                big.NewInt(100000000000000000),
		FeeRangeSelectedID:    126,
		RqOffSet:              7, //-1
	}
	txs := make([]transaction.AtomicTxItem, 2)
	txs[0] = tx1
	txs[1] = tx2

	// create PoolL2Txs
	atomicGroup := common.AtomicGroup{}
	fullTxs, err := transaction.CreateFullTxs(it.Client, txs)
	if err != nil {
		return err
	}
	atomicGroup.Txs = fullTxs

	// set AtomicGroupID
	atomicGroup = transaction.SetAtomicGroupID(atomicGroup)

	// Sign the txs
	for i := range txs {
		var txHash *big.Int
		txHash, err = atomicGroup.Txs[i].HashToSign(uint16(5))
		if err != nil {
			return err
		}
		signature := txs[i].SenderBjjWallet.PrivateKey.SignPoseidon(txHash)
		atomicGroup.Txs[i].Signature = signature.Compress()
	}

	// Post
	var serverResponse string
	serverResponse, err = transaction.SendAtomicTxsGroup(it.Client, atomicGroup)
	if err != nil {
		return err
	}
	log.Info("Txs sent successfuly: ", serverResponse)
	log.Info("You can manually check the status of the txs at: <node URL>/1/atomic-pool/")
	// Wait until txs are forged
	log.Info("Entering a wait loop until txs are forged")
	const timeBetweenChecks = 10 * time.Second
	if err := integration.WaitUntilTxsAreForged(atomicGroup.Txs, timeBetweenChecks, it.Client); err != nil {
		return err
	}
	log.Info("Atomic group has been forged")
	return nil
}

// caseRejectAndFollow this case is intended to test that atomic groups are rejected safely
// without side effects on the StateDB. To achieve that, the following txs will be sent:
func caseRejectAndFollow(integration.IntegrationTest) error {
	// TODO
	return nil
}
