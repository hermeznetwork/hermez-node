package coordinator

import "github.com/hermeznetwork/hermez-node/common"

type ServerProofInfo struct {
	// TODO
	Available bool
}

func (p *ServerProofInfo) CalculateProof(zkInputs *common.ZKInputs) error {
	return nil
}

func (p *ServerProofInfo) GetProof() (*Proof, error) {
	return nil, nil
}

type ServerProofPool struct {
	// pool []ServerProofInfo
}

func (p *ServerProofPool) GetNextAvailable() (*ServerProofInfo, error) {
	return nil, nil
}
