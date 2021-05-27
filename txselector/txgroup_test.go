package txselector

import (
	"fmt"
	"math/big"
	"sort"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxGroup_addPoolTxs(t *testing.T) {
	type args struct {
		atomic bool
		l1txs  []common.L1Tx
		l2Txs  []common.PoolL2Tx
	}
	type want struct {
		l1CoordTxs   []common.L1Tx
		l1UserTxs    []common.L1Tx
		l2Txs        []common.PoolL2Tx
		discardedTxs []common.PoolL2Tx
		coordIdxs    map[common.TokenID]common.Idx
		err          error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty txs",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{},
			},
			want: want{
				err: fmt.Errorf("empty txs"),
			},
		}, {
			name: "invalid tx without type",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2},
				},
			},
			want: want{
				err: fmt.Errorf("invalid tx (0x020000000000000000000000000000000000000000000000000000000000000001) type "),
			},
		}, {
			name: "invalid transfer",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID1, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2,
						Type: common.TxTypeTransfer, Info: ErrRecipientNotFound},
				},
				coordIdxs:  map[common.TokenID]common.Idx{},
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs:  []common.L1Tx{},
			},
		}, {
			name: "valid transfer to bjj",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		}, {
			name: "invalid transfer to bjj",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 34,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 34,
						Info: statedb.ErrIdxNotFound.Error(), Amount: big.NewInt(10), Fee: 33, Nonce: 1,
						Type: common.TxTypeTransferToBJJ},
				},
				coordIdxs:  map[common.TokenID]common.Idx{},
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   common.FFAddr,
						FromBJJ:       _bjj1,
						TokenID:       34,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				err: nil,
			},
		}, {
			name: "invalid transfer to bjj without coordinator idx",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ}},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, Fee: 33, Nonce: 1,
						Info: ErrCoordIdxNotFound, Amount: big.NewInt(10), Type: common.TxTypeTransferToBJJ},
				},
				coordIdxs: map[common.TokenID]common.Idx{},
				l1CoordTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _coordAccount.Addr,
						FromBJJ:       _coordAccount.BJJ,
						TokenID:       35,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				l1UserTxs: []common.L1Tx{},
				err:       nil,
			},
		}, {
			name: "valid transfer to eth",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		}, {
			name: "valid transfer",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 1,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 1,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		}, {
			name: "multiple transfers",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		}, {
			name: "maximum transfers",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
					{TxID: txID7, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35,
						Amount: big.NewInt(10), Fee: 33, Nonce: 7, Type: common.TxTypeTransferToBJJ},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID7, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, Fee: 33, Nonce: 7,
						Info: ErrCoordIdxNotFound, Amount: big.NewInt(10), Type: common.TxTypeTransferToBJJ},
				},
				coordIdxs: map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _coordAccount.Addr,
						FromBJJ:       _coordAccount.BJJ,
						TokenID:       35,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				l1UserTxs: []common.L1Tx{},
				err:       nil,
			},
		}, {
			name: "no balance transfer besides position",
			args: args{
				atomic: true,
				l1txs:  []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
					{TxID: txID2, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID2, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		}, {
			name: "balance transfer besides position",
			args: args{
				atomic: true,
				l1txs:  []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID2, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID2, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxs:    map[common.TokenID]common.Idx{0: 350},
				l1CoordTxs:   []common.L1Tx{},
				l1UserTxs:    []common.L1Tx{},
				err:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{coordAccount: _coordAccount}
			db := createStateDbMock()
			err := txGroup.addPoolTxs(tt.args.atomic, tt.args.l2Txs, &txProcessorMock{db: db},
				createL2DBMock(), db, tt.args.l1txs)
			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
			assert.EqualValues(t, tt.want.l1CoordTxs, txGroup.l1CoordTxs)
			assert.EqualValues(t, tt.want.l1UserTxs, txGroup.l1UserTxs)
			assert.EqualValues(t, tt.want.l2Txs, txGroup.l2Txs)
			assert.EqualValues(t, tt.want.discardedTxs, txGroup.discardedTxs)
			assert.EqualValues(t, tt.want.coordIdxs, txGroup.coordIdxsMap)
		})
	}
}

func TestTxGroup_validate(t *testing.T) {
	type args struct {
		atomic bool
		l1txs  []common.L1Tx
		l2Txs  []common.PoolL2Tx
	}
	type want struct {
		l2Txs        []common.PoolL2Tx
		discardedTxs []common.PoolL2Tx
		coordIdxsMap map[common.TokenID]common.Idx
		err          error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "invalid tx without type",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2},
				},
			},
			want: want{
				err: fmt.Errorf("invalid tx (0x020000000000000000000000000000000000000000000000000000000000000001) type "),
			},
		}, {
			name: "error exit transaction with wrong toIdx",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{Type: common.TxTypeExit, Amount: big.NewInt(0), ToIdx: 420},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{Type: common.TxTypeExit, Amount: big.NewInt(0), ToIdx: 420, Info: ErrInvalidExitToIdx},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{},
			},
		}, {
			name: "error exit transaction with zero amount",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{Type: common.TxTypeExit, Amount: big.NewInt(0), ToIdx: 1},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{Type: common.TxTypeExit, Amount: big.NewInt(0), ToIdx: 1, Info: ErrExitZeroAmount},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{},
			},
		}, {
			name: "invalid transfer",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID1, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID2,
						Type: common.TxTypeTransfer, Info: ErrRecipientNotFound},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{},
			},
		}, {
			name: "valid transfer to bjj",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "invalid transfer to bjj without coordinator idx",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ}},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, Fee: 33, Nonce: 1,
						Info: ErrCoordIdxNotFound, Amount: big.NewInt(10), Type: common.TxTypeTransferToBJJ},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{},
				err:          nil,
			},
		}, {
			name: "valid transfer to eth",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "valid transfer",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 1,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 1,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "multiple transfers",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "maximum transfers",
			args: args{
				l1txs: []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
					{TxID: txID7, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35,
						Amount: big.NewInt(10), Fee: 33, Nonce: 7, Type: common.TxTypeTransferToBJJ},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID2, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 2, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID3, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 3,
						Type: common.TxTypeTransfer},
					{TxID: txID4, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 4, Type: common.TxTypeTransferToBJJ},
					{TxID: txID5, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0,
						Amount: big.NewInt(10), Fee: 33, Nonce: 5, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), Fee: 33, Nonce: 6,
						Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID7, FromIdx: 350, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, Fee: 33, Nonce: 7,
						Info: ErrCoordIdxNotFound, Amount: big.NewInt(10), Type: common.TxTypeTransferToBJJ},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "fail atomic transactions",
			args: args{
				atomic: true,
				l1txs:  []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0, RqTxID: txID2,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr},
					{TxID: txID2, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), RqTxID: txID3,
						Fee: 33, Nonce: 2, Type: common.TxTypeTransfer},
					{TxID: txID3, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, RqTxID: txID4,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ},
					{TxID: txID4, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID1,
						Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{},
				discardedTxs: []common.PoolL2Tx{
					{TxID: txID3, FromIdx: 349, ToEthAddr: common.FFAddr, ToBJJ: _bjj1, TokenID: 35, RqTxID: txID4,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToBJJ,
						Info: ErrCoordIdxNotFound},
					{TxID: txID4, ToIdx: 444, TokenID: 0, Amount: big.NewInt(1000), Fee: 33, RqTxID: txID1,
						Type: common.TxTypeTransfer, Info: ErrRecipientNotFound},
					{TxID: txID1, FromIdx: 350, ToEthAddr: _ethAddr2, ToBJJ: _bjj1, TokenID: 0, RqTxID: txID2,
						Amount: big.NewInt(10), Fee: 33, Nonce: 1, Type: common.TxTypeTransferToEthAddr,
						Info: ErrAtomicGroupFail},
					{TxID: txID2, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(10), RqTxID: txID3,
						Fee: 33, Nonce: 2, Type: common.TxTypeTransfer, Info: ErrAtomicGroupFail},
				},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "no balance transfer besides position",
			args: args{
				atomic: true,
				l1txs:  []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
					{TxID: txID2, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID2, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID1, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		}, {
			name: "balance transfer besides position",
			args: args{
				atomic: true,
				l1txs:  []common.L1Tx{},
				l2Txs: []common.PoolL2Tx{
					{TxID: txID5, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{
					{TxID: txID5, FromIdx: 352, ToIdx: 350, TokenID: 0, Amount: big.NewInt(999999), Fee: 22,
						Nonce: 33, Type: common.TxTypeTransfer},
					{TxID: txID6, FromIdx: 350, ToIdx: 351, TokenID: 0, Amount: big.NewInt(999900), Fee: 10,
						Nonce: 1, Type: common.TxTypeTransfer},
				},
				discardedTxs: []common.PoolL2Tx{},
				coordIdxsMap: map[common.TokenID]common.Idx{0: 350},
				err:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{
				atomic:       tt.args.atomic,
				l2Txs:        tt.args.l2Txs,
				coordAccount: _coordAccount,
			}
			txGroup.calcFeeAverage()
			txGroup.sort()
			db := createStateDbMock()
			coordIdxsMap, err := txGroup.validate(&txProcessorMock{db: db}, createL2DBMock(), db, tt.args.l1txs)
			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
			assert.EqualValues(t, tt.want.coordIdxsMap, coordIdxsMap)
			assert.EqualValues(t, tt.want.l2Txs, txGroup.l2Txs)
			assert.EqualValues(t, tt.want.discardedTxs, txGroup.discardedTxs)
		})
	}
}

func TestTxGroup_distributeFee(t *testing.T) {
	tests := []struct {
		name         string
		coordIdxsMap map[common.TokenID]common.Idx
		err          error
	}{
		{
			name:         "test nil map",
			coordIdxsMap: nil,
			err:          nil,
		}, {
			name:         "test zero idx",
			coordIdxsMap: map[common.TokenID]common.Idx{},
			err:          nil,
		}, {
			name:         "test one idx",
			coordIdxsMap: map[common.TokenID]common.Idx{10: 20},
			err:          nil,
		}, {
			name:         "test two idx",
			coordIdxsMap: map[common.TokenID]common.Idx{10: 20, 11: 21},
			err:          nil,
		}, {
			name:         "test multiple idx",
			coordIdxsMap: map[common.TokenID]common.Idx{10: 20, 11: 21, 12: 22, 13: 23, 14: 24},
			err:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{}
			err := txGroup.distributeFee(tt.coordIdxsMap, createTxProcessorMock(), createStateDbMock())
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
		})
	}
}

func TestTxGroup_createL1Txs(t *testing.T) {
	type args struct {
		discardedTxs []common.PoolL2Tx

		l2Txs []common.PoolL2Tx
		l1Txs []common.L1Tx
	}
	type want struct {
		l1CoordTxs []common.L1Tx
		l1UserTxs  []common.L1Tx
		accAuths   [][]byte
		err        error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty result",
			args: args{l2Txs: []common.PoolL2Tx{}, l1Txs: []common.L1Tx{}},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs:  []common.L1Tx{},
				accAuths:   [][]byte{},
				err:        nil,
			},
		}, {
			name: "create coordinator account tx",
			args: args{
				l2Txs:        []common.PoolL2Tx{{TokenID: 33, Amount: big.NewInt(1000), Fee: 33}},
				discardedTxs: []common.PoolL2Tx{},
				l1Txs:        []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _coordAccount.Addr,
						FromBJJ:       _coordAccount.BJJ,
						TokenID:       33,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				l1UserTxs: []common.L1Tx{},
				accAuths:  [][]byte{{48, 120, 53, 56, 51, 56, 52, 102, 100, 100, 100, 101, 56, 49, 98, 99, 100, 101, 56, 97, 56, 102, 49, 57, 53, 49, 97, 49, 48, 49, 102, 48, 54, 98, 53, 102, 102, 50, 102, 50, 100, 101, 54, 49, 51, 49, 56, 101, 52, 50, 49, 100, 102, 53, 52, 56, 100, 57, 51, 97, 56, 97, 48, 99, 49, 97, 51, 48, 57, 50, 49, 53, 52, 102, 52, 101, 56, 97, 50, 98, 99, 101, 97, 97, 98, 98, 54, 48, 51, 57, 51, 56, 48, 100, 52, 54, 100, 55, 52, 99, 97, 98, 98, 53, 51, 57, 48, 99, 54, 57, 57, 97, 53, 101, 100, 56, 54, 52, 100, 56, 48, 101, 50, 48, 55, 49, 100, 57, 57, 49, 49, 99}},
				err:       nil,
			},
		}, {
			name: "create coordinator account tx twice",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{TokenID: 33, Amount: big.NewInt(1000), Fee: 33},
					{TokenID: 33, Amount: big.NewInt(1000), Fee: 33},
				},
				discardedTxs: []common.PoolL2Tx{},
				l1Txs:        []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _coordAccount.Addr,
						FromBJJ:       _coordAccount.BJJ,
						TokenID:       33,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				l1UserTxs: []common.L1Tx{},
				accAuths:  [][]byte{{48, 120, 53, 56, 51, 56, 52, 102, 100, 100, 100, 101, 56, 49, 98, 99, 100, 101, 56, 97, 56, 102, 49, 57, 53, 49, 97, 49, 48, 49, 102, 48, 54, 98, 53, 102, 102, 50, 102, 50, 100, 101, 54, 49, 51, 49, 56, 101, 52, 50, 49, 100, 102, 53, 52, 56, 100, 57, 51, 97, 56, 97, 48, 99, 49, 97, 51, 48, 57, 50, 49, 53, 52, 102, 52, 101, 56, 97, 50, 98, 99, 101, 97, 97, 98, 98, 54, 48, 51, 57, 51, 56, 48, 100, 52, 54, 100, 55, 52, 99, 97, 98, 98, 53, 51, 57, 48, 99, 54, 57, 57, 97, 53, 101, 100, 56, 54, 52, 100, 56, 48, 101, 50, 48, 55, 49, 100, 57, 57, 49, 49, 99}},
				err:       nil,
			},
		}, {
			name: "create duplicated coordinator account tx",
			args: args{
				l2Txs: []common.PoolL2Tx{{TokenID: 33, Amount: big.NewInt(1000), Fee: 33}},
				l1Txs: []common.L1Tx{{FromEthAddr: _coordAccount.Addr, FromBJJ: _coordAccount.BJJ, TokenID: 33}},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs:  []common.L1Tx{},
				accAuths:   [][]byte{},
				err:        nil,
			},
		}, {
			name: "create TransferToEthAddr account tx",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{TokenID: 0, Amount: big.NewInt(1), Fee: 10, ToIdx: 0, ToEthAddr: _invalidEthAddr1,
						FromIdx: 351, Nonce: 10},
				},
				l1Txs: []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs: []common.L1Tx{{
					UserOrigin:    false,
					FromEthAddr:   _invalidEthAddr1,
					FromBJJ:       _bjj1,
					TokenID:       0,
					Amount:        big.NewInt(0),
					DepositAmount: big.NewInt(0),
					Position:      0,
					Type:          common.TxTypeCreateAccountDeposit,
				}},
				accAuths: [][]byte{{48, 120, 54, 100, 97, 101, 55, 99, 102, 98, 98, 54, 102, 99, 99, 53, 56, 48, 97, 99, 54, 102, 50, 101, 97, 101, 50, 101, 97, 102, 50, 100, 99, 97, 57, 49, 100, 53, 53, 97, 52, 49, 51, 57, 101, 54, 55, 56, 54, 53, 50, 52, 57, 100, 97, 53, 52, 54, 52, 53, 56, 100, 49, 55, 54, 55, 52, 102, 100, 101, 57, 97, 52, 56, 48, 102, 99, 50, 57, 53, 56, 52, 51, 56, 52, 98, 97, 98, 49, 101, 48, 100, 97, 102, 53, 101, 101, 49, 101, 99, 52, 53, 52, 97, 101, 50, 50, 102, 97, 50, 54, 52, 48, 52, 54, 48, 102, 99, 51, 54, 98, 57, 99, 51, 97, 57, 54, 57, 56, 55, 49, 98}},
				err:      nil,
			},
		}, {
			name: "create TransferToEthAddr account tx twice",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{TokenID: 0, Amount: big.NewInt(1), Fee: 10, ToIdx: 0, ToEthAddr: _invalidEthAddr1,
						FromIdx: 351, Nonce: 10},
					{TokenID: 0, Amount: big.NewInt(1), Fee: 10, ToIdx: 0, ToEthAddr: _invalidEthAddr1,
						FromIdx: 351, Nonce: 10},
				},
				l1Txs: []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs: []common.L1Tx{{
					UserOrigin:    false,
					FromEthAddr:   _invalidEthAddr1,
					FromBJJ:       _bjj1,
					TokenID:       0,
					Amount:        big.NewInt(0),
					DepositAmount: big.NewInt(0),
					Position:      0,
					Type:          common.TxTypeCreateAccountDeposit,
				}},
				accAuths: [][]byte{{48, 120, 54, 100, 97, 101, 55, 99, 102, 98, 98, 54, 102, 99, 99, 53, 56, 48, 97, 99, 54, 102, 50, 101, 97, 101, 50, 101, 97, 102, 50, 100, 99, 97, 57, 49, 100, 53, 53, 97, 52, 49, 51, 57, 101, 54, 55, 56, 54, 53, 50, 52, 57, 100, 97, 53, 52, 54, 52, 53, 56, 100, 49, 55, 54, 55, 52, 102, 100, 101, 57, 97, 52, 56, 48, 102, 99, 50, 57, 53, 56, 52, 51, 56, 52, 98, 97, 98, 49, 101, 48, 100, 97, 102, 53, 101, 101, 49, 101, 99, 52, 53, 52, 97, 101, 50, 50, 102, 97, 50, 54, 52, 48, 52, 54, 48, 102, 99, 51, 54, 98, 57, 99, 51, 97, 57, 54, 57, 56, 55, 49, 98}},
				err:      nil,
			},
		}, {
			name: "create duplicated TransferToEthAddr account tx",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{TokenID: 0, Amount: big.NewInt(1), Fee: 10, ToIdx: 0, ToEthAddr: _ethAddr3, FromIdx: 351, Nonce: 10},
				},
				l1Txs: []common.L1Tx{
					{FromEthAddr: _ethAddr3, TokenID: 0},
				},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs:  []common.L1Tx{},
				accAuths:   [][]byte{},
				err:        nil,
			},
		}, {
			name: "create TransferToBJJ account tx",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{
						TokenID:   34,
						Amount:    big.NewInt(1),
						Fee:       10,
						ToIdx:     0,
						ToEthAddr: common.FFAddr,
						ToBJJ:     _bjj1,
						FromIdx:   351,
						Nonce:     10,
					},
				},
				l1Txs: []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs: []common.L1Tx{{
					UserOrigin:    false,
					FromEthAddr:   common.FFAddr,
					FromBJJ:       _bjj1,
					TokenID:       34,
					Amount:        big.NewInt(0),
					DepositAmount: big.NewInt(0),
					Position:      0,
					Type:          common.TxTypeCreateAccountDeposit,
				}},
				accAuths: [][]byte{common.EmptyEthSignature},
				err:      nil,
			},
		}, {
			name: "create TransferToBJJ account tx twice",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{
						TokenID:   34,
						Amount:    big.NewInt(1),
						Fee:       10,
						ToIdx:     0,
						ToEthAddr: common.FFAddr,
						ToBJJ:     _bjj1,
						FromIdx:   351,
						Nonce:     10,
					}, {
						TokenID:   34,
						Amount:    big.NewInt(1),
						Fee:       10,
						ToIdx:     0,
						ToEthAddr: common.FFAddr,
						ToBJJ:     _bjj1,
						FromIdx:   351,
						Nonce:     10,
					},
				},
				l1Txs: []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs: []common.L1Tx{{
					UserOrigin:    false,
					FromEthAddr:   common.FFAddr,
					FromBJJ:       _bjj1,
					TokenID:       34,
					Amount:        big.NewInt(0),
					DepositAmount: big.NewInt(0),
					Position:      0,
					Type:          common.TxTypeCreateAccountDeposit,
				}},
				accAuths: [][]byte{common.EmptyEthSignature},
				err:      nil,
			},
		}, {
			name: "create duplicated TransferToBJJ account tx",
			args: args{
				l2Txs: []common.PoolL2Tx{
					{
						TokenID:   0,
						Amount:    big.NewInt(1),
						Fee:       10,
						ToIdx:     0,
						ToEthAddr: common.FFAddr,
						ToBJJ:     _bjj1,
						FromIdx:   351,
						Nonce:     10,
					},
				},
				l1Txs: []common.L1Tx{
					{FromEthAddr: common.FFAddr, FromBJJ: _bjj1, TokenID: 0},
				},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{},
				l1UserTxs:  []common.L1Tx{},
				accAuths:   [][]byte{},
				err:        nil,
			},
		}, {
			name: "create multiple txs",
			args: args{
				discardedTxs: []common.PoolL2Tx{},
				l2Txs: []common.PoolL2Tx{
					{TokenID: 33, Amount: big.NewInt(1000), Fee: 33},
					{TokenID: 33, Amount: big.NewInt(1000), Fee: 33},
					{TokenID: 0, Amount: big.NewInt(1), Fee: 10, ToIdx: 0, ToEthAddr: _invalidEthAddr1,
						FromIdx: 351, Nonce: 10},
					{
						TokenID:   34,
						Amount:    big.NewInt(1),
						Fee:       10,
						ToIdx:     0,
						ToEthAddr: common.FFAddr,
						ToBJJ:     _bjj1,
						FromIdx:   351,
						Nonce:     10,
					},
				},
				l1Txs: []common.L1Tx{},
			},
			want: want{
				l1CoordTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _coordAccount.Addr,
						FromBJJ:       _coordAccount.BJJ,
						TokenID:       33,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      0,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				l1UserTxs: []common.L1Tx{
					{
						UserOrigin:    false,
						FromEthAddr:   _invalidEthAddr1,
						FromBJJ:       _bjj1,
						TokenID:       0,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      1,
						Type:          common.TxTypeCreateAccountDeposit,
					}, {
						UserOrigin:    false,
						FromEthAddr:   common.FFAddr,
						FromBJJ:       _bjj1,
						TokenID:       34,
						Amount:        big.NewInt(0),
						DepositAmount: big.NewInt(0),
						Position:      2,
						Type:          common.TxTypeCreateAccountDeposit,
					},
				},
				accAuths: [][]byte{
					{48, 120, 53, 56, 51, 56, 52, 102, 100, 100, 100, 101, 56, 49, 98, 99, 100, 101, 56, 97, 56, 102, 49, 57, 53, 49, 97, 49, 48, 49, 102, 48, 54, 98, 53, 102, 102, 50, 102, 50, 100, 101, 54, 49, 51, 49, 56, 101, 52, 50, 49, 100, 102, 53, 52, 56, 100, 57, 51, 97, 56, 97, 48, 99, 49, 97, 51, 48, 57, 50, 49, 53, 52, 102, 52, 101, 56, 97, 50, 98, 99, 101, 97, 97, 98, 98, 54, 48, 51, 57, 51, 56, 48, 100, 52, 54, 100, 55, 52, 99, 97, 98, 98, 53, 51, 57, 48, 99, 54, 57, 57, 97, 53, 101, 100, 56, 54, 52, 100, 56, 48, 101, 50, 48, 55, 49, 100, 57, 57, 49, 49, 99},
					{48, 120, 54, 100, 97, 101, 55, 99, 102, 98, 98, 54, 102, 99, 99, 53, 56, 48, 97, 99, 54, 102, 50, 101, 97, 101, 50, 101, 97, 102, 50, 100, 99, 97, 57, 49, 100, 53, 53, 97, 52, 49, 51, 57, 101, 54, 55, 56, 54, 53, 50, 52, 57, 100, 97, 53, 52, 54, 52, 53, 56, 100, 49, 55, 54, 55, 52, 102, 100, 101, 57, 97, 52, 56, 48, 102, 99, 50, 57, 53, 56, 52, 51, 56, 52, 98, 97, 98, 49, 101, 48, 100, 97, 102, 53, 101, 101, 49, 101, 99, 52, 53, 52, 97, 101, 50, 50, 102, 97, 50, 54, 52, 48, 52, 54, 48, 102, 99, 51, 54, 98, 57, 99, 51, 97, 57, 54, 57, 56, 55, 49, 98},
					common.EmptyEthSignature,
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{
				l2Txs:        tt.args.l2Txs,
				discardedTxs: tt.args.discardedTxs,
				coordAccount: _coordAccount,
			}
			err := txGroup.createL1Txs(createTxProcessorMock(), createL2DBMock(), createStateDbMock(), tt.args.l1Txs)
			if tt.want.err != nil {
				assert.Equal(t, tt.want.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
			assert.EqualValues(t, tt.want.l1CoordTxs, txGroup.l1CoordTxs)
			assert.EqualValues(t, tt.want.l1UserTxs, txGroup.l1UserTxs)
			assert.EqualValues(t, tt.want.accAuths, txGroup.accAuths)
		})
	}
}

func TestTxGroup_sort(t *testing.T) {
	type args struct {
		atomic bool
		l2Txs  []common.PoolL2Tx
	}
	tests := []struct {
		name string
		args args
		want []common.PoolL2Tx
	}{
		{
			name: "sort non-atomic txs 1",
			args: args{
				atomic: false,
				l2Txs: []common.PoolL2Tx{
					{AbsoluteFee: 2, Nonce: 12},
					{AbsoluteFee: 4, Nonce: 14},
					{AbsoluteFee: 3, Nonce: 13},
					{AbsoluteFee: 5, Nonce: 15},
					{AbsoluteFee: 1, Nonce: 11},
				},
			},
			want: []common.PoolL2Tx{
				{AbsoluteFee: 1, Nonce: 11},
				{AbsoluteFee: 2, Nonce: 12},
				{AbsoluteFee: 3, Nonce: 13},
				{AbsoluteFee: 4, Nonce: 14},
				{AbsoluteFee: 5, Nonce: 15},
			},
		}, {
			name: "sort non-atomic txs 2",
			args: args{
				atomic: false,
				l2Txs: []common.PoolL2Tx{
					{AbsoluteFee: 3, Nonce: 13},
					{AbsoluteFee: 5, Nonce: 11},
					{AbsoluteFee: 2, Nonce: 14},
					{AbsoluteFee: 1, Nonce: 15},
					{AbsoluteFee: 4, Nonce: 12},
				},
			},
			want: []common.PoolL2Tx{
				{AbsoluteFee: 5, Nonce: 11},
				{AbsoluteFee: 4, Nonce: 12},
				{AbsoluteFee: 3, Nonce: 13},
				{AbsoluteFee: 2, Nonce: 14},
				{AbsoluteFee: 1, Nonce: 15},
			},
		}, {
			name: "sort non-atomic txs 3",
			args: args{
				atomic: false,
				l2Txs: []common.PoolL2Tx{
					{AbsoluteFee: 1, Nonce: 13},
					{AbsoluteFee: 1, Nonce: 11},
					{AbsoluteFee: 11, Nonce: 13},
					{AbsoluteFee: 2, Nonce: 12},
					{AbsoluteFee: 200, Nonce: 12},
				},
			},
			want: []common.PoolL2Tx{
				{AbsoluteFee: 1, Nonce: 11},
				{AbsoluteFee: 200, Nonce: 12},
				{AbsoluteFee: 2, Nonce: 12},
				{AbsoluteFee: 11, Nonce: 13},
				{AbsoluteFee: 1, Nonce: 13},
			},
		}, {
			name: "sort atomic txs 1",
			args: args{
				atomic: true,
				l2Txs: []common.PoolL2Tx{
					{AbsoluteFee: 2, Nonce: 1},
					{AbsoluteFee: 3, Nonce: 2},
					{AbsoluteFee: 1, Nonce: 10},
					{AbsoluteFee: 4, Nonce: 33},
					{AbsoluteFee: 5, Nonce: 4},
				},
			},
			want: []common.PoolL2Tx{
				{AbsoluteFee: 5, Nonce: 4},
				{AbsoluteFee: 4, Nonce: 33},
				{AbsoluteFee: 3, Nonce: 2},
				{AbsoluteFee: 2, Nonce: 1},
				{AbsoluteFee: 1, Nonce: 10},
			},
		}, {
			name: "sort atomic txs 2",
			args: args{
				atomic: true,
				l2Txs: []common.PoolL2Tx{
					{AbsoluteFee: 3, Nonce: 3},
					{AbsoluteFee: 1, Nonce: 5},
					{AbsoluteFee: 2, Nonce: 4},
					{AbsoluteFee: 4, Nonce: 2},
					{AbsoluteFee: 5, Nonce: 1},
				},
			},
			want: []common.PoolL2Tx{
				{AbsoluteFee: 5, Nonce: 1},
				{AbsoluteFee: 4, Nonce: 2},
				{AbsoluteFee: 3, Nonce: 3},
				{AbsoluteFee: 2, Nonce: 4},
				{AbsoluteFee: 1, Nonce: 5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{l2Txs: tt.args.l2Txs, atomic: tt.args.atomic}
			txGroup.sort()
			isSorted := false
			if txGroup.atomic {
				isSorted = sort.SliceIsSorted(txGroup.l2Txs, func(i, j int) bool {
					return txGroup.l2Txs[i].AbsoluteFee > txGroup.l2Txs[j].AbsoluteFee
				})
			} else {
				isSorted = sort.SliceIsSorted(txGroup.l2Txs, func(i, j int) bool {
					return txGroup.l2Txs[i].Nonce < txGroup.l2Txs[j].Nonce
				})
			}
			assert.True(t, isSorted)
			assert.EqualValues(t, tt.want, txGroup.l2Txs)
		})
	}
}

func TestTxGroup_addL1Positions(t *testing.T) {
	type args struct {
		firstPosition int
		l1CoordTxs    []common.L1Tx
		l1UserTxs     []common.L1Tx
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "start from zero",
			args: args{
				firstPosition: 0,
				l1CoordTxs:    []common.L1Tx{{}, {}, {}, {}},
				l1UserTxs:     []common.L1Tx{{}, {}, {}},
			},
		}, {
			name: "start from ten",
			args: args{
				firstPosition: 10,
				l1CoordTxs:    []common.L1Tx{{}, {}, {}, {}},
				l1UserTxs:     []common.L1Tx{{}, {}, {}},
			},
		}, {
			name: "start from nine thousand and nine",
			args: args{
				firstPosition: 9009,
				l1CoordTxs:    []common.L1Tx{{}, {}, {}, {}},
				l1UserTxs:     []common.L1Tx{{}, {}, {}},
			},
		}, {
			name: "start from a negative number",
			args: args{
				firstPosition: -100,
				l1CoordTxs:    []common.L1Tx{{}, {}, {}, {}},
				l1UserTxs:     []common.L1Tx{{}, {}, {}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{
				firstPosition: tt.args.firstPosition,
				l1CoordTxs:    tt.args.l1CoordTxs,
				l1UserTxs:     tt.args.l1UserTxs,
			}
			txGroup.addL1Positions()
			firstPosition := tt.args.firstPosition
			for i, l1Tx := range append(txGroup.l1CoordTxs, txGroup.l1UserTxs...) {
				assert.Equal(t, firstPosition+i, l1Tx.Position)
			}
		})
	}
}

func TestTxGroup_calcFeeAverage(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  *big.Float
	}{
		{
			name:  "test empty",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{}},
			want:  big.NewFloat(0),
		}, {
			name:  "test nil",
			group: &TxGroup{},
			want:  big.NewFloat(0),
		}, {
			name:  "test zero",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 0}, {AbsoluteFee: 0}}},
			want:  big.NewFloat(0),
		}, {
			name:  "test zero sum",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 0}, {AbsoluteFee: 1}}},
			want:  big.NewFloat(0.5),
		}, {
			name:  "test one tx",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 02.33}}},
			want:  big.NewFloat(2.33),
		}, {
			name:  "test two txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 10.33}, {AbsoluteFee: 1111}}},
			want:  big.NewFloat(560.665),
		}, {
			name: "test many txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{
				{AbsoluteFee: 10.111}, {AbsoluteFee: 1111}, {AbsoluteFee: 10.333333333}, {AbsoluteFee: 03.1},
			}},
			want: big.NewFloat(283.63608333325),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.group.calcFeeAverage()
			assert.Equal(t, tt.want, tt.group.feeAverage)
		})
	}
}

func TestTxGroup_feeSum(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  *big.Float
	}{
		{
			name:  "test empty",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{}},
			want:  big.NewFloat(0),
		}, {
			name:  "test nil",
			group: &TxGroup{},
			want:  big.NewFloat(0),
		}, {
			name:  "test zero",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 0}, {AbsoluteFee: 0}}},
			want:  big.NewFloat(0),
		}, {
			name:  "test zero sum",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 0}, {AbsoluteFee: 1}}},
			want:  big.NewFloat(1),
		}, {
			name:  "test one tx",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 02.33}}},
			want:  big.NewFloat(2.33),
		}, {
			name:  "test two txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{AbsoluteFee: 10.33}, {AbsoluteFee: 1111}}},
			want:  big.NewFloat(1121.33),
		}, {
			name: "test many txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{
				{AbsoluteFee: 10.111}, {AbsoluteFee: 1111}, {AbsoluteFee: 10.333333333}, {AbsoluteFee: 03.1},
			}},
			want: big.NewFloat(1134.544333),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.feeSum()
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestTxGroup_isEmpty(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  bool
	}{
		{
			name:  "test empty",
			group: &TxGroup{},
			want:  true,
		}, {
			name:  "test l2Txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{}}},
			want:  false,
		}, {
			name:  "test l1UserTxs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}}},
			want:  false,
		}, {
			name:  "test l1CoordTxs",
			group: &TxGroup{l1CoordTxs: []common.L1Tx{{}}},
			want:  false,
		}, {
			name:  "test l1Txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}}, l1CoordTxs: []common.L1Tx{{}}},
			want:  false,
		}, {
			name:  "test all txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}}, l1CoordTxs: []common.L1Tx{{}}, l2Txs: []common.PoolL2Tx{{}}},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.isEmpty()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTxGroup_l2Length(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  int
	}{
		{
			name:  "test nil",
			group: &TxGroup{},
			want:  0,
		}, {
			name:  "test zero",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{}},
			want:  0,
		}, {
			name:  "test one",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{}}},
			want:  1,
		}, {
			name:  "test two",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{}, {}}},
			want:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.l2Length()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTxGroup_l1Length(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  int
	}{
		{
			name:  "test nil",
			group: &TxGroup{},
			want:  0,
		}, {
			name:  "test coordinator txs nil",
			group: &TxGroup{l1UserTxs: []common.L1Tx{}},
			want:  0,
		}, {
			name:  "test user txs nil",
			group: &TxGroup{l1CoordTxs: []common.L1Tx{}},
			want:  0,
		}, {
			name:  "test both empty",
			group: &TxGroup{l1UserTxs: []common.L1Tx{}, l1CoordTxs: []common.L1Tx{}},
			want:  0,
		}, {
			name:  "test one user tx",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}}, l1CoordTxs: []common.L1Tx{}},
			want:  1,
		}, {
			name:  "test two user txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}, {}}, l1CoordTxs: []common.L1Tx{}},
			want:  2,
		}, {
			name:  "test one coordinator tx",
			group: &TxGroup{l1UserTxs: []common.L1Tx{}, l1CoordTxs: []common.L1Tx{{}}},
			want:  1,
		}, {
			name:  "test two coordinator txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{}, l1CoordTxs: []common.L1Tx{{}, {}}},
			want:  2,
		}, {
			name:  "test four txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}, {}}, l1CoordTxs: []common.L1Tx{{}, {}}},
			want:  4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.l1Length()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTxGroup_length(t *testing.T) {
	tests := []struct {
		name  string
		group *TxGroup
		want  int
	}{
		{
			name:  "test nil",
			group: &TxGroup{},
			want:  0,
		}, {
			name:  "test zero txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{}, l1CoordTxs: []common.L1Tx{}, l2Txs: []common.PoolL2Tx{}},
			want:  0,
		}, {
			name:  "test four txs",
			group: &TxGroup{l2Txs: []common.PoolL2Tx{{}, {}, {}, {}}},
			want:  4,
		}, {
			name:  "test six txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}}, l1CoordTxs: []common.L1Tx{{}, {}, {}, {}, {}}},
			want:  6,
		}, {
			name:  "test four txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}, {}, {}}, l1CoordTxs: []common.L1Tx{}, l2Txs: []common.PoolL2Tx{{}}},
			want:  4,
		}, {
			name:  "test eight txs",
			group: &TxGroup{l1UserTxs: []common.L1Tx{{}, {}, {}}, l1CoordTxs: []common.L1Tx{{}}, l2Txs: []common.PoolL2Tx{{}, {}, {}, {}}},
			want:  8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.group.length()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTxGroup_popAllTxs(t *testing.T) {
	t.Run("test the popAllTxs method", func(t *testing.T) {
		txGroup := &TxGroup{
			l1UserTxs:     []common.L1Tx{{}, {}, {}, {}},
			l1CoordTxs:    []common.L1Tx{{}, {}, {}, {}},
			l2Txs:         []common.PoolL2Tx{{}, {}, {}, {}},
			discardedTxs:  []common.PoolL2Tx{{}, {}, {}, {}},
			coordIdxsMap:  map[common.TokenID]common.Idx{0: 10, 1: 200, 2: 20, 3: 2},
			accAuths:      [][]byte{{}, {}, {}, {}},
			feeAverage:    big.NewFloat(100),
			atomic:        true,
			firstPosition: 1000,
			coordAccount:  _coordAccount,
		}
		txGroup.popAllTxs()
		assert.EqualValues(t, []common.L1Tx{}, txGroup.l1UserTxs)
		assert.EqualValues(t, []common.L1Tx{}, txGroup.l1CoordTxs)
		assert.EqualValues(t, []common.PoolL2Tx{}, txGroup.l2Txs)
		assert.EqualValues(t, []common.PoolL2Tx{}, txGroup.discardedTxs)
		assert.EqualValues(t, map[common.TokenID]common.Idx{}, txGroup.coordIdxsMap)
		assert.EqualValues(t, [][]byte{}, txGroup.accAuths)
		assert.Equal(t, new(big.Float), txGroup.feeAverage)
		assert.Equal(t, false, txGroup.atomic)
		assert.Equal(t, 0, txGroup.firstPosition)
		assert.Equal(t, CoordAccount{}, txGroup.coordAccount)
	})
}

func TestTxGroup_popTx(t *testing.T) {
	type args struct {
		atomic bool
		l2Txs  []common.PoolL2Tx
		number int
	}
	type want struct {
		popAll bool
		length int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "pop one transaction",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 1,
			},
			want: want{
				popAll: false,
				length: 2,
			},
		}, {
			name: "pop two transaction",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 2,
			},
			want: want{
				popAll: false,
				length: 1,
			},
		}, {
			name: "pop all transaction",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 3,
			},
			want: want{
				popAll: true,
				length: 0,
			},
		}, {
			name: "pop greater than length",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 5,
			},
			want: want{
				popAll: true,
				length: 0,
			},
		}, {
			name: "pop zero",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 0,
			},
			want: want{
				popAll: false,
				length: 3,
			},
		}, {
			name: "pop atomic",
			args: args{
				atomic: true,
				l2Txs:  []common.PoolL2Tx{{}, {}, {}},
				number: 1,
			},
			want: want{
				popAll: true,
				length: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{atomic: tt.args.atomic, l2Txs: tt.args.l2Txs}
			popAll := txGroup.popTx(tt.args.number)
			assert.Equal(t, tt.want.length, txGroup.l2Length())
			assert.Equal(t, tt.want.popAll, popAll)
		})
	}
}

func TestTxGroup_prune(t *testing.T) {
	type args struct {
		atomic bool
		l2Txs  []common.PoolL2Tx
	}
	type want struct {
		l2Txs        []common.PoolL2Tx
		discardedTxs []common.PoolL2Tx
		feeAverage   *big.Float
		pruned       bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "filter one tx",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 123}, {AbsoluteFee: 400}, {AbsoluteFee: 10}},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{{AbsoluteFee: 300}},
				discardedTxs: []common.PoolL2Tx{
					{AbsoluteFee: 123, Info: ErrMaxL2TxSlot},
					{AbsoluteFee: 400, Info: ErrMaxL2TxSlot},
					{AbsoluteFee: 10, Info: ErrMaxL2TxSlot},
				},
				feeAverage: big.NewFloat(300),
				pruned:     true,
			},
		}, {
			name: "filter two txs",
			args: args{
				atomic: false,
				l2Txs:  []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}, {AbsoluteFee: 35}, {AbsoluteFee: 10}},
			},
			want: want{
				l2Txs:        []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}},
				discardedTxs: []common.PoolL2Tx{{AbsoluteFee: 35, Info: ErrMaxL2TxSlot}, {AbsoluteFee: 10, Info: ErrMaxL2TxSlot}},
				feeAverage:   big.NewFloat(350),
				pruned:       true,
			},
		}, {
			name: "filter three txs",
			args: args{
				atomic: false,
				l2Txs: []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}, {AbsoluteFee: 400},
					{AbsoluteFee: 10}},
			},
			want: want{
				l2Txs:        []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}, {AbsoluteFee: 400}},
				discardedTxs: []common.PoolL2Tx{{AbsoluteFee: 10, Info: ErrMaxL2TxSlot}},
				feeAverage:   big.NewFloat(366.6666667),
				pruned:       true,
			},
		}, {
			name: "filter all txs",
			args: args{
				atomic: false,
				l2Txs: []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}, {AbsoluteFee: 400},
					{AbsoluteFee: 10000}},
			},
			want: want{
				l2Txs: []common.PoolL2Tx{{AbsoluteFee: 300}, {AbsoluteFee: 400}, {AbsoluteFee: 400},
					{AbsoluteFee: 10000}},
				feeAverage: big.NewFloat(2775),
				pruned:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txGroup := &TxGroup{atomic: tt.args.atomic, l2Txs: tt.args.l2Txs}
			txGroup.calcFeeAverage()
			pruned := txGroup.prune()
			assert.Equal(t, tt.want.pruned, pruned)
			assert.Equal(t, tt.want.feeAverage.String(), txGroup.feeAverage.String())
			assert.EqualValues(t, tt.want.l2Txs, txGroup.l2Txs)
			assert.EqualValues(t, tt.want.discardedTxs, txGroup.discardedTxs)
		})
	}
}

func Test_checkBalanceAndNonce(t *testing.T) {
	tests := []struct {
		name string
		tx   common.PoolL2Tx
		err  error
	}{
		{
			name: "invalid sender",
			tx:   common.PoolL2Tx{Amount: big.NewInt(0), FromIdx: 44},
			err:  fmt.Errorf(ErrSenderNotFound),
		}, {
			name: "invalid nonce",
			tx:   common.PoolL2Tx{Amount: big.NewInt(0), FromIdx: 349, Nonce: 22},
			err:  fmt.Errorf(ErrInvalidNonce),
		}, {
			name: "insufficient funds",
			tx:   common.PoolL2Tx{Amount: big.NewInt(30000000000000000), FromIdx: 349, Nonce: 1},
			err:  fmt.Errorf(ErrInsufficientFunds),
		}, {
			name: "valid tx",
			tx:   common.PoolL2Tx{Amount: big.NewInt(100), FromIdx: 349, Nonce: 1},
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkBalanceAndNonce(tt.tx, createStateDbMock())
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
		})
	}
}

func Test_hashAcc(t *testing.T) {
	type args struct {
		addr    ethCommon.Address
		bjj     babyjub.PublicKeyComp
		tokenID common.TokenID
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test valid only eth address",
			args: args{
				addr: _ethAddr1,
			},
			want: "00000000e1b71c09ade004a866574a724425a29b34a8ac37d93b51edebd4d98198306b68",
		}, {
			name: "test valid only bjj address",
			args: args{
				bjj: _bjj1,
			},
			want: "00000000b68e55293f81ab5ddd13dcf0121d342d94b06cd14dd3982ac9626a83958f88f9",
		}, {
			name: "test valid only token id",
			args: args{
				tokenID: 60,
			},
			want: "0000003c01c9454a1cc17966f7a68a050696e442661cd956a383c7346536121c12e78d73",
		}, {
			name: "test valid eth address and token id",
			args: args{
				addr:    _ethAddr1,
				tokenID: 60,
			},
			want: "0000003ce1b71c09ade004a866574a724425a29b34a8ac37d93b51edebd4d98198306b68",
		}, {
			name: "test valid eth and bjj addresses",
			args: args{
				addr: _ethAddr1,
				bjj:  _bjj1,
			},
			want: "00000000fff9185e80f8d292ff12553919569751c5dc15ddb5fffae4f0c51191a7f1a96e",
		}, {
			name: "test valid only bjj address and token id",
			args: args{
				bjj:     _bjj1,
				tokenID: 60,
			},
			want: "0000003cb68e55293f81ab5ddd13dcf0121d342d94b06cd14dd3982ac9626a83958f88f9",
		}, {
			name: "test valid",
			args: args{
				addr:    _ethAddr1,
				bjj:     _bjj1,
				tokenID: 60,
			},
			want: "0000003cfff9185e80f8d292ff12553919569751c5dc15ddb5fffae4f0c51191a7f1a96e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashAccount(tt.args.addr, tt.args.bjj, tt.args.tokenID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_validateTransferToBJJ(t *testing.T) {
	tests := []struct {
		name string
		tx   common.PoolL2Tx
		err  error
	}{
		{
			name: "error invalid destination idx",
			tx:   common.PoolL2Tx{ToIdx: 10},
			err:  fmt.Errorf(ErrInvalidToIdx),
		}, {
			name: "error invalid bjj destination address",
			tx:   common.PoolL2Tx{ToIdx: 0},
			err:  fmt.Errorf(ErrInvalidToBjjAddr),
		}, {
			name: "error invalid eth destination address",
			tx: common.PoolL2Tx{
				ToIdx:     0,
				ToEthAddr: _ethAddr1,
				ToBJJ:     _bjj1,
			},
			err: fmt.Errorf(ErrInvalidToFAddr),
		}, {
			name: "valid bjj address",
			tx: common.PoolL2Tx{
				ToIdx:     0,
				ToEthAddr: common.FFAddr,
				ToBJJ:     _bjj1,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransferToBJJ(tt.tx, createStateDbMock())
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
		})
	}
}

func Test_validateTransferToEthAddr(t *testing.T) {
	tests := []struct {
		name string
		tx   common.PoolL2Tx
		err  error
	}{
		{
			name: "error invalid destination idx",
			tx:   common.PoolL2Tx{ToIdx: 10},
			err:  fmt.Errorf(ErrInvalidToIdx),
		}, {
			name: "error invalid eth destination address",
			tx:   common.PoolL2Tx{ToIdx: 0},
			err:  fmt.Errorf(ErrInvalidToEthAddr),
		}, {
			name: "error invalid eth FFF... destination address",
			tx:   common.PoolL2Tx{ToIdx: 0, ToEthAddr: common.FFAddr},
			err:  fmt.Errorf(ErrInvalidToEthAddr),
		}, {
			name: "error invalid eth destination address",
			tx: common.PoolL2Tx{
				ToIdx:     0,
				ToEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
			},
			err: fmt.Errorf("GetIdxByEthAddr: Idx can not be found: ToEthAddr: " +
				"0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370, TokenID: 0"),
		}, {
			name: "error no auth for eth destination address",
			tx: common.PoolL2Tx{
				ToIdx:     0,
				ToEthAddr: ethCommon.HexToAddress("0xA3C88ac39A76789437AED31B9608da72e1bbfBF9"),
			},
			err: fmt.Errorf("GetIdxByEthAddr: Idx can not be found: ToEthAddr: 0xA3C88ac39A76789437AED31B9608da72e1bbfBF9, TokenID: 0"),
		}, {
			name: "valid eth transaction",
			tx: common.PoolL2Tx{
				ToIdx:     0,
				ToEthAddr: _ethAddr2,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransferToEthAddr(tt.tx, createL2DBMock(), createStateDbMock())
			if tt.err != nil {
				assert.Equal(t, tt.err, tracerr.Unwrap(err))
				return
			}
			require.NoError(t, err, err)
		})
	}
}
