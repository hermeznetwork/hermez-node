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
	ethSk, err := ethCrypto.HexToECDSA("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
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
	assert.Equal(t, "4f8df75e96fdce1ac90bb2f8d81c42047600f85bfcef80ce3b91c2a2afc58c1e",
		hex.EncodeToString(hash))
	sig, err := ethCrypto.Sign(hash, ethSk)
	require.NoError(t, err)
	sig[64] += 27
	assert.Equal(t, sig, a.Signature)

	assert.True(t, a.VerifySignature(chainID, hermezContractAddr))
}

func TestKeccak256JSComp(t *testing.T) {
	// check keccak256 compatible with js version
	h := ethCrypto.Keccak256Hash([]byte("test")).Bytes()
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
	tv0 := testVector{
		ethSk:              "0000000000000000000000000000000000000000000000000000000000000001",
		expectedAddress:    "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf",
		pkCompStr:          "21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7",
		chainID:            uint16(4),
		hermezContractAddr: "0x7e5f4552091a69125d5dfcb7b8c2659029395bdf",
		toHashExpected:     "19457468657265756d205369676e6564204d6573736167653a0a3132304920617574686f72697a65207468697320626162796a75626a7562206b657920666f72206865726d657a20726f6c6c7570206163636f756e74206372656174696f6e21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700047e5f4552091a69125d5dfcb7b8c2659029395bdf",
		hashExpected:       "39afea52d843a4de905b6b5ebb0ee8c678141f711d96d9b429c4aec10ef9911f",
		sigExpected:        "73d10d6ecf06ee8a5f60ac90f06b78bef9c650f414ba3ac73e176dc32e896159147457e9c86f0b4bd60fdaf2c0b2aec890a7df993d69a4805e242a6b845ebf231c",
	}
	tv1 := testVector{
		ethSk:              "0000000000000000000000000000000000000000000000000000000000000002",
		expectedAddress:    "0x2B5AD5c4795c026514f8317c7a215E218DcCD6cF",
		pkCompStr:          "093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d",
		chainID:            uint16(0),
		hermezContractAddr: "0x2b5ad5c4795c026514f8317c7a215e218dccd6cf",
		toHashExpected:     "19457468657265756d205369676e6564204d6573736167653a0a3132304920617574686f72697a65207468697320626162796a75626a7562206b657920666f72206865726d657a20726f6c6c7570206163636f756e74206372656174696f6e093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d00002b5ad5c4795c026514f8317c7a215e218dccd6cf",
		hashExpected:       "89a3895993a4736232212e59566294feb3da227af44375daf3307dcad5451d5d",
		sigExpected:        "bb4156156c705494ad5f99030342c64657e51e2994750f92125717c40bf56ad632044aa6bd00979feea92c417b552401e65fe5f531f15010d9d1c278da8be1df1b",
	}
	tv2 := testVector{
		ethSk:              "c5e8f61d1ab959b397eecc0a37a6517b8e67a0e7cf1f4bce5591f3ed80199122",
		expectedAddress:    "0xc783df8a850f42e7F7e57013759C285caa701eB6",
		pkCompStr:          "22870c1bcc451396202d62f566026eab8e438c6c91decf8ddf63a6c162619b52",
		chainID:            uint16(31337), // =0x7a69
		hermezContractAddr: "0xf4e77E5Da47AC3125140c470c71cBca77B5c638c",
		toHashExpected:     "19457468657265756d205369676e6564204d6573736167653a0a3132304920617574686f72697a65207468697320626162796a75626a7562206b657920666f72206865726d657a20726f6c6c7570206163636f756e74206372656174696f6e22870c1bcc451396202d62f566026eab8e438c6c91decf8ddf63a6c162619b527a69f4e77e5da47ac3125140c470c71cbca77b5c638c",
		hashExpected:       "4f6ead01278ba4597d4720e37482f585a713497cea994a95209f4c57a963b4a7",
		sigExpected:        "43b5818802a137a72a190c1d8d767ca507f7a4804b1b69b5e055abf31f4f2b476c80bb1ba63260d95610f6f831420d32130e7f22fec5d76e16644ddfcedd0d441c",
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

		toHash := a.toHash(chainID, hermezContractAddr)
		assert.Equal(t, tv.toHashExpected,
			hex.EncodeToString(toHash))
		assert.Equal(t, 120+len(EthMsgPrefix)+len([]byte("120")), len(toHash))

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

		assert.True(t, a.VerifySignature(chainID, hermezContractAddr))
	}
}
