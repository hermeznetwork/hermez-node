package coordinator

import (
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

// ServerProofInterface is the interface to a ServerProof that calculates zk proofs
type ServerProofInterface interface {
	CalculateProof(zkInputs *common.ZKInputs) error
	GetProof(stopCh chan bool) (*Proof, error)
}

// ServerProof contains the data related to a ServerProof
type ServerProof struct {
	// TODO
	URL       string
	Available bool
}

// NewServerProof creates a new ServerProof
func NewServerProof(URL string) *ServerProof {
	return &ServerProof{URL: URL}
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ServerProof) CalculateProof(zkInputs *common.ZKInputs) error {
	return errTODO
}

// GetProof retreives the Proof from the ServerProof
func (p *ServerProof) GetProof(stopCh chan bool) (*Proof, error) {
	return nil, errTODO
}

// ServerProofMock is a mock ServerProof to be used in tests.  It doesn't calculate anything
type ServerProofMock struct {
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ServerProofMock) CalculateProof(zkInputs *common.ZKInputs) error {
	return nil
}

// GetProof retreives the Proof from the ServerProof
func (p *ServerProofMock) GetProof(stopCh chan bool) (*Proof, error) {
	// Simulate a delay
	select {
	case <-time.After(200 * time.Millisecond): //nolint:gomnd
		return &Proof{}, nil
	case <-stopCh:
		return nil, ErrStop
	}
}

// ServerProofPool contains the multiple ServerProof
type ServerProofPool struct {
	pool chan ServerProofInterface
}

// NewServerProofPool creates a new pool of ServerProofs.
func NewServerProofPool(maxServerProofs int) *ServerProofPool {
	return &ServerProofPool{
		pool: make(chan ServerProofInterface, maxServerProofs),
	}
}

// Add a ServerProof to the pool
func (p *ServerProofPool) Add(serverProof ServerProofInterface) {
	p.pool <- serverProof
}

// Get returns the next available ServerProof
func (p *ServerProofPool) Get(stopCh chan bool) (ServerProofInterface, error) {
	select {
	case <-stopCh:
		log.Info("ServerProofPool.Get stopped")
		return nil, ErrStop
	default:
		select {
		case <-stopCh:
			log.Info("ServerProofPool.Get stopped")
			return nil, ErrStop
		case serverProof := <-p.pool:
			return serverProof, nil
		}
	}
}
