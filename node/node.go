package node

type mode string

// ModeCoordinator defines the mode of the HermezNode as Coordinator, which
// means that the node is set to forge (which also will be synchronizing with
// the L1 blockchain state)
const ModeCoordinator mode = "coordinator"

// ModeSynchronizer defines the mode of the HermezNode as Synchronizer, which
// means that the node is set to only synchronize with the L1 blockchain state
// and will not forge
const ModeSynchronizer mode = "synchronizer"
