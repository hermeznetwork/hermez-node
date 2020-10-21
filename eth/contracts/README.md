## Contracts

The go code of the contracts has been generated with the following command:

```
abigen --abi=WithdrawalDelayer.abi --bin=WithdrawalDelayer.bin --pkg=WithdrawalDelayer --out=WithdrawalDelayer.go
abigen --abi=Hermez.abi --bin=Hermez.bin --pkg=Hermez --out=Hermez.go
abigen --abi=HermezAuctionProtocol.abi --bin=HermezAuctionProtocol.bin --pkg=HermezAuctionProtocol --out=HermezAuctionProtocol.go
abigen --abi=HEZ.abi --bin=HEZ.bin --pkg=HEZ --out=HEZ.go
```
You must compile the contracts to get the `.bin` and `.abi` files. The contracts used are in the repo: https://github.com/hermeznetwork/contracts

Specifically they have been processed in the commit with hash: `7574ba47fd3d7dab2653a22f57b15c69280350dc`


Versions:
```
 $ abigen --version
abigen version 1.9.21-stable-0287d548
 $ solc --version
solc, the solidity compiler commandline interface
Version: 0.7.1+commit.f4a555be.Linux.g++
```
