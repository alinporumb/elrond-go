package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

type multipleShardsCoordinatorMock struct {
	ComputeIdCalled func(address state.AddressContainer) uint32
	noShards        uint32
	CurrentShard    uint32
}

// NewMultiShardsCoordinatorMock -
func NewMultiShardsCoordinatorMock(nrShard uint32) *multipleShardsCoordinatorMock {
	return &multipleShardsCoordinatorMock{noShards: nrShard}
}

// NumberOfShards -
func (scm *multipleShardsCoordinatorMock) NumberOfShards() uint32 {
	return scm.noShards
}

// ComputeId -
func (scm *multipleShardsCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	if scm.ComputeIdCalled == nil {
		return scm.SelfId()
	}
	return scm.ComputeIdCalled(address)
}

// SelfId -
func (scm *multipleShardsCoordinatorMock) SelfId() uint32 {
	return scm.CurrentShard
}

// SetSelfId -
func (scm *multipleShardsCoordinatorMock) SetSelfId(shardId uint32) error {
	return nil
}

// SameShard -
func (scm *multipleShardsCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

// SetNoShards -
func (scm *multipleShardsCoordinatorMock) SetNoShards(noShards uint32) {
	scm.noShards = noShards
}

// CommunicationIdentifier returns the identifier between current shard ID and destination shard ID
// identifier is generated such as the first shard from identifier is always smaller than the last
func (scm *multipleShardsCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == scm.CurrentShard {
		return fmt.Sprintf("_%d", scm.CurrentShard)
	}

	if destShardID < scm.CurrentShard {
		return fmt.Sprintf("_%d_%d", destShardID, scm.CurrentShard)
	}

	return fmt.Sprintf("_%d_%d", scm.CurrentShard, destShardID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *multipleShardsCoordinatorMock) IsInterfaceNil() bool {
	if scm == nil {
		return true
	}
	return false
}
