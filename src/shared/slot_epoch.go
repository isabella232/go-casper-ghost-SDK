package shared

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

/**
def compute_epoch_at_slot(slot: Slot) -> Epoch:
    """
    Return the epoch number at ``slot``.
    """
    return Epoch(slot // SLOTS_PER_EPOCH)v
 */
func ComputeEpochAtSlot(slot uint64) uint64 {
	return slot/ params.ChainConfig.SlotsInEpoch
}

/**
def compute_start_slot_at_epoch(epoch: Epoch) -> Slot:
    """
    Return the start slot of ``epoch``.
    """
    return Slot(epoch * SLOTS_PER_EPOCH)
 */
func ComputeStartSlotAtEpoch(epoch uint64) uint64 {
	return epoch * params.ChainConfig.SlotsInEpoch
}

/**
def get_current_epoch(state: BeaconState) -> Epoch:
    """
    Return the current epoch.
    """
    return compute_epoch_at_slot(state.slot)
 */
func GetCurrentEpoch(state *core.State) uint64 {
	return ComputeEpochAtSlot(state.CurrentSlot)
}

/**
def get_previous_epoch(state: BeaconState) -> Epoch:
    """`
    Return the previous epoch (unless the current epoch is ``GENESIS_EPOCH``).
    """
    current_epoch = get_current_epoch(state)
    return GENESIS_EPOCH if current_epoch == GENESIS_EPOCH else Epoch(current_epoch - 1)
 */
func GetPreviousEpoch(state *core.State) uint64 {
	if GetCurrentEpoch(state) == params.ChainConfig.GenesisEpoch {
		return params.ChainConfig.GenesisEpoch
	}
	return GetCurrentEpoch(state) - 1
}

/**
def get_block_root(state: BeaconState, epoch: Epoch) -> Root:
    """
    Return the block root at the start of a recent ``epoch``.
    """
    return get_block_root_at_slot(state, compute_start_slot_at_epoch(epoch))
 */
func GetBlockRoot(state *core.State, epoch uint64) (*core.SlotAndBytes, error) {
	return GetBlockRootAtSlot(state, ComputeStartSlotAtEpoch(epoch))
}

/**
def get_block_root_at_slot(state: BeaconState, slot: Slot) -> Root:
    """
    Return the block root at a recent ``slot``.
    """
    assert slot < state.slot <= slot + SLOTS_PER_HISTORICAL_ROOT
    return state.block_roots[slot % SLOTS_PER_HISTORICAL_ROOT]
 */
func GetBlockRootAtSlot(state *core.State, slot uint64) (*core.SlotAndBytes, error) {
	if slot >= state.CurrentSlot || state.CurrentSlot > slot + params.ChainConfig.SlotsPerHistoricalRoot {
		return nil, fmt.Errorf("block root at slot not found")
	}
	for _, blk := range state.XBlockRoots {
		if blk.Slot == slot {
			return blk, nil
		}
	}
	return nil, fmt.Errorf("block root at slot not found")
}