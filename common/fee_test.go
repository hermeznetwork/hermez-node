package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcFeeAmount(t *testing.T) {
	v := big.NewInt(1000)
	feeAmount := CalcFeeAmount(v, FeeSelector(225)) // 1000%
	assert.Equal(t, "10000", feeAmount.String())

	feeAmount = CalcFeeAmount(v, FeeSelector(224)) // 100%
	assert.Equal(t, "1000", feeAmount.String())

	feeAmount = CalcFeeAmount(v, FeeSelector(200)) // 10%
	assert.Equal(t, "100", feeAmount.String())

	feeAmount = CalcFeeAmount(v, FeeSelector(193)) // 5.11%
	assert.Equal(t, "51", feeAmount.String())

	feeAmount = CalcFeeAmount(v, FeeSelector(176)) // 1%
	assert.Equal(t, "10", feeAmount.String())

	feeAmount = CalcFeeAmount(v, FeeSelector(152)) // 0.1%
	assert.Equal(t, "1", feeAmount.String())
}
