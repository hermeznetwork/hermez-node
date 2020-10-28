package til

// sets of instructions to be used in tests of other packages

// SetBlockchain0 contains a set of transactions simulated to be from the smart contract
var SetBlockchain0 = `
// Set containing Blockchain transactions
Type: Blockchain
AddToken(1)
AddToken(2)
AddToken(3)

// block:0 batch:0

// Coordinator accounts, Idxs: 256, 257, 258, 259
CreateAccountCoordinator(0) Coord
CreateAccountCoordinator(1) Coord
CreateAccountCoordinator(2) Coord
CreateAccountCoordinator(3) Coord

> batch
// block:0 batch:1

// deposits TokenID: 1
CreateAccountDeposit(1) A: 50
CreateAccountDeposit(1) B: 5
CreateAccountDeposit(1) C: 200
CreateAccountDeposit(1) D: 25
CreateAccountDeposit(1) E: 25
CreateAccountDeposit(1) F: 25
CreateAccountDeposit(1) G: 25
CreateAccountDeposit(1) H: 25
CreateAccountDeposit(1) I: 25
CreateAccountDeposit(1) J: 25
CreateAccountDeposit(1) K: 25
CreateAccountDeposit(1) L: 25
CreateAccountDeposit(1) M: 25
CreateAccountDeposit(1) N: 25
CreateAccountDeposit(1) O: 25
CreateAccountDeposit(1) P: 25
CreateAccountDeposit(1) Q: 25
CreateAccountDeposit(1) R: 25
CreateAccountDeposit(1) S: 25
CreateAccountDeposit(1) T: 25
CreateAccountDeposit(1) U: 25
CreateAccountDeposit(1) V: 25
CreateAccountDeposit(1) W: 25
CreateAccountDeposit(1) X: 25
CreateAccountDeposit(1) Y: 25
CreateAccountDeposit(1) Z: 25
// deposits TokenID: 2
CreateAccountDeposit(2) B: 5
CreateAccountDeposit(2) A: 20
// deposits TokenID: 3
CreateAccountDeposit(3) B: 100
// deposits TokenID: 0
CreateAccountDeposit(0) B: 10000
CreateAccountDeposit(0) C: 1

> batchL1
// block:0 batch:2

// transactions TokenID: 1
Transfer(1) A-B: 5 (1)
Transfer(1) A-L: 10 (1)
Transfer(1) A-M: 5 (1)
Transfer(1) A-N: 5 (1)
Transfer(1) A-O: 5 (1)
Transfer(1) B-C: 3 (1)
Transfer(1) C-A: 10 (200)
Transfer(1) D-A: 5 (1)
Transfer(1) D-Z: 5 (1)
Transfer(1) D-Y: 5 (1)
Transfer(1) D-X: 5 (1)
Transfer(1) E-Z: 5 (2)
Transfer(1) E-Y: 5 (1)
Transfer(1) E-X: 5 (1)
Transfer(1) F-Z: 5 (1)
Transfer(1) G-K: 3 (1)
Transfer(1) G-K: 3 (1)
Transfer(1) G-K: 3 (1)
Transfer(1) H-K: 3 (2)
Transfer(1) H-K: 3 (1)
Transfer(1) H-K: 3 (1)
Transfer(0) B-C: 50 (192)

> batchL1
> block

// block:1 batch:0

// A (3) still does not exist, coordinator should create new L1Tx to create the account
CreateAccountCoordinator(3) A

Transfer(1) A-B: 1 (1)
Transfer(1) A-B: 1 (1)
Transfer(1) A-B: 1 (1)
Transfer(3) B-A: 5 (1)
Transfer(2) A-B: 5 (1)
Transfer(1) I-K: 3 (1)
Transfer(1) I-K: 3 (1)
Transfer(1) I-K: 3 (1)
Transfer(1) J-K: 3 (1)
Transfer(1) J-K: 3 (1)
Transfer(1) J-K: 3 (1)
Transfer(1) K-J: 3 (1)
Transfer(1) L-A: 5 (1)
Transfer(1) L-Z: 5 (1)
Transfer(1) L-Y: 5 (1)
Transfer(1) L-X: 5 (1)
Transfer(1) M-A: 5 (1)
Transfer(1) M-Z: 5 (1)
Transfer(1) M-Y: 5 (1)
Transfer(1) N-A: 5 (1)
Transfer(1) N-Z: 5 (2)
Transfer(1) N-Y: 5 (1)
Transfer(1) O-T: 3 (1)
Transfer(1) O-U: 3 (1)
Transfer(1) O-V: 3 (1)
Transfer(1) P-T: 3 (1)
Transfer(1) P-U: 3 (1)
Transfer(1) P-V: 3 (5)
Transfer(1) Q-O: 3 (1)
Transfer(1) Q-P: 3 (1)
Transfer(1) R-O: 3 (1)
Transfer(1) R-P: 3 (1)
Transfer(1) R-Q: 3 (1)
Transfer(1) S-O: 3 (1)
Transfer(1) S-P: 3 (1)
Transfer(1) S-Q: 3 (1)
Transfer(1) T-O: 3 (1)
Transfer(1) T-P: 3 (1)
Transfer(1) T-Q: 3 (1)
Transfer(1) U-Z: 5 (3)
Transfer(1) U-Y: 5 (1)
Transfer(1) U-T: 3 (1)
Transfer(1) V-Z: 5 (0)
Transfer(1) V-Y: 6 (1)
Transfer(1) V-T: 3 (1)
Transfer(1) W-K: 3 (1)
Transfer(1) W-J: 3 (1)
Transfer(1) W-A: 5 (1)
Transfer(1) W-Z: 5 (1)
Transfer(1) X-B: 5 (1)
Transfer(1) X-C: 10 (200)
Transfer(1) X-D: 5 (1)
Transfer(1) X-E: 5 (1)
Transfer(1) Y-B: 5 (1)
Transfer(1) Y-C: 5 (1)
Transfer(1) Y-D: 5 (1)
Transfer(1) Y-E: 5 (1)
Transfer(1) Z-A: 5 (1)
// exits
ForceExit(1) A: 5
Exit(1) K: 5 (1)
Exit(1) X: 5 (1)
Exit(1) Y: 5 (1)
Exit(1) Z: 5 (1)

> batch
// block:1 batch:1

Deposit(1) A: 50
Deposit(1) B: 5
Deposit(1) C: 20
Deposit(1) D: 25
Deposit(1) E: 25
Deposit(1) F: 25
Deposit(1) G: 25
Deposit(1) H: 25
Deposit(1) I: 25
Transfer(1) A-B: 5 (1)
Transfer(1) A-L: 10 (1)
Transfer(1) A-M: 5 (1)
Transfer(1) B-N: 5 (1)
Transfer(1) C-O: 5 (1)
Transfer(1) H-O: 5 (1)
Transfer(1) I-H: 5 (1)
Exit(1) A: 5 (1)

// create CoordinatorTx CreateAccount for D, TokenId 2, used at SetPool0 for 'PoolTransfer(2) B-D: 3 (1)'
CreateAccountCoordinator(2) D

> batchL1
> batchL1
> block
`

// SetPool0 contains a set of transactions from the PoolL2
var SetPool0 = `
Type: PoolL2
PoolTransfer(1) A-B: 6 (1)
PoolTransfer(1) B-C: 3 (1)
PoolTransfer(1) C-A: 3 (1)
PoolTransfer(1) A-B: 1 (1)
PoolTransfer(2) A-B: 15 (1)
PoolTransfer(2) B-D: 3 (1)
PoolExit(1) A: 3 (1)
PoolTransfer(1) A-B: 6 (1)
PoolTransfer(1) B-C: 3 (1)
PoolTransfer(1) A-C: 3 (1)
PoolTransferToEthAddr(1) A-C: 3 (1)
PoolTransferToBJJ(1) A-C: 3 (1)
`
