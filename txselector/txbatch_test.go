package txselector

import (
	"math/big"
	"sort"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxBatch_prune(t *testing.T) {
	type args struct {
		txs   []*TxGroup
		maxTx uint32
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "only group prune",
			args: args{
				maxTx: 3,
				txs: []*TxGroup{
					{
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 350, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 350, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 350, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 350, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 351, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 351, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 351, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 351, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		}, {
			name: "group prune and pop",
			args: args{
				maxTx: 3,
				txs: []*TxGroup{
					{
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		}, {
			name: "group prune and pop all group",
			args: args{
				maxTx: 3,
				txs: []*TxGroup{
					{
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 350, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 350, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 350, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 350, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 351, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 351, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 351, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 351, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 349, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID2, FromIdx: 349, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID3, FromIdx: 349, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID4, FromIdx: 349, ToIdx: 444, TokenID: 0,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		}, {
			name: "pop one atomic group",
			args: args{
				maxTx: 10,
				txs: []*TxGroup{
					{
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		}, {
			name: "pop two atomic groups",
			args: args{
				maxTx: 5,
				txs: []*TxGroup{
					{
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		}, {
			name: "prune all",
			args: args{
				maxTx: 5,
				txs: []*TxGroup{
					{
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: false,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					}, {
						atomic: true,
						l2Txs: []common.PoolL2Tx{
							{AbsoluteFee: 300, TxID: txID1, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 1,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 300, TxID: txID2, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 2,
								Amount: big.NewInt(1001), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 123, TxID: txID3, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 3,
								Amount: big.NewInt(1002), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 400, TxID: txID4, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 4,
								Amount: big.NewInt(1003), Fee: 33, Type: common.TxTypeTransfer},
							{AbsoluteFee: 10, TxID: txID5, FromIdx: 352, ToIdx: 444, TokenID: 0, Nonce: 5,
								Amount: big.NewInt(1004), Fee: 33, Type: common.TxTypeTransfer},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txBatch := &TxBatch{
				txs:             tt.args.txs,
				processor:       createTxProcessorMock(),
				l2db:            createL2DBMock(),
				localAccountsDB: createStateDbMock(),
				selectionConfig: txprocessor.Config{
					MaxTx:    tt.args.maxTx,
					MaxL1Tx:  _config.MaxL1Tx,
					NLevels:  _config.NLevels,
					MaxFeeTx: _config.MaxFeeTx,
					ChainID:  _config.ChainID,
				},
				coordAccount: _coordAccount,
			}
			for _, group := range txBatch.txs {
				group.coordAccount = _coordAccount
				group.calcFeeAverage()
			}
			err := txBatch.prune()
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
			assert.EqualValues(t, tt.args.maxTx, txBatch.length())
		})
	}
}

func Test_buildTxsMap(t *testing.T) {
	type want struct {
		txAtomicMapping map[common.TxID][]common.PoolL2Tx
		idxMapping      map[common.Idx][]common.PoolL2Tx
	}
	tests := []struct {
		name    string
		poolTxs []common.PoolL2Tx
		want    want
	}{
		{
			name: "test one atomic group",
			poolTxs: []common.PoolL2Tx{
				{TxID: txID1, FromIdx: 300, RqTxID: txID2},
				{TxID: txID2, FromIdx: 420, RqTxID: txID3},
				{TxID: txID3, FromIdx: 300, RqTxID: txID4},
				{TxID: txID4, FromIdx: 340, RqTxID: txID5},
				{TxID: txID5, FromIdx: 420, RqTxID: txID6},
				{TxID: txID6, FromIdx: 300, RqTxID: txID7},
				{TxID: txID7, FromIdx: 300, RqTxID: txID1},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					txID1: {
						{TxID: txID1, FromIdx: 300, RqTxID: txID2},
						{TxID: txID2, FromIdx: 420, RqTxID: txID3},
						{TxID: txID3, FromIdx: 300, RqTxID: txID4},
						{TxID: txID4, FromIdx: 340, RqTxID: txID5},
						{TxID: txID5, FromIdx: 420, RqTxID: txID6},
						{TxID: txID6, FromIdx: 300, RqTxID: txID7},
						{TxID: txID7, FromIdx: 300, RqTxID: txID1},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{},
			},
		}, {
			name: "test two atomic groups",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID1, FromIdx: 300, RqTxID: txID2},
				{TxID: txID2, FromIdx: 420, RqTxID: txID1},
				{TxID: txID3, FromIdx: 300, RqTxID: txID1},
				// Atomic 2
				{TxID: txID4, FromIdx: 300, RqTxID: txID5},
				{TxID: txID5, FromIdx: 420, RqTxID: txID4},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID1: {
						{TxID: txID1, FromIdx: 300, RqTxID: txID2},
						{TxID: txID2, FromIdx: 420, RqTxID: txID1},
						{TxID: txID3, FromIdx: 300, RqTxID: txID1},
					},
					// Atomic 2
					txID4: {
						{TxID: txID4, FromIdx: 300, RqTxID: txID5},
						{TxID: txID5, FromIdx: 420, RqTxID: txID4},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{},
			},
		}, {
			name: "test three atomic groups",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, FromIdx: 300, RqTxID: txID3},
				{TxID: txID1, FromIdx: 300, RqTxID: txID3},
				{TxID: txID3, FromIdx: 420, RqTxID: txID2},
				// Atomic 2
				{TxID: txID5, FromIdx: 300},
				{TxID: txID4, FromIdx: 420, RqTxID: txID5},
				// Atomic 3
				{TxID: txID7, FromIdx: 300, RqTxID: txID6},
				{TxID: txID6, FromIdx: 420, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, FromIdx: 300, RqTxID: txID3},
						{TxID: txID1, FromIdx: 300, RqTxID: txID3},
						{TxID: txID3, FromIdx: 420, RqTxID: txID2},
					},
					// Atomic 2
					txID5: {
						{TxID: txID5, FromIdx: 300},
						{TxID: txID4, FromIdx: 420, RqTxID: txID5},
					},
					// Atomic 3
					txID7: {
						{TxID: txID7, FromIdx: 300, RqTxID: txID6},
						{TxID: txID6, FromIdx: 420, RqTxID: txID7},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{},
			},
		}, {
			name: "test with non atomic transactions sorted",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, FromIdx: 420, RqTxID: txID3},
				{TxID: txID1, FromIdx: 420, RqTxID: txID3},
				{TxID: txID3, FromIdx: 300, RqTxID: txID2},
				// Non-atomic
				{TxID: txID4, FromIdx: 300},
				{TxID: txID5, FromIdx: 300},
				{TxID: txID6, FromIdx: 300},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, FromIdx: 420, RqTxID: txID3},
						{TxID: txID1, FromIdx: 420, RqTxID: txID3},
						{TxID: txID3, FromIdx: 300, RqTxID: txID2},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{
					300: {
						{TxID: txID4, FromIdx: 300},
						{TxID: txID5, FromIdx: 300},
						{TxID: txID6, FromIdx: 300},
					},
				},
			},
		}, {
			name: "test with non atomic transactions sorted",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, FromIdx: 420, RqTxID: txID3},
				{TxID: txID1, FromIdx: 420, RqTxID: txID3},
				{TxID: txID3, FromIdx: 300, RqTxID: txID2},
				// Non-atomic
				{TxID: txID4, FromIdx: 300},
				{TxID: txID5, FromIdx: 300},
				{TxID: txID6, FromIdx: 420},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, FromIdx: 420, RqTxID: txID3},
						{TxID: txID1, FromIdx: 420, RqTxID: txID3},
						{TxID: txID3, FromIdx: 300, RqTxID: txID2},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{
					300: {
						{TxID: txID4, FromIdx: 300},
						{TxID: txID5, FromIdx: 300},
					},
					420: {
						{TxID: txID6, FromIdx: 420},
					},
				},
			},
		}, {
			name: "test with non atomic transactions unsorted",
			poolTxs: []common.PoolL2Tx{
				// Non-atomic
				{TxID: txID5, FromIdx: 300},
				// Atomic 1
				{TxID: txID2, FromIdx: 300, RqTxID: txID3},
				// Non-atomic
				{TxID: txID4, FromIdx: 300},
				// Atomic 1
				{TxID: txID1, FromIdx: 300, RqTxID: txID3},
				// Atomic 1
				{TxID: txID3, FromIdx: 300, RqTxID: txID2},
				// Non-atomic
				{TxID: txID6, FromIdx: 420},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, FromIdx: 300, RqTxID: txID3},
						{TxID: txID1, FromIdx: 300, RqTxID: txID3},
						{TxID: txID3, FromIdx: 300, RqTxID: txID2},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{
					300: {
						{TxID: txID5, FromIdx: 300},
						{TxID: txID4, FromIdx: 300},
					},
					420: {
						{TxID: txID6, FromIdx: 420},
					},
				},
			},
		}, {
			name:    "test a empty tx list",
			poolTxs: []common.PoolL2Tx{},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				idxMapping:      map[common.Idx][]common.PoolL2Tx{},
			},
		}, {
			name: "test only non-atomic txs",
			poolTxs: []common.PoolL2Tx{
				// Non-atomics
				{TxID: txID1, FromIdx: 420},
				{TxID: txID2, FromIdx: 420},
				{TxID: txID3, FromIdx: 420},
				{TxID: txID4, FromIdx: 300},
				{TxID: txID5, FromIdx: 303},
				{TxID: txID6, FromIdx: 303},
				{TxID: txID7, FromIdx: 301},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				idxMapping: map[common.Idx][]common.PoolL2Tx{
					300: {
						{TxID: txID4, FromIdx: 300},
					},
					301: {
						{TxID: txID7, FromIdx: 301},
					},
					303: {
						{TxID: txID5, FromIdx: 303},
						{TxID: txID6, FromIdx: 303},
					},
					420: {
						{TxID: txID1, FromIdx: 420},
						{TxID: txID2, FromIdx: 420},
						{TxID: txID3, FromIdx: 420},
					},
				},
			},
		}, {
			name: "test invalid atomic txs",
			poolTxs: []common.PoolL2Tx{
				// invalid atomics
				{TxID: txID1, FromIdx: 300, RqTxID: txID2},
				{TxID: txID6, FromIdx: 300, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				idxMapping:      map[common.Idx][]common.PoolL2Tx{},
			},
		}, {
			name: "test invalid and valid atomic txs",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, FromIdx: 300, RqTxID: txID3},
				{TxID: txID1, FromIdx: 300, RqTxID: txID3},
				{TxID: txID3, FromIdx: 300, RqTxID: txID2},
				// Non-atomic
				{TxID: txID4, FromIdx: 300, RqTxID: txID5},
				{TxID: txID6, FromIdx: 300, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, FromIdx: 300, RqTxID: txID3},
						{TxID: txID1, FromIdx: 300, RqTxID: txID3},
						{TxID: txID3, FromIdx: 300, RqTxID: txID2},
					},
				},
				idxMapping: map[common.Idx][]common.PoolL2Tx{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txAtomicMapping, idxMapping := buildTxsMap(tt.poolTxs)
			assert.EqualValues(t, tt.want.txAtomicMapping, txAtomicMapping)
			assert.EqualValues(t, tt.want.idxMapping, idxMapping)
		})
	}
}

func TestTxBatch_sort(t *testing.T) {
	tests := []struct {
		name string
		txs  []*TxGroup
		want []*TxGroup
	}{
		{
			name: "test non-atomic values",
			txs: []*TxGroup{
				{atomic: false, feeAverage: big.NewFloat(430390)},
				{atomic: false, feeAverage: big.NewFloat(10)},
				{atomic: false, feeAverage: big.NewFloat(1)},
				{atomic: false, feeAverage: big.NewFloat(0.5)},
				{atomic: false, feeAverage: big.NewFloat(100)},
				{atomic: false, feeAverage: big.NewFloat(101)},
				{atomic: false, feeAverage: big.NewFloat(0.0000005)},
				{atomic: false, feeAverage: big.NewFloat(1340)},
				{atomic: false, feeAverage: big.NewFloat(0.111)},
				{atomic: false, feeAverage: big.NewFloat(3400)},
			},
			want: []*TxGroup{
				{atomic: false, feeAverage: big.NewFloat(430390)},
				{atomic: false, feeAverage: big.NewFloat(3400)},
				{atomic: false, feeAverage: big.NewFloat(1340)},
				{atomic: false, feeAverage: big.NewFloat(101)},
				{atomic: false, feeAverage: big.NewFloat(100)},
				{atomic: false, feeAverage: big.NewFloat(10)},
				{atomic: false, feeAverage: big.NewFloat(1)},
				{atomic: false, feeAverage: big.NewFloat(0.5)},
				{atomic: false, feeAverage: big.NewFloat(0.111)},
				{atomic: false, feeAverage: big.NewFloat(0.0000005)},
			},
		}, {
			name: "test atomic values",
			txs: []*TxGroup{
				{atomic: true, feeAverage: big.NewFloat(10)},
				{atomic: true, feeAverage: big.NewFloat(310)},
				{atomic: true, feeAverage: big.NewFloat(1)},
				{atomic: true, feeAverage: big.NewFloat(100)},
			},
			want: []*TxGroup{
				{atomic: true, feeAverage: big.NewFloat(310)},
				{atomic: true, feeAverage: big.NewFloat(100)},
				{atomic: true, feeAverage: big.NewFloat(10)},
				{atomic: true, feeAverage: big.NewFloat(1)},
			},
		}, {
			name: "test atomic and non-atomic values",
			txs: []*TxGroup{
				{atomic: false, feeAverage: big.NewFloat(430390)},
				{atomic: false, feeAverage: big.NewFloat(10)},
				{atomic: false, feeAverage: big.NewFloat(1)},
				{atomic: true, feeAverage: big.NewFloat(0.5)},
				{atomic: true, feeAverage: big.NewFloat(100)},
				{atomic: true, feeAverage: big.NewFloat(101)},
				{atomic: false, feeAverage: big.NewFloat(0.0000005)},
				{atomic: false, feeAverage: big.NewFloat(1340)},
				{atomic: true, feeAverage: big.NewFloat(0.111)},
				{atomic: false, feeAverage: big.NewFloat(3400)},
			},
			want: []*TxGroup{
				{atomic: true, feeAverage: big.NewFloat(101)},
				{atomic: true, feeAverage: big.NewFloat(100)},
				{atomic: true, feeAverage: big.NewFloat(0.5)},
				{atomic: true, feeAverage: big.NewFloat(0.111)},
				{atomic: false, feeAverage: big.NewFloat(430390)},
				{atomic: false, feeAverage: big.NewFloat(3400)},
				{atomic: false, feeAverage: big.NewFloat(1340)},
				{atomic: false, feeAverage: big.NewFloat(10)},
				{atomic: false, feeAverage: big.NewFloat(1)},
				{atomic: false, feeAverage: big.NewFloat(0.0000005)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txBatch := &TxBatch{txs: tt.txs}
			txBatch.sort()
			isSorted := sort.SliceIsSorted(txBatch.txs, func(i, j int) bool {
				txI := txBatch.txs[i]
				txJ := txBatch.txs[j]
				// atomic transactions always first
				if txI.atomic != txJ.atomic {
					return txI.atomic
				}
				// sort by the highest fee
				return txI.feeAverage.Cmp(txJ.feeAverage) > 0
			})
			assert.True(t, isSorted)
			assert.EqualValues(t, tt.want, txBatch.txs)
		})
	}
}

func Test_buildAtomicTxs(t *testing.T) {
	type want struct {
		txAtomicMapping map[common.TxID][]common.PoolL2Tx
		discarded       map[common.TxID]bool
		usedTxs         map[common.TxID]common.TxID
	}
	tests := []struct {
		name    string
		poolTxs []common.PoolL2Tx
		want    want
	}{
		{
			name: "test one atomic group",
			poolTxs: []common.PoolL2Tx{
				{TxID: txID1, RqTxID: txID2},
				{TxID: txID2, RqTxID: txID3},
				{TxID: txID3, RqTxID: txID4},
				{TxID: txID4, RqTxID: txID5},
				{TxID: txID5, RqTxID: txID6},
				{TxID: txID6, RqTxID: txID7},
				{TxID: txID7, RqTxID: txID1},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					txID1: {
						{TxID: txID1, RqTxID: txID2},
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID4},
						{TxID: txID4, RqTxID: txID5},
						{TxID: txID5, RqTxID: txID6},
						{TxID: txID6, RqTxID: txID7},
						{TxID: txID7, RqTxID: txID1},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs: map[common.TxID]common.TxID{txID1: txID1, txID2: txID1, txID3: txID1, txID4: txID1, txID5: txID1,
					txID6: txID1, txID7: txID1},
			},
		}, {
			name: "test two atomic groups",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID1, RqTxID: txID2},
				{TxID: txID2, RqTxID: txID1},
				{TxID: txID3, RqTxID: txID1},
				// Atomic 2
				{TxID: txID4, RqTxID: txID5},
				{TxID: txID5, RqTxID: txID4},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID1: {
						{TxID: txID1, RqTxID: txID2},
						{TxID: txID2, RqTxID: txID1},
						{TxID: txID3, RqTxID: txID1},
					},
					// Atomic 2
					txID4: {
						{TxID: txID4, RqTxID: txID5},
						{TxID: txID5, RqTxID: txID4},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs:   map[common.TxID]common.TxID{txID1: txID1, txID2: txID1, txID4: txID4, txID5: txID4},
			},
		}, {
			name: "test two atomic groups",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID1, RqTxID: txID2},
				{TxID: txID2, RqTxID: txID3},
				{TxID: txID3, RqTxID: txID1},
				// Atomic 2
				{TxID: txID4, RqTxID: txID5},
				{TxID: txID5},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID1: {
						{TxID: txID1, RqTxID: txID2},
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID1},
					},
					// Atomic 2
					txID4: {
						{TxID: txID4, RqTxID: txID5},
						{TxID: txID5},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs:   map[common.TxID]common.TxID{txID1: txID1, txID2: txID1, txID3: txID1, txID4: txID5, txID5: txID4},
			},
		}, {
			name: "test three atomic groups",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, RqTxID: txID3},
				{TxID: txID1, RqTxID: txID3},
				{TxID: txID3, RqTxID: txID2},
				// Atomic 2
				{TxID: txID5},
				{TxID: txID4, RqTxID: txID5},
				// Atomic 3
				{TxID: txID7, RqTxID: txID6},
				{TxID: txID6, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID1, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID2},
					},
					// Atomic 2
					txID5: {
						{TxID: txID5},
						{TxID: txID4, RqTxID: txID5},
					},
					// Atomic 3
					txID7: {
						{TxID: txID7, RqTxID: txID6},
						{TxID: txID6, RqTxID: txID7},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs:   map[common.TxID]common.TxID{txID2: txID2, txID3: txID2, txID5: txID5, txID6: txID7, txID7: txID7},
			},
		}, {
			name: "test with non atomic transactions sorted",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, RqTxID: txID3},
				{TxID: txID1, RqTxID: txID3},
				{TxID: txID3, RqTxID: txID2},
				// Non-atomic
				{TxID: txID4},
				{TxID: txID5},
				{TxID: txID6},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID1, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID2},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs:   map[common.TxID]common.TxID{txID2: txID2, txID3: txID2},
			},
		}, {
			name: "test with non atomic transactions unsorted",
			poolTxs: []common.PoolL2Tx{
				// Non-atomic
				{TxID: txID5},
				// Atomic 1
				{TxID: txID2, RqTxID: txID3},
				// Non-atomic
				{TxID: txID4},
				// Atomic 1
				{TxID: txID1, RqTxID: txID3},
				// Atomic 1
				{TxID: txID3, RqTxID: txID2},
				// Non-atomic
				{TxID: txID6},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID1, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID2},
					},
				},
				discarded: map[common.TxID]bool{},
				usedTxs:   map[common.TxID]common.TxID{txID2: txID2, txID3: txID2},
			},
		}, {
			name:    "test a empty tx list",
			poolTxs: []common.PoolL2Tx{},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				discarded:       map[common.TxID]bool{},
				usedTxs:         map[common.TxID]common.TxID{},
			},
		}, {
			name: "test only non-atomic txs",
			poolTxs: []common.PoolL2Tx{
				// Non-atomics
				{TxID: txID1},
				{TxID: txID2},
				{TxID: txID3},
				{TxID: txID4},
				{TxID: txID5},
				{TxID: txID6},
				{TxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				discarded:       map[common.TxID]bool{},
				usedTxs:         map[common.TxID]common.TxID{},
			},
		}, {
			name: "test invalid atomic txs",
			poolTxs: []common.PoolL2Tx{
				// invalid atomics
				{TxID: txID1, RqTxID: txID2},
				{TxID: txID6, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{},
				discarded:       map[common.TxID]bool{txID1: true, txID6: true},
				usedTxs:         map[common.TxID]common.TxID{},
			},
		}, {
			name: "test invalid and valid atomic txs",
			poolTxs: []common.PoolL2Tx{
				// Atomic 1
				{TxID: txID2, RqTxID: txID3},
				{TxID: txID1, RqTxID: txID3},
				{TxID: txID3, RqTxID: txID2},
				// Non-atomic
				{TxID: txID4, RqTxID: txID5},
				{TxID: txID6, RqTxID: txID7},
			},
			want: want{
				txAtomicMapping: map[common.TxID][]common.PoolL2Tx{
					// Atomic 1
					txID2: {
						{TxID: txID2, RqTxID: txID3},
						{TxID: txID1, RqTxID: txID3},
						{TxID: txID3, RqTxID: txID2},
					},
				},
				discarded: map[common.TxID]bool{txID4: true, txID6: true},
				usedTxs:   map[common.TxID]common.TxID{txID2: txID2, txID3: txID2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txAtomicMapping, discarded, usedTxs := buildAtomicTxs(tt.poolTxs)
			assert.EqualValues(t, tt.want.txAtomicMapping, txAtomicMapping)
			assert.EqualValues(t, tt.want.discarded, discarded)
			assert.EqualValues(t, tt.want.usedTxs, usedTxs)
		})
	}
}

func TestTxBatch_getSelection(t *testing.T) {
	type want struct {
		coordIdxs      []common.Idx
		auths          [][]byte
		l1UserTxs      []common.L1Tx
		l1CoordTxs     []common.L1Tx
		poolL2Txs      []common.PoolL2Tx
		discardedL2Txs []common.PoolL2Tx
	}
	tests := []struct {
		name string
		txs  []*TxGroup
		want want
		err  error
	}{
		{
			name: "select idxs",
			txs:  []*TxGroup{{coordIdxsMap: map[common.TokenID]common.Idx{0: 10, 1: 20, 2: 30}}},
			want: want{
				coordIdxs:      []common.Idx{10, 20, 30},
				auths:          [][]byte{},
				l1UserTxs:      []common.L1Tx{},
				l1CoordTxs:     []common.L1Tx{},
				poolL2Txs:      []common.PoolL2Tx{},
				discardedL2Txs: []common.PoolL2Tx{},
			},
		}, {
			name: "select auths",
			txs:  []*TxGroup{{accAuths: [][]byte{{10}, {20}, {30}}}},
			want: want{
				coordIdxs:      []common.Idx{},
				auths:          [][]byte{{10}, {20}, {30}},
				l1UserTxs:      []common.L1Tx{},
				l1CoordTxs:     []common.L1Tx{},
				poolL2Txs:      []common.PoolL2Tx{},
				discardedL2Txs: []common.PoolL2Tx{},
			},
		}, {
			name: "select l1 user txs",
			txs:  []*TxGroup{{l1UserTxs: []common.L1Tx{{FromIdx: 10}}}},
			want: want{
				coordIdxs:      []common.Idx{},
				auths:          [][]byte{},
				l1UserTxs:      []common.L1Tx{{FromIdx: 10}},
				l1CoordTxs:     []common.L1Tx{},
				poolL2Txs:      []common.PoolL2Tx{},
				discardedL2Txs: []common.PoolL2Tx{},
			},
		}, {
			name: "select l1 coordinator txs",
			txs:  []*TxGroup{{l1CoordTxs: []common.L1Tx{{FromIdx: 10}}}},
			want: want{
				coordIdxs:      []common.Idx{},
				auths:          [][]byte{},
				l1UserTxs:      []common.L1Tx{},
				l1CoordTxs:     []common.L1Tx{{FromIdx: 10}},
				poolL2Txs:      []common.PoolL2Tx{},
				discardedL2Txs: []common.PoolL2Tx{},
			},
		}, {
			name: "select l2 pool txs",
			txs:  []*TxGroup{{l2Txs: []common.PoolL2Tx{{FromIdx: 10}}}},
			want: want{
				coordIdxs:      []common.Idx{},
				auths:          [][]byte{},
				l1UserTxs:      []common.L1Tx{},
				l1CoordTxs:     []common.L1Tx{},
				poolL2Txs:      []common.PoolL2Tx{{FromIdx: 10}},
				discardedL2Txs: []common.PoolL2Tx{},
			},
		}, {
			name: "select l2 discarded txs",
			txs:  []*TxGroup{{discardedTxs: []common.PoolL2Tx{{FromIdx: 10}}}},
			want: want{
				coordIdxs:      []common.Idx{},
				auths:          [][]byte{},
				l1UserTxs:      []common.L1Tx{},
				l1CoordTxs:     []common.L1Tx{},
				poolL2Txs:      []common.PoolL2Tx{},
				discardedL2Txs: []common.PoolL2Tx{{FromIdx: 10}},
			},
		}, {
			name: "select all selection",
			txs: []*TxGroup{{
				coordIdxsMap: map[common.TokenID]common.Idx{0: 10, 1: 20, 2: 30},
				accAuths:     [][]byte{{10}, {20}, {30}},
				l1UserTxs:    []common.L1Tx{{FromIdx: 10}},
				l1CoordTxs:   []common.L1Tx{{FromIdx: 10}},
				l2Txs:        []common.PoolL2Tx{{FromIdx: 10}},
				discardedTxs: []common.PoolL2Tx{{FromIdx: 10}},
			}},
			want: want{
				coordIdxs:      []common.Idx{10, 20, 30},
				auths:          [][]byte{{10}, {20}, {30}},
				l1UserTxs:      []common.L1Tx{{FromIdx: 10}},
				l1CoordTxs:     []common.L1Tx{{FromIdx: 10}},
				poolL2Txs:      []common.PoolL2Tx{{FromIdx: 10}},
				discardedL2Txs: []common.PoolL2Tx{{FromIdx: 10}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &TxBatch{txs: tt.txs}
			got1, got2, got3, got4, got5, got6, err := b.getSelection()
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
			assert.Equal(t, tt.want.coordIdxs, got1)
			assert.Equal(t, tt.want.auths, got2)
			assert.Equal(t, tt.want.l1UserTxs, got3)
			assert.Equal(t, tt.want.l1CoordTxs, got4)
			assert.Equal(t, tt.want.poolL2Txs, got5)
			assert.Equal(t, tt.want.discardedL2Txs, got6)
		})
	}
}

func TestTxBatch_last(t *testing.T) {
	tests := []struct {
		name string
		txs  []*TxGroup
		want *TxGroup
	}{
		{
			name: "test 1",
			txs:  []*TxGroup{{firstPosition: 0}},
			want: &TxGroup{firstPosition: 0},
		}, {
			name: "test 2",
			txs:  []*TxGroup{{firstPosition: 0}, {firstPosition: 1}},
			want: &TxGroup{firstPosition: 1},
		}, {
			name: "test 3",
			txs:  []*TxGroup{{firstPosition: 0}, {firstPosition: 1}, {firstPosition: 3}},
			want: &TxGroup{firstPosition: 3},
		}, {
			name: "test empty",
			txs:  []*TxGroup{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txBatch := &TxBatch{txs: tt.txs}
			got := txBatch.last()
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestTxBatch_length(t *testing.T) {
	type args struct {
		txs       []*TxGroup
		l1UserTxs []common.L1Tx
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test empty",
			args: args{
				txs:       []*TxGroup{},
				l1UserTxs: []common.L1Tx{},
			},
			want: 0,
		}, {
			name: "one l2",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{{}}}},
				l1UserTxs: []common.L1Tx{},
			},
			want: 1,
		}, {
			name: "two l2",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{{}, {}}}},
				l1UserTxs: []common.L1Tx{},
			},
			want: 2,
		}, {
			name: "one l1",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{}}},
				l1UserTxs: []common.L1Tx{{}},
			},
			want: 1,
		}, {
			name: "two l1",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{}}},
				l1UserTxs: []common.L1Tx{{}, {}},
			},
			want: 2,
		}, {
			name: "two l2 and one l1",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{{}, {}}}},
				l1UserTxs: []common.L1Tx{},
			},
			want: 2,
		}, {
			name: "two l2 and one l1",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{{}, {}}}},
				l1UserTxs: []common.L1Tx{{}},
			},
			want: 3,
		}, {
			name: "two l2 and two l1",
			args: args{
				txs:       []*TxGroup{{l2Txs: []common.PoolL2Tx{{}, {}}}},
				l1UserTxs: []common.L1Tx{{}, {}},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txBatch := &TxBatch{l1UserTxs: tt.args.l1UserTxs, txs: tt.args.txs}
			got := txBatch.length()
			assert.Equal(t, tt.want, got)
		})
	}
}
