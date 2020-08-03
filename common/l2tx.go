package common

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	Tx
	Position int // Position among all the L1Txs in that batch
}
