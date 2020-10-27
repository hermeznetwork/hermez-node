package api

import (
	"fmt"
	"testing"

	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

type testCoordinatorsResponse struct {
	Coordinators []historydb.CoordinatorAPI `json:"coordinators"`
	Pagination   *db.Pagination             `json:"pagination"`
}

func (t *testCoordinatorsResponse) GetPagination() *db.Pagination {
	if t.Coordinators[0].ItemID < t.Coordinators[len(t.Coordinators)-1].ItemID {
		t.Pagination.FirstReturnedItem = t.Coordinators[0].ItemID
		t.Pagination.LastReturnedItem = t.Coordinators[len(t.Coordinators)-1].ItemID
	} else {
		t.Pagination.LastReturnedItem = t.Coordinators[0].ItemID
		t.Pagination.FirstReturnedItem = t.Coordinators[len(t.Coordinators)-1].ItemID
	}
	return t.Pagination
}

func (t *testCoordinatorsResponse) Len() int { return len(t.Coordinators) }

func TestGetCoordinators(t *testing.T) {
	endpoint := apiURL + "coordinators"
	fetchedCoordinators := []historydb.CoordinatorAPI{}

	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testCoordinatorsResponse).Coordinators); i++ {
			tmp, err := copystructure.Copy(intr.(*testCoordinatorsResponse).Coordinators[i])
			if err != nil {
				panic(err)
			}
			fetchedCoordinators = append(fetchedCoordinators, tmp.(historydb.CoordinatorAPI))
		}
	}

	limit := 5

	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &testCoordinatorsResponse{}, appendIter)
	assert.NoError(t, err)
	for i := 0; i < len(fetchedCoordinators); i++ {
		assert.Equal(t, tc.coordinators[i].ItemID, fetchedCoordinators[i].ItemID)
		assert.Equal(t, tc.coordinators[i].Bidder, fetchedCoordinators[i].Bidder)
		assert.Equal(t, tc.coordinators[i].Forger, fetchedCoordinators[i].Forger)
		assert.Equal(t, tc.coordinators[i].EthBlockNum, fetchedCoordinators[i].EthBlockNum)
		assert.Equal(t, tc.coordinators[i].URL, fetchedCoordinators[i].URL)
	}

	// Reverse Order
	reversedCoordinators := []historydb.CoordinatorAPI{}
	appendIter = func(intr interface{}) {
		for i := 0; i < len(intr.(*testCoordinatorsResponse).Coordinators); i++ {
			tmp, err := copystructure.Copy(intr.(*testCoordinatorsResponse).Coordinators[i])
			if err != nil {
				panic(err)
			}
			reversedCoordinators = append(reversedCoordinators, tmp.(historydb.CoordinatorAPI))
		}
	}
	err = doGoodReqPaginated(path, historydb.OrderDesc, &testCoordinatorsResponse{}, appendIter)
	assert.NoError(t, err)
	for i := 0; i < len(fetchedCoordinators); i++ {
		assert.Equal(t, reversedCoordinators[i].ItemID, fetchedCoordinators[len(fetchedCoordinators)-1-i].ItemID)
		assert.Equal(t, reversedCoordinators[i].Bidder, fetchedCoordinators[len(fetchedCoordinators)-1-i].Bidder)
		assert.Equal(t, reversedCoordinators[i].Forger, fetchedCoordinators[len(fetchedCoordinators)-1-i].Forger)
		assert.Equal(t, reversedCoordinators[i].EthBlockNum, fetchedCoordinators[len(fetchedCoordinators)-1-i].EthBlockNum)
		assert.Equal(t, reversedCoordinators[i].URL, fetchedCoordinators[len(fetchedCoordinators)-1-i].URL)
	}

	// Test GetCoordinator
	path = fmt.Sprintf("%s/%s", endpoint, fetchedCoordinators[2].Forger.String())
	coordinator := historydb.CoordinatorAPI{}
	assert.NoError(t, doGoodReq("GET", path, nil, &coordinator))
	assert.Equal(t, fetchedCoordinators[2], coordinator)

	// 400
	path = fmt.Sprintf("%s/0x001", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// 404
	path = fmt.Sprintf("%s/0xaa942cfcd25ad4d90a62358b0dd84f33b398262a", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}
