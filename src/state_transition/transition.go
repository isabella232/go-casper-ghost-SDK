package state_transition

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/prysmaticlabs/go-ssz"
	"log"
)

type IStateTransition interface {
	// ExecuteStateTransition runs ExecuteNoVerify and verifies post transition state root
	//
	// Spec pseudocode definition:
	//  def state_transition(state: BeaconState, signed_block: SignedBeaconBlock, validate_result: bool=True) -> BeaconState:
	//    block = signed_block.message
	//    # Process slots (including those with no blocks) since block
	//    process_slots(state, block.slot)
	//    # Verify signature
	//    if validate_result:
	//        assert verify_block_signature(state, signed_block)
	//    # Process block
	//    process_block(state, block)
	//    if validate_result:
	//        assert block.state_root == hash_tree_root(state)
	//    # Return post-state
	//    return state
	ExecuteStateTransition(state *core.State, signedBlock *core.SignedPoolBlock) (newState *core.State, err error)

	// ComputeStateRoot defines the procedure for a state transition function.
	// This does not validate any BLS signatures in a block, it is used for calculating the
	// state root of the state for the block proposer to use.
	// This does not modify state.
	//
	// WARNING: This method does not validate any BLS signatures. This is used for proposer to compute
	// state root before proposing a new block, and this does not modify state.
	//
	// Spec pseudocode definition:
	//  def state_transition(state: BeaconState, block: BeaconBlock, validate_state_root: bool=False) -> BeaconState:
	//    # Process slots (including those with no blocks) since block
	//    process_slots(state, block.slot)
	//    # Process block
	//    process_block(state, block)
	//    # Return post-state
	//    return state
	ComputeStateRoot(state *core.State, signedBlock *core.SignedPoolBlock) ([32]byte, error)

	// ProcessBlock creates a new, modified beacon state by applying block operation
	// transformations as defined in the Ethereum Serenity specification, including processing proposer slashings,
	// processing block attestations, and more.
	//
	// Spec pseudocode definition:
	//
	//  def process_block(state: BeaconState, block: BeaconBlock) -> None:
	//    process_block_header(state, block)
	//    process_randao(state, block.block)
	//    process_eth1_data(state, block.block)
	//    process_operations(state, block.block)
	ProcessBlock(state *core.State, newBlockBody *core.SignedPoolBlock) error
	// ProcessSlots process through skip slots and apply epoch transition when it's needed
	//
	// Spec pseudocode definition:
	//  def process_slots(state: BeaconState, slot: Slot) -> None:
	//    assert state.slot <= slot
	//    while state.slot < slot:
	//        process_slot(state)
	//        # Process epoch on the first slot of the next epoch
	//        if (state.slot + 1) % SLOTS_PER_EPOCH == 0:
	//            process_epoch(state)
	//        state.slot += 1
	//    ]
	ProcessSlots(state *core.State, slot uint64) error
}

type StateTransition struct {}
func NewStateTransition() *StateTransition { return &StateTransition{} }

func (st *StateTransition)ExecuteStateTransition(state *core.State, signedBlock *core.SignedPoolBlock) (newState *core.State, err error) {
	log.Printf("processing block at slot %d\n", signedBlock.Block.Slot)
	newState = shared.CopyState(state)

	if err := st.ProcessSlots(newState, signedBlock.Block.Slot); err != nil {
		return nil, err
	}

	if err := st.ProcessBlock(newState, signedBlock); err != nil {
		return nil, err
	}

	postStateRoot, err := ssz.HashTreeRoot(newState)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(signedBlock.Block.StateRoot, postStateRoot[:]) {
		return nil, fmt.Errorf("new block state root is wrong, expected %s", hex.EncodeToString(postStateRoot[:]))
	}

	return newState, nil
}


func (st *StateTransition) ComputeStateRoot(state *core.State, signedBlock *core.SignedPoolBlock) ([32]byte, error) {
	stateCopy := shared.CopyState(state)

	if err := st.ProcessSlots(stateCopy, signedBlock.Block.Slot); err != nil {
		return [32]byte{}, err
	}

	if err := st.processBlockForStateRoot(stateCopy, signedBlock); err != nil {
		return [32]byte{}, err
	}

	return ssz.HashTreeRoot(stateCopy)
}


//// A helper function to insert the post block state root to the block block
//// TODO - move from here
//func CalculateAndInsertStateRootToBlock(state *core.State, signedBlock *core.SignedBlock) ([]byte, error) {
//	st := NewStateTransition()
//
//	newState := shared.CopyState(state)
//	err := st.ProcessBlock(newState, signedBlock)
//	if err != nil {
//		return []byte{}, err
//	}
//
//	root := shared.GetStateRoot(newState, newState.CurrentSlot)
//	if len(root) == 0 {
//		return []byte{}, fmt.Errorf("could not find statet root for epoch %d", newState.CurrentSlot)
//	}
//
//	return root[:], nil
//}