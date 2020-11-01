package state_transition

import (
	"bytes"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
)

func (st *StateTransition) ProcessSlots(state *core.State, slot uint64) error {
	for state.Slot < slot {
		if err := processSlot(state); err != nil {
			return err
		}
		// Process epoch on the first slot of the next epoch
		if canProcessEpoch(state) {
			if err := processEpoch(state); err != nil {
				return err
			}
		}
		state.Slot++
	}

	return nil
}

// ProcessSlot happens every slot and focuses on the slot counter and block roots record updates.
// It happens regardless if there's an incoming block or not.
// Spec pseudocode definition:
//
//  def process_slot(state: BeaconState) -> None:
//    # Cache state root
//    previous_state_root = hash_tree_root(state)
//    state.state_roots[state.slot % SLOTS_PER_HISTORICAL_ROOT] = previous_state_root
//
//    # Cache latest block header state root
//    if state.latest_block_header.state_root == Bytes32():
//        state.latest_block_header.state_root = previous_state_root
//
//    # Cache block root
//    previous_block_root = hash_tree_root(state.latest_block_header)
//    state.block_roots[state.slot % SLOTS_PER_HISTORICAL_ROOT] = previous_block_root
func processSlot(state *core.State) error {
	// state prevBlockRoot
	prevStateRoot, err := state.HashTreeRoot()
	if err != nil {
		return err
	}
	state.StateRoots[state.Slot % params.ChainConfig.SlotsPerHistoricalRoot] = prevStateRoot[:]

	// update latest header
	if state.LatestBlockHeader.StateRoot == nil || bytes.Equal(state.LatestBlockHeader.StateRoot, params.ChainConfig.ZeroHash) {
		state.LatestBlockHeader.StateRoot = prevStateRoot[:]
	}

	// add block prevBlockRoot
	prevBlockRoot, err := state.LatestBlockHeader.HashTreeRoot()
	if err != nil {
		return err
	}
	state.BlockRoots[state.Slot % params.ChainConfig.SlotsPerHistoricalRoot] = prevBlockRoot[:]
	return nil
}

func canProcessEpoch(state *core.State) bool {
	return (state.Slot+ 1) % params.ChainConfig.SlotsInEpoch == 0
}