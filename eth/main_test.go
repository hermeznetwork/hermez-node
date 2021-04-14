package eth

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var errEnvVar = fmt.Errorf("Some environment variable is missing")

var (
	ethClientDialURL = "http://localhost:8545"
	password         = "pass"
	deadline, _      = new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)
	mnemonic         = "explain tackle mirror kit van hammer degree position ginger unfair soup bonus"
)

func genAcc(w *hdwallet.Wallet, ks *keystore.KeyStore, i int) (*accounts.Account,
	ethCommon.Address) {
	path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", i))
	account, err := w.Derive(path, false)
	if err != nil {
		log.Fatal(err)
	}

	key, err := w.PrivateKey(account)
	if err != nil {
		log.Fatal(err)
	}
	_, err = ks.ImportECDSA(key, password)
	if err != nil {
		log.Fatal(err)
	}
	if err := ks.Unlock(account, password); err != nil {
		log.Fatal(err)
	}

	return &account, account.Address
}

// Smart Contract Addresses
var (
	genesisBlock             int64
	auctionAddressConst      ethCommon.Address
	auctionTestAddressConst  ethCommon.Address
	tokenHEZAddressConst     ethCommon.Address
	hermezRollupAddressConst ethCommon.Address
	wdelayerAddressConst     ethCommon.Address
	wdelayerTestAddressConst ethCommon.Address
	tokenHEZ                 ethCommon.Address

	donationAccount      *accounts.Account
	donationAddressConst ethCommon.Address

	bootCoordinatorAccount      *accounts.Account
	bootCoordinatorAddressConst ethCommon.Address
)

// Ethereum Accounts
var (
	emergencyCouncilAccount      *accounts.Account
	emergencyCouncilAddressConst ethCommon.Address

	governanceAccount      *accounts.Account
	governanceAddressConst ethCommon.Address

	auxAccount      *accounts.Account
	auxAddressConst ethCommon.Address

	aux2Account      *accounts.Account
	aux2AddressConst ethCommon.Address

	hermezRollupTestAccount      *accounts.Account
	hermezRollupTestAddressConst ethCommon.Address
)

var (
	ks                             *keystore.KeyStore
	ethClient                      *ethclient.Client
	ethereumClientEmergencyCouncil *EthereumClient
	ethereumClientAux              *EthereumClient
	ethereumClientAux2             *EthereumClient
	ethereumClientHermez           *EthereumClient
)

func getEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Variables loaded from environment")
	} else {
		fmt.Println("Variables loaded from .env file")
	}
	var auctionAddressStr = os.Getenv("AUCTION")
	var auctionTestAddressStr = os.Getenv("AUCTION_TEST")
	var tokenHEZAddressStr = os.Getenv("TOKENHEZ")
	var hermezRollupAddressStr = os.Getenv("HERMEZ")
	var wdelayerAddressStr = os.Getenv("WDELAYER")
	var wdelayerTestAddressStr = os.Getenv("WDELAYER_TEST")
	genesisBlockEnv := os.Getenv("GENESIS_BLOCK")
	genesisBlock, err = strconv.ParseInt(genesisBlockEnv, 10, 64)
	if err != nil {
		log.Fatal(errEnvVar)
	}
	if auctionAddressStr == "" || auctionTestAddressStr == "" || tokenHEZAddressStr == "" ||
		hermezRollupAddressStr == "" || wdelayerAddressStr == "" || wdelayerTestAddressStr == "" ||
		genesisBlockEnv == "" {
		log.Fatal(errEnvVar)
	}

	auctionAddressConst = ethCommon.HexToAddress(auctionAddressStr)
	auctionTestAddressConst = ethCommon.HexToAddress(auctionTestAddressStr)
	tokenHEZAddressConst = ethCommon.HexToAddress(tokenHEZAddressStr)
	hermezRollupAddressConst = ethCommon.HexToAddress(hermezRollupAddressStr)
	wdelayerAddressConst = ethCommon.HexToAddress(wdelayerAddressStr)
	wdelayerTestAddressConst = ethCommon.HexToAddress(wdelayerTestAddressStr)
	tokenHEZ = tokenHEZAddressConst
}

func TestMain(m *testing.M) {
	exitVal := 0

	if os.Getenv("INTEGRATION") != "" {
		getEnvVariables()
		dir, err := ioutil.TempDir("", "tmpks")
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				log.Fatal(err)
			}
		}()
		ks = keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

		w, err := hdwallet.NewFromMnemonic(mnemonic)
		if err != nil {
			log.Fatal(err)
		}

		// Create ethereum accounts from mnemonic and load private keys
		// into the keystore
		bootCoordinatorAccount, bootCoordinatorAddressConst = genAcc(w, ks, 0)
		governanceAccount, governanceAddressConst = genAcc(w, ks, 1)
		emergencyCouncilAccount, emergencyCouncilAddressConst = genAcc(w, ks, 2)
		donationAccount, donationAddressConst = genAcc(w, ks, 3)
		hermezRollupTestAccount, hermezRollupTestAddressConst = genAcc(w, ks, 4)
		auxAccount, auxAddressConst = genAcc(w, ks, 5)
		aux2Account, aux2AddressConst = genAcc(w, ks, 6)

		ethClient, err = ethclient.Dial(ethClientDialURL)
		if err != nil {
			log.Fatal(err)
		}

		// Controllable Governance Address
		ethereumClientGov, err := NewEthereumClient(ethClient, governanceAccount, ks, nil)
		if err != nil {
			log.Fatal(err)
		}
		auctionClient, err = NewAuctionClient(ethereumClientGov, auctionAddressConst, tokenHEZ)
		if err != nil {
			log.Fatal(err)
		}
		auctionClientTest, err = NewAuctionClient(ethereumClientGov, auctionTestAddressConst, tokenHEZ)
		if err != nil {
			log.Fatal(err)
		}
		rollupClient, err = NewRollupClient(ethereumClientGov, hermezRollupAddressConst)
		if err != nil {
			log.Fatal(err)
		}
		wdelayerClient, err = NewWDelayerClient(ethereumClientGov, wdelayerAddressConst)
		if err != nil {
			log.Fatal(err)
		}
		wdelayerClientTest, err = NewWDelayerClient(ethereumClientGov, wdelayerTestAddressConst)
		if err != nil {
			log.Fatal(err)
		}

		ethereumClientEmergencyCouncil, err = NewEthereumClient(ethClient,
			emergencyCouncilAccount, ks, nil)
		if err != nil {
			log.Fatal(err)
		}
		ethereumClientAux, err = NewEthereumClient(ethClient, auxAccount, ks, nil)
		if err != nil {
			log.Fatal(err)
		}
		ethereumClientAux2, err = NewEthereumClient(ethClient, aux2Account, ks, nil)
		if err != nil {
			log.Fatal(err)
		}
		ethereumClientHermez, err = NewEthereumClient(ethClient, hermezRollupTestAccount, ks, nil)
		if err != nil {
			log.Fatal(err)
		}

		exitVal = m.Run()
	}
	os.Exit(exitVal)
}
