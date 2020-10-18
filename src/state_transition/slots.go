package state_transition

import (
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/prysmaticlabs/go-ssz"
)

func (st *StateTransition) ProcessSlots(state *core.State, slot uint64) error {
	for state.CurrentSlot < slot {
		if err := processSlot(state); err != nil {
			return err
		}

		// Process epoch on the first slot of the next epoch
		if canProcessEpoch(state) {
			if err := processEpoch(state); err != nil {
				return err
			}
		}

		state.CurrentSlot ++
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
	// state root
	stateRoot, err := ssz.HashTreeRoot(state)
	if err != nil {
		return err
	}
	state.XStateRoots = append(state.XStateRoots, &core.SlotAndBytes{
		Slot:                 state.CurrentSlot,
		Bytes:                stateRoot[:],// TODO - SLOTS_PER_HISTORICAL_ROOT
	})

	// update latest header
	state.LatestBlockHeader.StateRoot = stateRoot[:]

	// add block root
	root, err := ssz.HashTreeRoot(state.LatestBlockHeader)
	if err != nil {
		return err
	}
	state.XBlockRoots = append(state.XBlockRoots, &core.SlotAndBytes{
		Slot:                state.CurrentSlot,
		Bytes:               root[:], // TODO - SLOTS_PER_HISTORICAL_ROOT
	})

	return nil
}

func canProcessEpoch(state *core.State) bool {
	return (state.CurrentSlot + 1) % params.ChainConfig.SlotsInEpoch == 0
}