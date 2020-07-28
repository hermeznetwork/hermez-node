package common

// Fee is a type that represents the percentage of tokens that will be payed in a transaction
// to incentivaise the materialization of it
type Fee float64

// RecommendedFee is the recommended fee to pay in USD per transaction set by the coordinator
// according to the tx type (if the tx requires to create an account and register, only register or he account already esists)
type RecommendedFee struct {
	ExistingAccount           float64
	CreatesAccount            float64
	CreatesAccountAndRegister float64
}

// FeeSelector is used to select a percentage from the FeePlan.
type FeeSelector uint8

// FeePlan represents the fee model, a position in the array indicates the percentage of tokens paid in concept of fee for a transaction
var FeePlan = [256]float64{}
