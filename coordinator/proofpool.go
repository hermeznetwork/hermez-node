package coordinator

import (
	"context"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/ztrue/tracerr"
)

// ServerProofInterface is the interface to a ServerProof that calculates zk proofs
type ServerProofInterface interface {
	CalculateProof(zkInputs *common.ZKInputs) error
	GetProof(ctx context.Context) (*Proof, error)
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
	log.Error("TODO")
	return tracerr.Wrap(errTODO)
}

// GetProof retreives the Proof from the ServerProof
func (p *ServerProof) GetProof(ctx context.Context) (*Proof, error) {
	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
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
func (p *ServerProofMock) GetProof(ctx context.Context) (*Proof, error) {
	// Simulate a delay
	select {
	case <-time.After(200 * time.Millisecond): //nolint:gomnd
		return &Proof{}, nil
	case <-ctx.Done():
		return nil, ErrDone
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
func (p *ServerProofPool) Get(ctx context.Context) (ServerProofInterface, error) {
	select {
	case <-ctx.Done():
		log.Info("ServerProofPool.Get done")
		return nil, ErrDone
	case serverProof := <-p.pool:
		return serverProof, nil
	}
}
