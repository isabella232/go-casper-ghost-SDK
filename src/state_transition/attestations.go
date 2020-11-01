package state_transition

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
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

	//    # Check signature
	//    assert is_valid_indexed_attestation(state, get_indexed_attestation(state, attestation))
	indexedAttestation, err := shared.GetIndexedAttestation(state, attestation)
	if err != nil {
		return err
	}
	if err := isValidIndexedAttestation(state, indexedAttestation); err != nil {
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
		InclusionDelay:       state.Slot - attestation.Data.Slot,
		ProposerIndex:        proposer,
	}

	if attestation.Data.Target.Epoch == shared.ComputeEpochAtSlot(state.Slot) {
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
	expectedCommittee, err := shared.GetBeaconCommittee(state, attestation.Data.Slot, uint64(attestation.Data.CommitteeIndex))
	if err != nil {
		return err
	}
	if uint64(len(expectedCommittee)) != attestation.AggregationBits.Len() {
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

	if data.Slot + params.ChainConfig.MinAttestationInclusionDelay > state.Slot {
		return fmt.Errorf("min att. inclusion delay did not pass")
	}
	if state.Slot > data.Slot + params.ChainConfig.SlotsInEpoch {
		return fmt.Errorf("slot to submit att. has passed")
	}

	if data.CommitteeIndex >= shared.GetCommitteeCountPerSlot(state, data.Slot) {
		return fmt.Errorf("committee index out of range")
	}

	return nil
}

//    # Check signature
//    assert is_valid_indexed_attestation(state, get_indexed_attestation(state, attestation))
/**
def is_valid_indexed_attestation(state: BeaconState, indexed_attestation: IndexedAttestation) -> bool:
    """
    Check if ``indexed_attestation`` is not empty, has sorted and unique indices and has a valid aggregate signature.
    """
    # Verify indices are sorted and unique
    indices = indexed_attestation.attesting_indices
    if len(indices) == 0 or not indices == sorted(set(indices)):
        return False
    # Verify aggregate signature
    pubkeys = [state.validators[i].pubkey for i in indices]
    domain = get_domain(state, DOMAIN_BEACON_ATTESTER, indexed_attestation.data.target.epoch)
    signing_root = compute_signing_root(indexed_attestation.data, domain)
    return bls.FastAggregateVerify(pubkeys, signing_root, indexed_attestation.signature)
 */
func isValidIndexedAttestation(state *core.State, attestation *core.IndexedAttestation) error {
	if len(attestation.AttestingIndices) == 0 {
		return fmt.Errorf("attesting indices length is 0")
	}

	// get pubkeys by aggregation bits
	pks := make([]bls.PublicKey,0)
	for _, id := range attestation.AttestingIndices {
		validator := shared.GetValidator(state, id)
		if validator == nil {
			return fmt.Errorf("BP %d is inactivee ", id)
		}

		pk := bls.PublicKey{}
		err := pk.Deserialize(validator.PublicKey)
		if err != nil {
			return err
		}
		pks = append(pks, pk)
	}

	// sig
	sig := &bls.Sign{}
	err := sig.Deserialize(attestation.Signature)
	if err != nil {
		return err
	}

	// domain
	domain, err := shared.GetDomain(state, params.ChainConfig.DomainBeaconAttester, attestation.Data.Target.Epoch)
	if err != nil {
		return err
	}

	// root
	root, err := shared.ComputeSigningRoot(attestation.Data, domain)
	if err != nil {
		return err
	}

	// verify sig
	res := sig.FastAggregateVerify(pks, root[:])
	if !res {
		return fmt.Errorf("attestation signature not vrified")
	}

	return nil
}
