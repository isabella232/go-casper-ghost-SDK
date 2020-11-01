package state_transition

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/go-ssz"
)

func (st *StateTransition) ProcessBlock(state *core.State, block *core.Block) error {
	if err := ProcessBlockHeader(state, block); err != nil {
		return fmt.Errorf("ProcessBlock: %s", err.Error())
	}
	if err := processRANDAO(state, block); err != nil {
		return fmt.Errorf("ProcessBlock: %s", err.Error())
	}
	if err := processEth1Data(state, block.Body); err != nil {
		return fmt.Errorf("ProcessBlock: %s", err.Error())
	}
	if err := processOperations(state, block.Body); err != nil {
		return fmt.Errorf("ProcessBlock: %s", err.Error())
	}
	return nil
}

func (st *StateTransition) processBlockForStateRoot(state *core.State, signedBlock *core.SignedBlock) error {
	if err := ProcessBlockHeader(state, signedBlock.Block); err != nil {
		return fmt.Errorf("processBlockForStateRoot: %s", err.Error())
	}
	if err := processRANDAONoVerify(state, signedBlock.Block); err != nil {
		return fmt.Errorf("processBlockForStateRoot: %s", err.Error())
	}
	if err := processEth1Data(state, signedBlock.Block.Body); err != nil {
		return fmt.Errorf("processBlockForStateRoot: %s", err.Error())
	}
	if err := processOperationsNoVerify(state, signedBlock.Block.Body); err != nil {
		return fmt.Errorf("processBlockForStateRoot: %s", err.Error())
	}
	return nil
}

/**
def process_block_header(state: BeaconState, block: BeaconBlock) -> None:
    # Verify that the slots match
    assert block.slot == state.slot
    # Verify that the block is newer than latest block header
    assert block.slot > state.latest_block_header.slot
    # Verify that proposer index is the correct index
    assert block.proposer_index == get_beacon_proposer_index(state)
    # Verify that the parent matches
    assert block.parent_root == hash_tree_root(state.latest_block_header)
    # Cache current block as the new latest block
    state.latest_block_header = BeaconBlockHeader(
        slot=block.slot,
        proposer_index=block.proposer_index,
        parent_root=block.parent_root,
        state_root=Bytes32(),  # Overwritten in the next process_slot call
        body_root=hash_tree_root(block.body),
    )

    # Verify proposer is not slashed
    proposer = state.validators[block.proposer_index]
    assert not proposer.slashed
 */
func ProcessBlockHeader(state *core.State, block *core.Block) error {
	// slot
	if state.Slot != block.Slot {
		return fmt.Errorf("block slot doesn't match state slot")
	}

	// Verify that the block is newer than latest block header
	if block.Slot <= state.LatestBlockHeader.Slot {
		return fmt.Errorf("bad block header")
	}

	// proposer
	expectedProposer, err := shared.GetBlockProposerIndex(state)
	if err != nil {
		return err
	}
	proposerId :=  block.GetProposer()
	if expectedProposer != proposerId {
		return fmt.Errorf("block expectedProposer is worng, expected %d but received %d", expectedProposer, proposerId)
	}

	// parent
	root,err := state.LatestBlockHeader.HashTreeRoot()
	if err != nil {
		return err
	}
	if !bytes.Equal(block.ParentRoot, root[:]) {
		return fmt.Errorf("parent block root doesn't match, expected %s", hex.EncodeToString(root[:]))
	}

	// save
	root,err = ssz.HashTreeRoot(block.Body)
	if err != nil {
		return err
	}
	state.LatestBlockHeader = &core.BlockHeader{
		Slot:                 block.Slot,
		ProposerIndex:        block.Proposer,
		ParentRoot:           block.ParentRoot,
		BodyRoot:             root[:],
		StateRoot: 			  params.ChainConfig.ZeroHash, // state_root: zeroed, overwritten in the next `process_slot` call
	}

	// verify proposer is not slashed
	val := shared.GetValidator(state, expectedProposer)
	if val == nil {
		return fmt.Errorf("could not find proposer")
	}
	if val.Slashed {
		return fmt.Errorf("block proposer is slashed")
	}

	return nil
}

/**
def process_eth1_data(state: BeaconState, body: BeaconBlockBody) -> None:
    state.eth1_data_votes.append(body.eth1_data)
    if state.eth1_data_votes.count(body.eth1_data) * 2 > EPOCHS_PER_ETH1_VOTING_PERIOD * SLOTS_PER_EPOCH:
        state.eth1_data = body.eth1_data
 */
func processEth1Data(state *core.State, body *core.BlockBody) error {
	state.Eth1DataVotes = append(state.Eth1DataVotes, body.Eth1Data)

	// count support
	voteCount := 0
	for _, vote := range state.Eth1DataVotes {
		if AreEth1DataEqual(vote, body.Eth1Data) {
			voteCount ++
		}
	}
	// If 50+% majority converged on the same eth1data, then it has enough support to update the
	// state.
	support := params.ChainConfig.EpochsPerETH1VotingPeriod * params.ChainConfig.SlotsInEpoch
	if hasSupport := uint64(voteCount) * 2 > support; hasSupport {
		state.Eth1Data = body.Eth1Data
	}
	return nil
}

/**
def process_operations(state: BeaconState, body: BeaconBlockBody) -> None:
    # Verify that outstanding deposits are processed up to the maximum number of deposits
    assert len(body.deposits) == min(MAX_DEPOSITS, state.eth1_data.deposit_count - state.eth1_deposit_index)

    def for_ops(operations: Sequence[Any], fn: Callable[[BeaconState, Any], None]) -> None:
        for operation in operations:
            fn(state, operation)

    for_ops(body.proposer_slashings, process_proposer_slashing)
    for_ops(body.attester_slashings, process_attester_slashing)
    for_ops(body.attestations, process_attestation)
    for_ops(body.deposits, process_deposit)
    for_ops(body.voluntary_exits, process_voluntary_exit)
 */
func processOperations(state *core.State, body *core.BlockBody) error {
	if err := ProcessProposerSlashings(state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := ProcessAttesterSlashings(state, body.AttesterSlashings); err != nil {
		return err
	}
	if err := ProcessBlockAttestations(state, body.Attestations); err != nil {
		return err
	}
	if err := ProcessDeposits(state, body.Deposits); err != nil {
		return err
	}
	if err := ProcessExits(state, body.VoluntaryExits); err != nil {
		return err
	}

	return nil
}

func processOperationsNoVerify(state *core.State, body *core.BlockBody) error {
	if err := ProcessProposerSlashings(state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := ProcessAttesterSlashings(state, body.AttesterSlashings); err != nil {
		return err
	}
	for _, att := range body.Attestations {
		if err := processAttestationNoSigVerify(state, att); err != nil {
			return err
		}
	}
	if err := ProcessDeposits(state, body.Deposits); err != nil {
		return err
	}
	if err := ProcessExits(state, body.VoluntaryExits); err != nil {
		return err
	}
	return nil
}

// AreEth1DataEqual checks equality between two eth1 data objects.
func AreEth1DataEqual(a, b *core.ETH1Data) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.DepositCount == b.DepositCount &&
		bytes.Equal(a.BlockHash, b.BlockHash) &&
		bytes.Equal(a.DepositRoot, b.DepositRoot)
}
