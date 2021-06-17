# Spec for adding atomic txs

## Legend

- *Italic means inprovement proposal, may be just some thinking while going through the code*
- **Bold means potential error/bug/undesired side effect**

## Current flow

1. Process L1 User txs
2. `l2TxsForgable, l2TxsNonForgable := splitL2ForgableAndNonForgable(tp, l2TxsFromDB)` (checks if the nonce is inmediatly valid and if the fromIdx exists). *Add description of the function, maybe add the other checks needed to state if the tx is forjable inmediatly (destinatary exists, balance (although balance may be affected by other txs, so has to be checked in another way))*
3. If there are no `l2TxsForgable`, return
4. Sort `l2TxsForgable` by (fee, nonce) *It's pointless to sort by nonce, because for each account only one tx is within this group (the one that can be forged with the current nonce). Should only sort by fee*
5. Process (`processL2Txs`) `l2TxsForgable` *Personaly I don't like methods that get pointers as input, and return the same pointers*
   1. (From now on, steps on this identation level are done for each `l2TxsForgable`)
   2. Check if there is room for more L2 txs, if not return
   3. Sanity check: discard exits with amount 0
   4. Check balance
   5. Check nonce *This is probably not needed, since for the forjable txs we already filtered in a way that each account only gets the current valid nonce*
   6. Check if coordinator has account to get the fee. If not, try to create and process it, if the L1 max capacity is already reached, discard the L2 tx *If the different tokenIDs that can be used to get fees is reached shouldn't create an account, and discard the L2 anyway (although filtering here could get worst results as even if txs are sorted by fee, there could be more txs with lower fee for this particular token that end up being more profitable)*
   7. If it's a transfer to ether addr or to bjj
      1. `txsel.processTxToEthAddrBJJ`: add a L1CoordTxn the array in the case destinatary doesn't exist and it's posible to create it. If it's not posible to create it (because missing authorization or L1 already reached), and destinatary doesn't exist the L2Tx is marked as invalid *change the name of the function as `process` it's widely used when the function alters the state*
      2. If the l2tx is valid but requires a l1CoordTx (to create an account for the destinatary), and the L1 limit is already reached, discard the l2tx *It looks like this check is already done in `txsel.processTxToEthAddrBJJ`*
      3. Process l1CoordTx (if a new tx was needed to create destinatary account)
   8. If it's a transfer to idx
      1. Check if the destinatary exists, if not discard the l2tx
   9. *Checks for transfer to eth addr + transfer to bjj vs to idx are both using ifs. It should be impossible to have any other case, but I'd like to add an else (instead of else if) or an additional else to make this fully robust*
   10. Get coordinator idx to receive fee, and if not present on `tp.AccumulatedFees`, add it
   11. Process L2 tx
   12. **Some L1 txs could be created as a "side effect" of a L2 tx (create destinatary account or coord account to receive feer), and later on the L2 tx be discarded. This is not a bug, but it could create unnecesary accounts**
6. If there is room for more L2Txs, repeat the process of `processL2Txs` but with the ones that was marked as `l2TxsNonForgable` in the beginning, since they could now become forjable (mostly due to nonce advancing). If max capacity for L2s is reachead, mark txs as not selected because of that
7. Build and sort coordIdxs
8. Distribute accumulated fees
9. Make checkpoint
10. **MaxFeeTx is not checked, so we could exceed the amount of idxs used by the coordinator to collect the fees.** I've to double check if this is evaluated in `txprocessor`, and what kind of error it returns and how is this error handled

## Implementation proposal for atomic txs

### TODOs having in mind current flow

- We need to set `RqOffset` for the atomic txs
- We need a mechanism to inform `processL2Txs` that a given tx is part of a specific atomic group (knowing that is atomic which could be done by checkint `RqTxID` is not enougth)
- The main problem is that L2xs are processed one by one (step 5.11), so if we iterated normally and for any reason a tx from an atomic group was rejected, we'd had to invalidate the already processed txs from that group.

#### Seting `RqOffset`

This should be done inside `splitL2ForgableAndNonForgable`, since this is done in the begining, this way we can discard txs that cannot be forged no matter what. This could happen for two different reasons:

- Missing tx(s) to build the complete atomic group. To do this filtering we can reuse this code:

```go
// This code can be find in feature/atomic-tx branch, file txselector/txbatch.go

// buildAtomicTxs build the atomic transactions groups and add into a mapping
func buildAtomicTxs(poolTxs []common.PoolL2Tx) (map[common.TxID][]common.PoolL2Tx, map[common.TxID]bool, map[common.TxID]common.TxID) {
	atomics := make(map[common.TxID][]common.PoolL2Tx)
	discarded := make(map[common.TxID]bool)
	owners := make(map[common.TxID]common.TxID)
	if len(poolTxs) == 0 {
		return atomics, discarded, owners
	}
	txMap := make(map[common.TxID]bool)
	for _, tx := range poolTxs {
		txMap[tx.TxID] = true
	}
	for _, tx := range poolTxs {
		// check if the tx rq tx exist
		_, ok := txMap[tx.RqTxID]
		if tx.RqTxID != common.EmptyTxID && !ok {
			discarded[tx.TxID] = true
			continue
		}

		// check if the tx already have a group owner
		rootTxID, ok := owners[tx.TxID]
		if !ok {
			rootTxID = tx.TxID
		}
		// check if the root tx already exist into the mapping
		txs, ok := atomics[rootTxID]
		if ok {
			// only add if exist
			atomics[rootTxID] = append(txs, tx)
		} else if tx.RqTxID != common.EmptyTxID {
			// if not exist, check if the nested atomic transaction exist
			auxTxID, ok := owners[tx.RqTxID]
			if ok {
				// set the nested atomic as a root and add the child
				rootTxID = auxTxID
				atomics[rootTxID] = append(atomics[rootTxID], tx)
			} else {
				// create a new atomic group if not exist
				atomics[rootTxID] = []common.PoolL2Tx{tx}
			}
		} else {
			// create a new atomic group if not exist
			atomics[rootTxID] = []common.PoolL2Tx{tx}
		}
		// add the tx to the owner mapping
		if tx.RqTxID != common.EmptyTxID {
			owners[tx.RqTxID] = rootTxID
		} else {
			owners[rootTxID] = tx.TxID
		}
	}
	// sanitize the atomic transaction removing the non-atomics
	for key, group := range atomics {
		if len(group) > 1 {
			continue
		}
		delete(atomics, key)
		delete(owners, key)
		tx := group[0]
		if tx.RqTxID != common.EmptyTxID {
			discarded[tx.TxID] = true
		}
	}
	return atomics, discarded, owners
}
```

- It's impossible to set corret RqOffset within the group.

Note that both cases should be avoidable if the API filters such txs. Current state:

- atomic endpoint already rejects groups that are incomplete (missing tx(s) to "close" the group), however right now it's posible to send atomic txs with the normal `POST /transactions-pool` endpoint, we shuld consider filtering such txs in that endpoint.
- we haven't proper logic for the `RqOffset` yet, and it's more tricky than it seems (the amount of txs that can fit in a group it's greater than 7, which is the maximum distance that the [protocol](https://docs.hermez.io/#/developers/protocol/hermez-protocol/circuits/circuits?id=rq-tx-verifier) allow from one tx to another). Once we figure out the actual maximum we could also add a filter in the atomic endpoint.

In conclusion the code should:

- create the groups using `buildAtomicTxs`
- set the `RqOffset` using some logic that we don't have right now
- discard txs that are inside malformed groups

#### Mechanism to differentiate atomic from non atomic

This can be done in many different ways, but it should also be done in `splitL2ForgableAndNonForgable`, since this function is the one that will be aware of which atomic tx belongs to which atomic group.

The requisits are:

- we should know if a particular tx is atomic or not
- given an atomic tx, we should be able to find the txs from the same atomic group (so if one is discarded, we can also discard the others)

Draft idea:

```go
// Warning, this prop

type atomicGroup {
    TxIDs      []common.TxID
    AverageFee float64
}
type selectableTx struct {
    Tx      common.PoolL2Tx
    GroupID int // 0 means not belonging to a group, so non atomic tx
}

func splitL2ForgableAndNonForgable(tp *txprocessor.TxProcessor, // same as before
	l2Txs []common.PoolL2Tx) (  // same as before
        []selectableTx,         // change return type, but same concept
        map[int]atomicGroup,    // new return object that helps finding all the txs within the same group. 
                                // int is used as "atomic group id", we could create a type to increase readability
        []common.PoolL2Tx,     // same as before
    ) {
```

#### Processing atomic txs in a batch

Differnt ideas to tackle the issue:

- Make sure that atomic txs are processed consecutively (they get into the 5.1 consecutively), then make a checkpoint before starting to process the first atomic tx of the group. If one of them fail, go back to the checkpoint and discard the entire group
- Move step 5.11 outside the loop, so we can remove the l2txs from the array before being processed, and once we're sure they are all correct we process them all in a for loop.
