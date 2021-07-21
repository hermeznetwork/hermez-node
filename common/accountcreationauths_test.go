package common

import (
	"encoding/hex"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountCreationAuthSignVerify(t *testing.T) {
	// Ethereum key
	ethSk, err :=
		ethCrypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.NoError(t, err)
	ethAddr := ethCrypto.PubkeyToAddress(ethSk.PublicKey)

	// BabyJubJub key
	var sk babyjub.PrivateKey
	_, err = hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	require.NoError(t, err)

	chainID := uint16(0)
	hermezContractAddr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	a := AccountCreationAuth{
		EthAddr: ethAddr,
		BJJ:     sk.Public().Compress(),
	}

	// Sign using the Sign function (stores signature in a.Signature)
	err = a.Sign(func(hash []byte) ([]byte, error) {
		return ethCrypto.Sign(hash, ethSk)
	}, chainID, hermezContractAddr)
	require.NoError(t, err)

	// Hash and sign manually and compare the generated signature
	hash, err := a.HashToSign(chainID, hermezContractAddr)
	require.NoError(t, err)
	assert.Equal(t, "9414667457e658dd31949b82996b75c65a055512244c3bbfd22ff56add02ba65",
		hex.EncodeToString(hash))
	sig, err := ethCrypto.Sign(hash, ethSk)
	require.NoError(t, err)
	sig[64] += 27
	assert.Equal(t, sig, a.Signature)
	isValid, err := a.VerifySignature(chainID, hermezContractAddr)
	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestKeccak256JSComp(t *testing.T) {
	// check keccak256 compatible with js version
	h := ethCrypto.Keccak256([]byte("test"))
	assert.Equal(t, "9c22ff5f21f0b81b113e63f7db6da94fedef11b2119b4088b89664fb9a3cb658",
		hex.EncodeToString(h))
}

func TestAccountCreationAuthJSComp(t *testing.T) {
	// The values of this test have been tested with the js implementation
	type testVector struct {
		ethSk              string
		expectedAddress    string
		pkCompStr          string
		chainID            uint16
		hermezContractAddr string
		toHashExpected     string
		hashExpected       string
		sigExpected        string
	}
	var tvs []testVector
	//nolint:lll
	tv0 := testVector{
		ethSk:              "0000000000000000000000000000000000000000000000000000000000000001",
		expectedAddress:    "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf",
		pkCompStr:          "21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7",
		chainID:            uint16(4),
		hermezContractAddr: "0x7e5f4552091a69125d5dfcb7b8c2659029395bdf",
		toHashExpected:     "190189658bba487e11c7da602676ee32bc90b77d3f32a305b147e4f3c3b35f19672e5d84ccc38d0ab245c469b719549d837113465c2abf9972c49403ca6fd10ed3dc",
		hashExpected:       "c56eba41e511df100c804c5c09288f35887efea4f033be956481af335df3bea2",
		sigExpected:        "dbedcc5ce02db8f48afbdb2feba9a3a31848eaa8fca5f312ce37b01db45d2199208335330d4445bd2f51d1db68dbc0d0bf3585c4a07504b4efbe46a69eaae5a21b",
	}
	//nolint:lll
	tv1 := testVector{
		ethSk:              "0000000000000000000000000000000000000000000000000000000000000002",
		expectedAddress:    "0x2B5AD5c4795c026514f8317c7a215E218DcCD6cF",
		pkCompStr:          "093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d",
		chainID:            uint16(0),
		hermezContractAddr: "0x2b5ad5c4795c026514f8317c7a215e218dccd6cf",
		toHashExpected:     "1901dafbc253dedf90d6421dc6e25d5d9efc6985133cb2a8d363d0a081a0e3eddddc65f603a88de36aaeabd3b4cf586538c7f3fd50c94780530a3707c8c14ad9fd11",
		hashExpected:       "deb9afa479282cf27b442ce8ba86b19448aa87eacef691521a33db5d0feb9959",
		sigExpected:        "6a0da90ba2d2b1be679a28ebe54ee03082d44b836087391cd7d2607c1e4dafe04476e6e88dccb8707c68312512f16c947524b35c80f26c642d23953e9bb84c701c",
	}
	//nolint:lll
	tv2 := testVector{
		ethSk:              "c5e8f61d1ab959b397eecc0a37a6517b8e67a0e7cf1f4bce5591f3ed80199122",
		expectedAddress:    "0xc783df8a850f42e7F7e57013759C285caa701eB6",
		pkCompStr:          "22870c1bcc451396202d62f566026eab8e438c6c91decf8ddf63a6c162619b52",
		chainID:            uint16(31337), // =0x7a69
		hermezContractAddr: "0xf4e77E5Da47AC3125140c470c71cBca77B5c638c",
		toHashExpected:     "190167617949b934d7e01add4009cd3d47415a26727b7d6288e5dce33fb3721d5a1a9ce511b19b694c9aaf8183f4987ed752f24884c54c003d11daa2e98c7547a79e",
		hashExpected:       "157b570c597e615b8356ce008ac39f43bc9b6d50080bc07d968031b9378acbbb",
		sigExpected:        "a0766181102428b5672e523dc4b905c10ddf025c10dbd0b3534ef864632a14652737610041c670b302fc7dca28edd5d6eac42b72d69ce58da8ce21287b244e381b",
	}
	tvs = append(tvs, tv0)
	tvs = append(tvs, tv1)
	tvs = append(tvs, tv2)

	for _, tv := range tvs {
		// Ethereum key
		ethSk, err := ethCrypto.HexToECDSA(tv.ethSk)
		require.NoError(t, err)
		ethAddr := ethCrypto.PubkeyToAddress(ethSk.PublicKey)
		assert.Equal(t, tv.expectedAddress, ethAddr.Hex())

		// BabyJubJub key
		pkCompStr := tv.pkCompStr
		pkComp, err := BJJFromStringWithChecksum(pkCompStr)
		require.NoError(t, err)

		chainID := tv.chainID
		hermezContractAddr := ethCommon.HexToAddress(tv.hermezContractAddr)
		a := AccountCreationAuth{
			EthAddr: ethAddr,
			BJJ:     pkComp,
		}

		toHash, err := a.toHash(chainID, hermezContractAddr)
		require.NoError(t, err)
		assert.Equal(t, tv.toHashExpected,
			hex.EncodeToString(toHash))

		msg, err := a.HashToSign(chainID, hermezContractAddr)
		require.NoError(t, err)
		assert.Equal(t, tv.hashExpected,
			hex.EncodeToString(msg))

		// sign AccountCreationAuth with eth key
		sig, err := ethCrypto.Sign(msg, ethSk)
		require.NoError(t, err)
		sig[64] += 27
		assert.Equal(t, tv.sigExpected,
			hex.EncodeToString(sig))
		a.Signature = sig
		isValid, err := a.VerifySignature(chainID, hermezContractAddr)
		require.NoError(t, err)
		assert.True(t, isValid)
	}
}
