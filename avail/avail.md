## Avail integration status

At this stage hez-node is syncing with avail network, but l2txsData are not tested.
Avail client sends l2tx to the avail by this format - "'stateRoot' + root + 'l1l2TxData'" and syncronizer is syncing with the avail, 
parse every block, and try to find root in the block data. If it found, then process txs

There are the avail api examples - https://github.com/prabal-banerjee/avail-gsrpc-examples
Avail explorer - https://devnet-avail.polygon.technology/#/explorer
Tool for running local ethereum node - https://github.com/hermeznetwork/contracts
Tool for sending txs to the hez-node - https://github.com/hermeznetwork/l2-multiple-transaction-sender
