## Transaction Selector
Package txselector is responsible to choose the transactions from the pool that will be forged in the next batch.
The main goal is to come with the most profitable selection, always respecting the constrains of the protocol.
This constrains can be splitted in two categories:

### Batch constrains

This information is passed to the txselector as `selectionConfig txprocessor.Config`:
- **NLevels**: limit the amount of accounts that can be created by NLevels^2 -1.
  **Note** that this constraint is not properly checked, if this situation is reached the entire selection will fail
- **MaxFeeTx**: maximum amount of different coordinator accounts (idx) that can be used to collect fees.
  Note that this constraint is not checked right now, if this situation is reached the `txprocessor` will fail later on preparing the `zki`
- **MaxTx**: maximum amount of transactions that can fit in a batch, in other words: len(l1UserTxs) + len(l1CoordinatorTxs) + len(l2Txs) <= MaxTx
- **MaxL1Tx**: maximum amount of L1 transactions that can fit in a batch, in other words: len(l1UserTxs) + len(l1CoordinatorTxs) <= MaxL1Tx

### Transaction constrains 

This takes into consideration the txs fetched from the pool and the current state stored in StateDB:
- **Sender account exists**: `FromIdx` must exist in the `StateDB`, and `tx.TokenID == account.TokenID` has to be respected
- **Sender account has enough balance**: tx.Amount + fee <= account.Balance
- **Sender account has correct nonce**: `tx.Nonce == account.Nonce`
- **Recipient account exists or can be created**:
    - In case of transfer to Idx: account MUST already exist on StateDB, and match the `tx.TokenID`
    - In case of transfer to Ethereum address: if the account doesn't exists, it can be created through a `l1CoordinatorTx` IF there is a valid `AccountCreationAuthorization`
    - In case of transfer to BJJ: if the account doesn't exists, it can be created through a `l1CoordinatorTx` (no need for `AccountCreationAuthorization`)
- **Atomic transactions**: requested transaction exist and can be linked,
  according to the `RqOffset` spec: https://docs.hermez.io/#/developers/protocol/hermez-protocol/circuits/circuits?id=rq-tx-txOrchestrator

### Important considerations:
- It's assumed that signatures are correct, since they're checked before inserting txs to the pool
- The state is processed sequentially meaning that each tx that is selected affects the state, in other words:
  the order in which txs are selected can make other txs became valid or invalid.
  This specially relevant for the constrains `Sender account has enough balance` and `Sender account has correct nonce`
- The creation of accounts using `l1CoordinatorTxs` increments the amount of used L1 transactions.
  This has to be taken into consideration for the constraint `MaxL1Tx`

### Current implementation:

The current approach is using pipeline concurrency pattern:
0. Process L1UserTxs (this transactions come from the Blockchain and it's mandatory by protocol to forge them)
1. Get transactions from the pool
2. Order transactions by (nonce, fee in USD)
3. Selection process:
   1. Init tx orchestrator with selection config, tx processor and tx selector. Tx orchestrator is responsible for distribution transactions to the according channels
   2. In the main thread tx selector is listening to the channels and doing according action
   3. In the separated goroutine tx orchestrator is distributing transactions. On every step we also check for errors and failed atomic groups. If error or failed atomic group is faced orchestrator exit tx selection process
      1. Transform l2txs array to the l2txs channel
      2. Check that batch in tx is less than max num batch. If not, send tx to unforjable txs channel
      3. Check that exit has amount more than zero. If not, send tx to unforjable txs channel
      4. Check for enough space for l1/l2 tx, enough balance on sender and correct nonce. If not, send to non selected txs channel
      5. Try to process l2 txs. If process is successful, than orchestrator can send tx to the selected txs channel
      6. Repeat the process of distributing tx, if any tx was sent to the selected txs channel. If not, end selection process
         ![](https://i.imgur.com/3M3bFKo.png)

4. Choose coordinator idxs to collect the fees. Note that `MaxFeeTx` constrain is ignored in this step.
5. Return the selected L2 txs as well as the necessary `l1CoordinatorTxs` the `l1UserTxs`
   (this is redundant as the return will always be the same as the input) and the coordinator idxs used to collect fees

The previous flow description doesn't take in consideration the constrain `Atomic transactions`.
This constrain alters the previous step as follow:
- Atomic transactions are grouped into `AtomicGroups`, and each group has an average fee that is used to sort the transactions together with the non atomic transactions,
  in a way that all the atomic transactions from the same group preserve the relative order as found in the pool.
  This is done in this way because it's assumed that transactions from the same `AtomicGroup`
  have the same `AtomicGroupID`, and are ordered with valid `RqOffset` in the pool.
- If one atomic transaction fails to be processed in the `Selection loop`,
  the group will be marked as invalid and the entire process will reboot from the beginning with the only difference being that
  txs belonging to failed atomic groups will be discarded before reaching the `Selection loop`.
  This is done this way because the state is altered sequentially, so if a transaction belonging to an atomic group is selected,
  but later on a transaction from the same group can't be selected, the selection will be invalid since there will be a selected tx that depends on a tx that
  doesn't exist in the selection. Right now the mechanism that the StateDB has to revert changes is to go back to a previous checkpoint (checkpoints are created per batch).
  This limitation forces the txselector to restart from the beginning of the batch selection.
  This should be improved once the StateDB has more granular mechanisms to revert the effects of processed txs.
