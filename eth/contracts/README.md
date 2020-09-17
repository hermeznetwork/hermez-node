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