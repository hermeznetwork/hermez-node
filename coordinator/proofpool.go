package coordinator

import (
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

// ServerProofInfo contains the data related to a ServerProof
type ServerProofInfo struct {
	// TODO
	Available bool
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ServerProofInfo) CalculateProof(zkInputs *common.ZKInputs) error {
	return nil
}

// GetProof retreives the Proof from the ServerProof
func (p *ServerProofInfo) GetProof() (*Proof, error) {
	return nil, nil
}

// ServerProofPool contains the multiple ServerProofInfo
type ServerProofPool struct {
	// pool []ServerProofInfo
}

// GetNextAvailable returns the available ServerProofInfo
func (p *ServerProofPool) GetNextAvailable(stopCh chan bool) (*ServerProofInfo, error) {
	select {
	case <-stopCh:
		log.Info("ServerProofPool.GetNextAvailable stopped")
		return nil, ErrStop
	default:
	}
	return nil, nil
}
