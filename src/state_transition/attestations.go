package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-ssz"
)

func ProcessBlockAttestations(state *core.State, attestations []*core.Attestation) error {
	for _, att := range attestations {
		if err := processAttestation(state, att); err != nil {
			return err
		}
	}
	return nil
}

func processAttestation(state *core.State, attestation *core.Attestation) error {
	if err := processAttestationNoSigVerify(state, attestation); err != nil {
		return err
	}
	if err := validateAttestationSignature(state, attestation, attestation.Data.Slot); err != nil {
		return err
	}
	return nil
}

// ProcessAttestation verifies an input attestation can pass through processing using the given beacon state.
//
// https://github.com/ethereum/eth2.0-specs/blob/dev/specs/phase0/beacon-chain.md#attestations
// Spec pseudocode definition:
//  def process_attestation(state: BeaconState, attestation: Attestation) -> None:
//    data = attestation.data
//    assert data.target.epoch in (get_previous_epoch(state), get_current_epoch(state))
//    assert data.target.epoch == compute_epoch_at_slot(data.slot)
//    assert data.slot + MIN_ATTESTATION_INCLUSION_DELAY <= state.slot <= data.slot + SLOTS_PER_EPOCH
//    assert data.index < get_committee_count_per_slot(state, data.target.epoch)
//
//    committee = get_beacon_committee(state, data.slot, data.index)
//    assert len(attestation.aggregation_bits) == len(committee)
//
//    pending_attestation = PendingAttestation(
//        data=data,
//        aggregation_bits=attestation.aggregation_bits,
//        inclusion_delay=state.slot - data.slot,
//        proposer_index=get_beacon_proposer_index(state),
//    )
//
//    if data.target.epoch == get_current_epoch(state):
//        assert data.source == state.current_justified_checkpoint
//        state.current_epoch_attestations.append(pending_attestation)
//    else:
//        assert data.source == state.previous_justified_checkpoint
//        state.previous_epoch_attestations.append(pending_attestation)
//
//    # Check signature
//    assert is_valid_indexed_attestation(state, get_indexed_attestation(state, attestation))
func processAttestationNoSigVerify(state *core.State, attestation *core.Attestation) error {
	if err := validateAttestationData(state, attestation.Data); err != nil {
		return err
	}
	if err := validateAggregationBits(state, attestation); err != nil {
		return err
	}
	if err := appendPendingAttestation(state, attestation); err != nil {
		return err
	}
	if err := ProcessExecutionSummaries(state, attestation.Data.ExecutionSummaries); err != nil {
		return err
	}
	return nil
}

//    pending_attestation = PendingAttestation(
//        data=data,
//        aggregation_bits=attestation.aggregation_bits,
//        inclusion_delay=state.slot - data.slot,
//        proposer_index=get_beacon_proposer_index(state),
//    )
//
//    if data.target.epoch == get_current_epoch(state):
//        assert data.source == state.current_justified_checkpoint
//        state.current_epoch_attestations.append(pending_attestation)
//    else:
//        assert data.source == state.previous_justified_checkpoint
//        state.previous_epoch_attestations.append(pending_attestation)
func appendPendingAttestation(state *core.State, attestation *core.Attestation) error {
	proposer, err := shared.GetBlockProposerIndex(state)
	if err != nil {
		return err
	}
	pendingAtt := &core.PendingAttestation{
		AggregationBits:      attestation.AggregationBits,
		Data:                 attestation.Data,
		InclusionDelay:       state.CurrentSlot - attestation.Data.Slot,
		ProposerIndex:        proposer,
	}

	if attestation.Data.Target.Epoch == shared.ComputeEpochAtSlot(state.CurrentSlot) {
		if !core.CheckpointsEqual(attestation.Data.Source, state.CurrentJustifiedCheckpoint) {
			return fmt.Errorf("source doesn't equal current justified checkpoint")
		}
		state.CurrentEpochAttestations = append(state.CurrentEpochAttestations, pendingAtt)
	} else {
		if !core.CheckpointsEqual(attestation.Data.Source, state.PreviousJustifiedCheckpoint) {
			return fmt.Errorf("source doesn't equal previous justified checkpoint")
		}
		state.PreviousEpochAttestations = append(state.PreviousEpochAttestations, pendingAtt)
	}
	return nil
}

func validateAggregationBits(state *core.State, attestation *core.Attestation) error {
	expectedCommittee, err := shared.GetAttestationCommittee(state, attestation.Data.Slot, uint64(attestation.Data.CommitteeIndex))
	if err != nil {
		return err
	}
	if len(expectedCommittee) != len(attestation.AggregationBits) {
		return fmt.Errorf("aggregation bits != committee size")
	}
	return nil
}

//    assert data.target.epoch in (get_previous_epoch(state), get_current_epoch(state))
//    assert data.target.epoch == compute_epoch_at_slot(data.slot)
//    assert data.slot + MIN_ATTESTATION_INCLUSION_DELAY <= state.slot <= data.slot + SLOTS_PER_EPOCH
//    assert data.index < get_committee_count_per_slot(state, data.target.epoch)
func validateAttestationData(state *core.State, data *core.AttestationData) error {
	currentEpoch := shared.GetCurrentEpoch(state)
	previousEpoch := shared.GetPreviousEpoch(state)

	if data.Target.Epoch != currentEpoch && data.Target.Epoch != previousEpoch {
		return fmt.Errorf("taregt not in current/ previous epoch")
	}

	if shared.ComputeEpochAtSlot(data.Slot) != data.Target.Epoch {
		return fmt.Errorf("target slot not in the correct epoch")
	}

	if data.Slot + params.ChainConfig.MinAttestationInclusionDelay > state.CurrentSlot {
		return fmt.Errorf("min att. inclusion delay did not pass")
	}
	if state.CurrentSlot > data.Slot + params.ChainConfig.SlotsInEpoch {
		return fmt.Errorf("slot to submit att. has passed")
	}

	if data.CommitteeIndex >= shared.GetCommitteeCountPerSlot(state, data.Slot) {
		return fmt.Errorf("committee index out of range")
	}

	return nil
}

//    # Check signature
//    assert is_valid_indexed_attestation(state, get_indexed_attestation(state, attestation))
func validateAttestationSignature(state *core.State, attestation *core.Attestation, slot uint64) error {
	// reconstruct committee
	expectedCommittee, err := shared.GetAttestationCommittee(state, attestation.Data.Slot, uint64(attestation.Data.CommitteeIndex))
	if err != nil {
		return err
	}

	// get pubkeys by aggregation bits
	pks := make([]bls.PublicKey,0)

	for i, id := range expectedCommittee {
		bp := shared.GetBlockProducer(state, id)
		if bp == nil {
			return fmt.Errorf("BP %d is inactivee ", id)
		}

		// deserialize pk and aggregate
		if attestation.AggregationBits.BitAt(uint64(i)) {
			pk := bls.PublicKey{}
			err := pk.Deserialize(bp.PubKey)
			if err != nil {
				return err
			}
			pks = append(pks, pk)
		}
	}

	// threshold passed
	if len(expectedCommittee) * 2 > 3 * len(pks) {
		return fmt.Errorf("attestation did not pass threshold")
	}

	// verify
	sig := &bls.Sign{}
	err = sig.Deserialize(attestation.Signature)
	if err != nil {
		return err
	}
	root, err := ssz.HashTreeRoot(attestation.Data)
	if err != nil {
		return err
	}
	res := sig.FastAggregateVerify(pks, root[:])
	if !res {
		return fmt.Errorf("attestation signature not vrified")
	}

	return nil
}
