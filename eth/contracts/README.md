## Contracts

The go code of the contracts has been generated with the following command:

```
abigen --abi=WithdrawalDelayer.abi --bin=WithdrawalDelayer.bin --pkg=WithdrawalDelayer --out=WithdrawalDelayer.go
abigen --abi=Hermez.abi --bin=Hermez.bin --pkg=Hermez --out=Hermez.go
abigen --abi=HermezAuctionProtocol.abi --bin=HermezAuctionProtocol.bin --pkg=HermezAuctionProtocol --out=HermezAuctionProtocol.go
```
You must compile the contracts to get the `.bin` and `.abi` files. The contracts used are in the repo: https://github.com/hermeznetwork/contracts-circuits

Specifically they have been processed in the commit with hash: `745e8d588496d7762d4084a54bafd4435061ae35`

> abigen version 1.9.21

---

ERC20 go code was generated with the following command:
```
abigen --sol erc20.sol --pkg erc20 --out erc20/erc20.go
```

Versions:
```
 $ abigen --version
abigen version 1.9.21-stable-0287d548
 $ solc --version
solc, the solidity compiler commandline interface
Version: 0.7.1+commit.f4a555be.Linux.g++
```
