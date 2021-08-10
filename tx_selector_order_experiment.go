package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

type tx struct {
	Idx, Nonce, AbsoluteFee int
}

var nonAtomicTxs = []tx{
	{1, 1, 1},
	{1, 2, 2},
	{1, 3, 3},

	{2, 1, 3},
	{2, 2, 2},
	{2, 3, 1},

	{3, 2, 3},
	{3, 3, 3},
	{3, 4, 3},

	{4, 2, 2},
	{4, 3, 2},
	{4, 4, 2},
}

// var nonAtomicTxs = []tx{
// 	{1, 1, 4},
// 	{1, 2, 3},
// 	{1, 3, 2},
// 	{1, 4, 1},

// 	{2, 6, 4},
// 	{2, 7, 3},
// 	{2, 8, 2},
// 	{2, 9, 1},

// 	{3, 1, 5},
// 	{3, 2, 5},
// 	{3, 3, 5},
// 	{3, 4, 5},
// }

func main() {

	cop := append([]tx{}, nonAtomicTxs...)

	// Sort non atomic txs by absolute fee with SliceStable, so that txs with same
	// AbsoluteFee are not rearranged and nonce order is kept in such case
	sort.SliceStable(nonAtomicTxs, func(i, j int) bool {
		return nonAtomicTxs[i].AbsoluteFee > nonAtomicTxs[j].AbsoluteFee
	})
	print(nonAtomicTxs)

	// sort non atomic txs by Nonce. This can be done in many different ways, what
	// is needed is to output the l2Txs where the Nonce of l2Txs for each
	// Account is sorted, but the l2Txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.Slice(nonAtomicTxs, func(i, j int) bool {
		return nonAtomicTxs[i].Nonce < nonAtomicTxs[j].Nonce
	})
	print(nonAtomicTxs)

	sort.Slice(cop, func(i, j int) bool {
		if cop[i].AbsoluteFee != cop[j].AbsoluteFee {
			return cop[i].AbsoluteFee > cop[j].AbsoluteFee
		} else if cop[i].Idx != cop[j].Idx {
			return cop[i].Idx < cop[j].Idx
		} else {
			return cop[i].Nonce < cop[j].Nonce
		}
	})
	print(cop)

}

func print(arr []tx) {
	for i := 0; i < len(arr); i++ {
		p, _ := json.Marshal(arr[i])
		fmt.Println(string(p))
	}
	fmt.Println()
}
