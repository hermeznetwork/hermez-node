package common

// SCVariables joins all the smart contract variables in a single struct
type SCVariables struct {
	Rollup   RollupVariables   `validate:"required"`
	Auction  AuctionVariables  `validate:"required"`
	WDelayer WDelayerVariables `validate:"required"`
}

// AsPtr returns the SCVariables as a SCVariablesPtr using pointers to the
// original SCVariables
func (v *SCVariables) AsPtr() *SCVariablesPtr {
	return &SCVariablesPtr{
		Rollup:   &v.Rollup,
		Auction:  &v.Auction,
		WDelayer: &v.WDelayer,
	}
}

// SCVariablesPtr joins all the smart contract variables as pointers in a single
// struct
type SCVariablesPtr struct {
	Rollup   *RollupVariables   `validate:"required"`
	Auction  *AuctionVariables  `validate:"required"`
	WDelayer *WDelayerVariables `validate:"required"`
}

// SCConsts joins all the smart contract constants in a single struct
type SCConsts struct {
	Rollup   RollupConstants
	Auction  AuctionConstants
	WDelayer WDelayerConstants
}
