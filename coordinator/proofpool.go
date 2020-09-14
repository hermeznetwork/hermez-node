package coordinator

import (
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

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

func NewServerProof(URL string) *ServerProof {
	return &ServerProof{URL: URL}
}

// CalculateProof sends the *common.ZKInputs to the ServerProof to compute the
// Proof
func (p *ServerProof) CalculateProof(zkInputs *common.ZKInputs) error {
	return nil
}

// GetProof retreives the Proof from the ServerProof
func (p *ServerProof) GetProof(stopCh chan bool) (*Proof, error) {
	return nil, nil
}

// ServerProofPool contains the multiple ServerProof
type ServerProofPool struct {
	pool chan ServerProofInterface
}

func NewServerProofPool(maxServerProofs int) *ServerProofPool {
	return &ServerProofPool{
		pool: make(chan ServerProofInterface, maxServerProofs),
	}
}

func (p *ServerProofPool) Add(serverProof ServerProofInterface) {
	p.pool <- serverProof
}

// Get returns the available ServerProof
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
