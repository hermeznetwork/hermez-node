package common

import (
	"time"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	Tx
	Forged   time.Time // time when received by the tx pool
	BatchNum BatchNum  // Batch in which the tx was forged, 0 means not forged
}
