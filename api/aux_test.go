package api

import (
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree"
)

func AddAdditionalInformation(blocks []common.BlockData) {
	for i := range blocks {
		blocks[i].Block.Timestamp = time.Now().Add(time.Second * 13).UTC()
		blocks[i].Block.Hash = ethCommon.BigToHash(big.NewInt(blocks[i].Block.Num))
		for j := range blocks[i].Rollup.AddedTokens {
			blocks[i].Rollup.AddedTokens[j].Name = "Test Token " + strconv.Itoa(int(blocks[i].Rollup.AddedTokens[j].TokenID))
			blocks[i].Rollup.AddedTokens[j].Symbol = "TKN" + strconv.Itoa(int(blocks[i].Rollup.AddedTokens[j].TokenID))
			blocks[i].Rollup.AddedTokens[j].Decimals = 18
		}
		for x := range blocks[i].Rollup.Batches {
			for q := range blocks[i].Rollup.Batches[x].CreatedAccounts {
				blocks[i].Rollup.Batches[x].CreatedAccounts[q].Balance =
					big.NewInt(int64(blocks[i].Rollup.Batches[x].CreatedAccounts[q].Idx * 10000000))
			}
			for y := range blocks[i].Rollup.Batches[x].ExitTree {
				blocks[i].Rollup.Batches[x].ExitTree[y].MerkleProof =
					&merkletree.CircomVerifierProof{
						Root: &merkletree.Hash{byte(y), byte(y + 1)},
						Siblings: []*merkletree.Hash{
							merkletree.NewHashFromBigInt(big.NewInt(int64(y) * 10)),
							merkletree.NewHashFromBigInt(big.NewInt(int64(y)*100 + 1)),
							merkletree.NewHashFromBigInt(big.NewInt(int64(y)*1000 + 2))},
						OldKey:   &merkletree.Hash{byte(y * 1), byte(y*1 + 1)},
						OldValue: &merkletree.Hash{byte(y * 2), byte(y*2 + 1)},
						IsOld0:   y%2 == 0,
						Key:      &merkletree.Hash{byte(y * 3), byte(y*3 + 1)},
						Value:    &merkletree.Hash{byte(y * 4), byte(y*4 + 1)},
						Fnc:      y % 2,
					}
			}
		}
	}
}
