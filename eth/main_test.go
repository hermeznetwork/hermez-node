package eth

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

/*var donationAddressStr = os.Getenv("DONATION_ADDRESS")
var bootCoordinatorStr = os.Getenv("BOOT_COORDINATOR_ADDRESS")
var auctionAddressStr = os.Getenv("AUCTION_ADDRESS")
var tokenHezStr = os.Getenv("TOKEN_ADDRESS")
var hermezStr = os.Getenv("HERMEZ_ADDRESS")
var governanceAddressStr = os.Getenv("GOV_ADDRESS")
var governancePrivateKey = os.Getenv("GOV_PK")
var ethClientDialURL = os.Getenv("ETHCLIENT_DIAL_URL")*/
var ethClientDialURL = "http://localhost:8545"
var password = "pass"
var deadline, _ = new(big.Int).SetString("ffffffffffffffffffffffffffffffff", 16)

// Smart Contract Addresses
var (
	auctionAddressStr           = "0x038B86d9d8FAFdd0a02ebd1A476432877b0107C8"
	auctionAddressConst         = ethCommon.HexToAddress(auctionAddressStr)
	auctionTestAddressStr       = "0xEcc0a6dbC0bb4D51E4F84A315a9e5B0438cAD4f0"
	auctionTestAddressConst     = ethCommon.HexToAddress(auctionTestAddressStr)
	donationAddressStr          = "0x6c365935CA8710200C7595F0a72EB6023A7706Cd"
	donationAddressConst        = ethCommon.HexToAddress(donationAddressStr)
	bootCoordinatorAddressStr   = "0xc783df8a850f42e7f7e57013759c285caa701eb6"
	bootCoordinatorAddressConst = ethCommon.HexToAddress(bootCoordinatorAddressStr)
	tokenHEZAddressStr          = "0xf4e77E5Da47AC3125140c470c71cBca77B5c638c" //nolint:gosec
	tokenHEZAddressConst        = ethCommon.HexToAddress(tokenHEZAddressStr)
	hermezRollupAddressStr      = "0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe"
	hermezRollupAddressConst    = ethCommon.HexToAddress(hermezRollupAddressStr)
	wdelayerAddressStr          = "0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e"
	wdelayerAddressConst        = ethCommon.HexToAddress(wdelayerAddressStr)
	wdelayerTestAddressStr      = "0x1d80315fac6aBd3EfeEbE97dEc44461ba7556160"
	wdelayerTestAddressConst    = ethCommon.HexToAddress(wdelayerTestAddressStr)
	safetyAddressStr            = "0xE5904695748fe4A84b40b3fc79De2277660BD1D3"
	safetyAddressConst          = ethCommon.HexToAddress(safetyAddressStr)
)

// Ethereum Accounts
var (
	hermezGovernanceDAOAddressSK    = "2a8aede924268f84156a00761de73998dac7bf703408754b776ff3f873bcec60"
	hermezGovernanceDAOAddressStr   = "0x84Fae3d3Cba24A97817b2a18c2421d462dbBCe9f"
	hermezGovernanceDAOAddressConst = ethCommon.HexToAddress(hermezGovernanceDAOAddressStr)

	whiteHackGroupAddressSK    = "8b24fd94f1ce869d81a34b95351e7f97b2cd88a891d5c00abc33d0ec9501902e"
	whiteHackGroupAddressStr   = "0xfa3BdC8709226Da0dA13A4d904c8b66f16c3c8BA"
	whiteHackGroupAddressConst = ethCommon.HexToAddress(whiteHackGroupAddressStr)

	hermezKeeperAddressSK    = "7f307c41137d1ed409f0a7b028f6c7596f12734b1d289b58099b99d60a96efff"
	hermezKeeperAddressStr   = "0xFbC51a9582D031f2ceaaD3959256596C5D3a5468"
	hermezKeeperAddressConst = ethCommon.HexToAddress(hermezKeeperAddressStr)

	governanceAddressSK    = "d49743deccbccc5dc7baa8e69e5be03298da8688a15dd202e20f15d5e0e9a9fb"
	governanceAddressStr   = "0xead9c93b79ae7c1591b1fb5323bd777e86e150d4"
	governanceAddressConst = ethCommon.HexToAddress(governanceAddressStr)

	auxAddressSK    = "28d1bfbbafe9d1d4f5a11c3c16ab6bf9084de48d99fbac4058bdfa3c80b29089"
	auxAddressStr   = "0x3d91185a02774C70287F6c74Dd26d13DFB58ff16"
	auxAddressConst = ethCommon.HexToAddress(auxAddressStr)

	aux2AddressSK    = "28d1bfbbafe9d1d4f5a11c3c16ab6bf9084de48d99fbac4058bdfa3c80b29087"
	aux2AddressStr   = "0x532792b73c0c6e7565912e7039c59986f7e1dd1f"
	aux2AddressConst = ethCommon.HexToAddress(aux2AddressStr)

	hermezRollupTestSK           = "28d1bfbbafe9d1d4f5a11c3c16ab6bf9084de48d99fbac4058bdfa3c80b29088"
	hermezRollupTestAddressStr   = "0xEa960515F8b4C237730F028cBAcF0a28E7F45dE0"
	hermezRollupAddressTestConst = ethCommon.HexToAddress(hermezRollupTestAddressStr)
)

var tokenHEZ = TokenConfig{
	Address: tokenHEZAddressConst,
	Name:    "Hermez Network Token",
}

var (
	accountGov           *accounts.Account
	accountKep           *accounts.Account
	accountWhite         *accounts.Account
	accountGovDAO        *accounts.Account
	accountAux           *accounts.Account
	accountAux2          *accounts.Account
	accountHermez        *accounts.Account
	ks                   *keystore.KeyStore
	ethClient            *ethclient.Client
	ethereumClientWhite  *EthereumClient
	ethereumClientKep    *EthereumClient
	ethereumClientGovDAO *EthereumClient
	ethereumClientAux    *EthereumClient
	ethereumClientAux2   *EthereumClient
	ethereumClientHermez *EthereumClient
)

func addKey(ks *keystore.KeyStore, skHex string) *accounts.Account {
	key, err := crypto.HexToECDSA(skHex)
	if err != nil {
		panic(err)
	}
	account, err := ks.ImportECDSA(key, password)
	if err != nil {
		panic(err)
	}
	err = ks.Unlock(account, password)
	if err != nil {
		panic(err)
	}
	return &account
}

func TestMain(m *testing.M) {
	exitVal := 0

	if os.Getenv("INTEGRATION") != "" {
		dir, err := ioutil.TempDir("", "tmpks")
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				panic(err)
			}
		}()
		ks = keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

		// Load ethereum accounts from private keys
		accountGov = addKey(ks, governanceAddressSK)
		accountKep = addKey(ks, hermezKeeperAddressSK)
		accountWhite = addKey(ks, whiteHackGroupAddressSK)
		accountGovDAO = addKey(ks, hermezGovernanceDAOAddressSK)
		accountAux = addKey(ks, auxAddressSK)
		accountAux2 = addKey(ks, aux2AddressSK)
		accountHermez = addKey(ks, hermezRollupTestSK)

		ethClient, err = ethclient.Dial(ethClientDialURL)
		if err != nil {
			panic(err)
		}

		// Controllable Governance Address

		ethereumClientGov := NewEthereumClient(ethClient, accountGov, ks, nil)
		auctionClient, err = NewAuctionClient(ethereumClientGov, auctionAddressConst, tokenHEZ)
		if err != nil {
			panic(err)
		}
		auctionClientTest, err = NewAuctionClient(ethereumClientGov, auctionTestAddressConst, tokenHEZ)
		if err != nil {
			panic(err)
		}
		rollupClient, err = NewRollupClient(ethereumClientGov, hermezRollupAddressConst, tokenHEZ)
		if err != nil {
			panic(err)
		}
		wdelayerClient, err = NewWDelayerClient(ethereumClientGov, wdelayerAddressConst)
		if err != nil {
			panic(err)
		}
		wdelayerClientTest, err = NewWDelayerClient(ethereumClientGov, wdelayerTestAddressConst)
		if err != nil {
			panic(err)
		}

		ethereumClientKep = NewEthereumClient(ethClient, accountKep, ks, nil)
		ethereumClientWhite = NewEthereumClient(ethClient, accountWhite, ks, nil)
		ethereumClientGovDAO = NewEthereumClient(ethClient, accountGovDAO, ks, nil)
		ethereumClientAux = NewEthereumClient(ethClient, accountAux, ks, nil)
		ethereumClientAux2 = NewEthereumClient(ethClient, accountAux2, ks, nil)
		ethereumClientHermez = NewEthereumClient(ethClient, accountHermez, ks, nil)

		exitVal = m.Run()
	}
	os.Exit(exitVal)
}
