package coordinator

import (
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func TestBatchQueue(t *testing.T) {
	bq := BatchQueue{}

	bq.Push(&BatchInfo{
		batchNum: 0,
	})
	bq.Push(&BatchInfo{
		batchNum: 2,
	})
	bq.Push(&BatchInfo{
		batchNum: 1,
	})

	assert.Equal(t, common.BatchNum(0), bq.Pop().batchNum)
	assert.Equal(t, common.BatchNum(2), bq.Pop().batchNum)
	assert.Equal(t, common.BatchNum(1), bq.Pop().batchNum)
	assert.Nil(t, bq.Pop())
}
