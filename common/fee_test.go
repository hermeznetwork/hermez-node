package common

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFeePercentage(t *testing.T) {
	assert.InEpsilon(t, 2.68E-18, FeeSelector(1).Percentage(), 0.002)
	assert.InEpsilon(t, 6.76E-14, FeeSelector(10).Percentage(), 0.002)
	assert.InEpsilon(t, 3.91E-03, FeeSelector(32).Percentage(), 0.002)
	assert.InEpsilon(t, 7.29E-03, FeeSelector(50).Percentage(), 0.002)
	assert.InEpsilon(t, 4.12E-02, FeeSelector(100).Percentage(), 0.002)
	assert.InEpsilon(t, 2.33E-01, FeeSelector(150).Percentage(), 0.002)
	assert.InEpsilon(t, 1.00E+00, FeeSelector(192).Percentage(), 0.002)
	assert.InEpsilon(t, 2.56E+02, FeeSelector(200).Percentage(), 0.002)
	assert.InEpsilon(t, 2.88E+17, FeeSelector(250).Percentage(), 0.002)
}

func TestCalcFeeAmount(t *testing.T) {
	v := big.NewInt(1000)
	feeAmount, err := CalcFeeAmount(v, FeeSelector(195)) // 800%
	assert.Nil(t, err)
	assert.Equal(t, "8000", feeAmount.String())

	feeAmount, err = CalcFeeAmount(v, FeeSelector(192)) // 100%
	assert.Nil(t, err)
	assert.Equal(t, "1000", feeAmount.String())

	feeAmount, err = CalcFeeAmount(v, FeeSelector(172)) // 50%
	assert.Nil(t, err)
	assert.Equal(t, "500", feeAmount.String())

	feeAmount, err = CalcFeeAmount(v, FeeSelector(126)) // 10.2%
	assert.Nil(t, err)
	assert.Equal(t, "101", feeAmount.String())

	feeAmount, err = CalcFeeAmount(v, FeeSelector(60)) // 1.03%
	assert.Nil(t, err)
	assert.Equal(t, "10", feeAmount.String())

	feeAmount, err = CalcFeeAmount(v, FeeSelector(31)) // 0.127%
	assert.Nil(t, err)
	assert.Equal(t, "1", feeAmount.String())
}

func TestFeePrintSQLSwitch(t *testing.T) {
	debug := false
	for i := 0; i < 256; i++ {
		f := FeeSelector(i).Percentage()
		if debug {
			fmt.Printf("        WHEN $1 = %03d THEN %.6e\n", i, f)
		}
	}
}
