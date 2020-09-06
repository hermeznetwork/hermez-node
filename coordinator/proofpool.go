package coordinator

import "github.com/hermeznetwork/hermez-node/common"

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
func (p *ServerProofPool) GetNextAvailable() (*ServerProofInfo, error) {
	return nil, nil
}
