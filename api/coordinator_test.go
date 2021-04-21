package api

import (
	"fmt"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

type testCoordinatorsResponse struct {
	Coordinators []historydb.CoordinatorAPI `json:"coordinators"`
	PendingItems uint64                     `json:"pendingItems"`
}

func (t testCoordinatorsResponse) GetPending() (pendingItems, lastItemID uint64) {
	if len(t.Coordinators) == 0 {
		return 0, 0
	}
	pendingItems = t.PendingItems
	lastItemID = t.Coordinators[len(t.Coordinators)-1].ItemID
	return pendingItems, lastItemID
}

func (t *testCoordinatorsResponse) Len() int { return len(t.Coordinators) }

func (t testCoordinatorsResponse) New() Pendinger { return &testCoordinatorsResponse{} }

func genTestCoordinators(coordinators []common.Coordinator) []historydb.CoordinatorAPI {
	testCoords := []historydb.CoordinatorAPI{}
	for i := 0; i < len(coordinators); i++ {
		alreadyRegistered := false
		for j := 0; j < len(testCoords); j++ {
			// coordinator already registered
			if coordinators[i].Bidder == testCoords[j].Bidder {
				alreadyRegistered = true
				if coordinators[i].EthBlockNum > testCoords[j].EthBlockNum {
					// This occurrence is more updated, delete current and append new
					testCoords = append(testCoords[:j], testCoords[j+1:]...)
					alreadyRegistered = false
				}
				break
			}
		}
		if !alreadyRegistered {
			testCoords = append(testCoords, historydb.CoordinatorAPI{
				Bidder:      coordinators[i].Bidder,
				Forger:      coordinators[i].Forger,
				EthBlockNum: coordinators[i].EthBlockNum,
				URL:         coordinators[i].URL,
			})
		}
	}
	return testCoords
}

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

	// All
	limit := 5
	path := fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err := doGoodReqPaginated(path, db.OrderAsc, &testCoordinatorsResponse{}, appendIter)
	assert.NoError(t, err)
	assertCoordinators(t, tc.coordinators, fetchedCoordinators)

	// All in reverse order
	fetchedCoordinators = []historydb.CoordinatorAPI{}
	err = doGoodReqPaginated(path, db.OrderDesc, &testCoordinatorsResponse{}, appendIter)
	assert.NoError(t, err)
	reversedCoordinators := []historydb.CoordinatorAPI{}
	for i := 0; i < len(tc.coordinators); i++ {
		reversedCoordinators = append(reversedCoordinators, tc.coordinators[len(tc.coordinators)-1-i])
	}
	assertCoordinators(t, reversedCoordinators, fetchedCoordinators)
	for _, filteredCoord := range tc.coordinators {
		// By bidder
		fetchedCoordinators = []historydb.CoordinatorAPI{}
		err = doGoodReqPaginated(
			fmt.Sprintf(path+"&bidderAddr=%s", filteredCoord.Bidder.String()),
			db.OrderAsc, &testCoordinatorsResponse{}, appendIter,
		)
		assert.NoError(t, err)
		assertCoordinators(t, []historydb.CoordinatorAPI{filteredCoord}, fetchedCoordinators)
		// By forger
		fetchedCoordinators = []historydb.CoordinatorAPI{}
		err = doGoodReqPaginated(
			fmt.Sprintf(path+"&forgerAddr=%s", filteredCoord.Forger.String()),
			db.OrderAsc, &testCoordinatorsResponse{}, appendIter,
		)
		assert.NoError(t, err)
		assertCoordinators(t, []historydb.CoordinatorAPI{filteredCoord}, fetchedCoordinators)
	}

	// Empty array
	fetchedCoordinators = []historydb.CoordinatorAPI{}
	path = fmt.Sprintf("%s?bidderAddr=0xaa942cfcd25ad4d90a62358b0dd84f33b398262a", endpoint)
	err = doGoodReqPaginated(path, db.OrderDesc, &testCoordinatorsResponse{}, appendIter)
	assert.NoError(t, err)
	assertCoordinators(t, []historydb.CoordinatorAPI{}, fetchedCoordinators)

	// 400
	path = fmt.Sprintf("%s?bidderAddr=0x001", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
}

func assertCoordinator(t *testing.T, expected, actual historydb.CoordinatorAPI) {
	actual.ItemID = 0
	assert.Equal(t, expected, actual)
}

func assertCoordinators(t *testing.T, expected, actual []historydb.CoordinatorAPI) {
	assert.Equal(t, len(expected), len(actual))
	for i := 0; i < len(expected); i++ {
		assertCoordinator(t, expected[i], actual[i])
	}
}
