## Contracts

The go code of the contracts has been generated with the following command:

```
abigen --abi=WithdrawalDelayer.abi --bin=WithdrawalDelayer.bin --pkg=WithdrawalDelayer --out=WithdrawalDelayer.go
abigen --abi=Hermez.abi --bin=Hermez.bin --pkg=Hermez --out=Hermez.go
abigen --abi=HermezAuctionProtocol.abi --bin=HermezAuctionProtocol.bin --pkg=HermezAuctionProtocol --out=HermezAuctionProtocol.go
abigen --abi=HEZ.abi --bin=HEZ.bin --pkg=HEZ --out=HEZ.go
```
You must compile the contracts to get the `.bin` and `.abi` files. The contracts used are in the repo: https://github.com/hermeznetwork/contracts

Branch: `feature/newDeploymentScript-eth-edu`
at the commit with hash: `e6c5b7db8da2de1b9cc55e281c8d1dfa524b06f0`

Alternatively, you can run the `update.sh` script like this:
```
./update.sh CONTRACT_REPO_PATH
```

Versions:
```
 solidity version 0.6.12
```
```
 $ abigen --version
abigen version 1.9.25-stable-e7872729
```
