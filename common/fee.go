package common

import "math"

// Fee is a type that represents the percentage of tokens that will be paid in a transaction
// to incentivaise the materialization of it
type Fee float64

// RecommendedFee is the recommended fee to pay in USD per transaction set by
// the coordinator according to the tx type (if the tx requires to create an
// account and register, only register or he account already esists)
type RecommendedFee struct {
	ExistingAccount           float64
	CreatesAccount            float64
	CreatesAccountAndRegister float64
}

// FeeSelector is used to select a percentage from the FeePlan.
type FeeSelector uint8

// Percentage returns the associated percentage of the FeeSelector
func (f FeeSelector) Percentage() float64 {
	if f == 0 {
		return 0
		//nolint:gomnd
	} else if f <= 32 { //nolint:gomnd
		return math.Pow(10, -24+(float64(f)/2)) //nolint:gomnd
	} else if f <= 223 { //nolint:gomnd
		return math.Pow(10, -8+(0.041666666666667*(float64(f)-32))) //nolint:gomnd
	} else {
		return math.Pow(10, float64(f)-224) //nolint:gomnd
	}
}

// MaxFeePlan is the maximum value of the FeePlan
const MaxFeePlan = 256

// FeePlan represents the fee model, a position in the array indicates the
// percentage of tokens paid in concept of fee for a transaction
var FeePlan = [MaxFeePlan]float64{}
