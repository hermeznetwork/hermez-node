package eth

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

func addBlock(url string) {
	method := "POST"

	payload := strings.NewReader(
		"{\n    \"jsonrpc\":\"2.0\",\n    \"method\":\"evm_mine\",\n    \"params\":[],\n    \"id\":1\n}")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()
}

func addBlocks(numBlocks int64, url string) {
	for i := int64(0); i < numBlocks; i++ {
		addBlock(url)
	}
}

func addTime(seconds float64, url string) {
	secondsStr := strconv.FormatFloat(seconds, 'E', -1, 32)

	method := "POST"
	payload := strings.NewReader(
		"{\n    \"jsonrpc\":\"2.0\",\n    \"method\":\"evm_increaseTime\",\n    \"params\":[" +
			secondsStr + "],\n    \"id\":1\n}")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Println("Error when closing:", err)
		}
	}()
}

func createPermitDigest(tokenAddr, owner, spender ethCommon.Address, chainID, value, nonce,
	deadline *big.Int, tokenName string) ([]byte, error) {
	// NOTE: We ignore hash.Write errors because we are writing to a memory
	// buffer and don't expect any errors to occur.
	abiPermit :=
		[]byte("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)")
	hashPermit := sha3.NewLegacyKeccak256()
	hashPermit.Write(abiPermit) //nolint:errcheck,gosec
	abiEIP712Domain :=
		[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	hashEIP712Domain := sha3.NewLegacyKeccak256()
	hashEIP712Domain.Write(abiEIP712Domain) //nolint:errcheck,gosec
	var encodeBytes []byte
	paddedHash := ethCommon.LeftPadBytes(hashEIP712Domain.Sum(nil), 32)
	hashName := sha3.NewLegacyKeccak256()
	hashName.Write([]byte(tokenName)) //nolint:errcheck,gosec
	paddedName := ethCommon.LeftPadBytes(hashName.Sum(nil), 32)
	hashVersion := sha3.NewLegacyKeccak256()
	hashVersion.Write([]byte("1")) //nolint:errcheck,gosec
	paddedX := ethCommon.LeftPadBytes(hashVersion.Sum(nil), 32)
	paddedChainID := ethCommon.LeftPadBytes(chainID.Bytes(), 32)
	paddedAddr := ethCommon.LeftPadBytes(tokenAddr.Bytes(), 32)
	encodeBytes = append(encodeBytes, paddedHash...)
	encodeBytes = append(encodeBytes, paddedName...)
	encodeBytes = append(encodeBytes, paddedX...)
	encodeBytes = append(encodeBytes, paddedChainID...)
	encodeBytes = append(encodeBytes, paddedAddr...)
	_domainSeparator := sha3.NewLegacyKeccak256()
	_domainSeparator.Write(encodeBytes) //nolint:errcheck,gosec

	var bytes1 []byte
	paddedHashPermit := ethCommon.LeftPadBytes(hashPermit.Sum(nil), 32)
	paddedOwner := ethCommon.LeftPadBytes(owner.Bytes(), 32)
	paddedSpender := ethCommon.LeftPadBytes(spender.Bytes(), 32)
	paddedValue := ethCommon.LeftPadBytes(value.Bytes(), 32)
	paddedNonce := ethCommon.LeftPadBytes(nonce.Bytes(), 32)
	paddedDeadline := ethCommon.LeftPadBytes(deadline.Bytes(), 32)
	bytes1 = append(bytes1, paddedHashPermit...)
	bytes1 = append(bytes1, paddedOwner...)
	bytes1 = append(bytes1, paddedSpender...)
	bytes1 = append(bytes1, paddedValue...)
	bytes1 = append(bytes1, paddedNonce...)
	bytes1 = append(bytes1, paddedDeadline...)
	hashBytes1 := sha3.NewLegacyKeccak256()
	hashBytes1.Write(bytes1) //nolint:errcheck,gosec

	var bytes2 []byte
	paddedY := ethCommon.LeftPadBytes([]byte{0x19}, 1)
	paddedZ := ethCommon.LeftPadBytes([]byte{0x01}, 1)
	paddedDomainSeparator := ethCommon.LeftPadBytes(_domainSeparator.Sum(nil), 32)
	paddedHashBytes1 := ethCommon.LeftPadBytes(hashBytes1.Sum(nil), 32)
	bytes2 = append(bytes2, paddedY...)
	bytes2 = append(bytes2, paddedZ...)
	bytes2 = append(bytes2, paddedDomainSeparator...)
	bytes2 = append(bytes2, paddedHashBytes1...)
	hashBytes2 := sha3.NewLegacyKeccak256()
	hashBytes2.Write(bytes2) //nolint:errcheck,gosec

	return hashBytes2.Sum(nil), nil
}

func createPermit(owner, spender ethCommon.Address, amount, deadline *big.Int, digest,
	signature []byte) []byte {
	r := signature[0:32]
	s := signature[32:64]
	v := signature[64] + byte(27) //nolint:gomnd

	ABIpermit := []byte("permit(address,address,uint256,uint256,uint8,bytes32,bytes32)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(ABIpermit) //nolint:errcheck,gosec
	methodID := hash.Sum(nil)[:4]

	var permit []byte
	paddedOwner := ethCommon.LeftPadBytes(owner.Bytes(), 32)
	paddedSpender := ethCommon.LeftPadBytes(spender.Bytes(), 32)
	paddedAmount := ethCommon.LeftPadBytes(amount.Bytes(), 32)
	paddedDeadline := ethCommon.LeftPadBytes(deadline.Bytes(), 32)
	paddedV := ethCommon.LeftPadBytes([]byte{v}, 32)

	permit = append(permit, methodID...)
	permit = append(permit, paddedOwner...)
	permit = append(permit, paddedSpender...)
	permit = append(permit, paddedAmount...)
	permit = append(permit, paddedDeadline...)
	permit = append(permit, paddedV...)
	permit = append(permit, r...)
	permit = append(permit, s...)

	return permit
}
