## Contracts

The go code of the contracts has been generated with the following command:

```
abigen --abi=WithdrawalDelayer.abi --bin=WithdrawalDelayer.bin --pkg=WithdrawalDelayer --out=WithdrawalDelayer.go
abigen --abi=Hermez.abi --bin=Hermez.bin --pkg=Hermez --out=Hermez.go
abigen --abi=HermezAuctionProtocol.abi --bin=HermezAuctionProtocol.bin --pkg=HermezAuctionProtocol --out=HermezAuctionProtocol.go
abigen --abi=ERC777.abi --bin=ERC777.bin --pkg=ERC777 --out=ERC777.go
```
You must compile the contracts to get the `.bin` and `.abi` files. The contracts used are in the repo: https://github.com/hermeznetwork/contracts

Specifically they have been processed in the commit with hash: `729b5b95e9d0df021aac6b601e18224cb0efee97`

> abigen version 1.9.21