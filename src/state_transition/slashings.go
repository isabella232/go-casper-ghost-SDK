package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

func ProcessProposerSlashings(state *core.State, slashings []*core.ProposerSlashing) error {
	for _, s := range slashings {
		if err := processProposerSlashing(state, s); err != nil {
			return err
		}
	}
	return nil
}

/**
def process_proposer_slashing(state: BeaconState, proposer_slashing: ProposerSlashing) -> None:
    header_1 = proposer_slashing.signed_header_1.message
    header_2 = proposer_slashing.signed_header_2.message

    # Verify header slots match
    assert header_1.slot == header_2.slot
    # Verify header proposer indices match
    assert header_1.proposer_index == header_2.proposer_index
    # Verify the headers are different
    assert header_1 != header_2
    # Verify the proposer is slashable
    proposer = state.validators[header_1.proposer_index]
    assert is_slashable_validator(proposer, get_current_epoch(state))
    # Verify signatures
    for signed_header in (proposer_slashing.signed_header_1, proposer_slashing.signed_header_2):
        domain = get_domain(state, DOMAIN_BEACON_PROPOSER, compute_epoch_at_slot(signed_header.message.slot))
        signing_root = compute_signing_root(signed_header.message, domain)
        assert bls.Verify(proposer.pubkey, signing_root, signed_header.signature)

    slash_validator(state, header_1.proposer_index)
*/
func processProposerSlashing(state *core.State, slashing *core.ProposerSlashing) error {
	header1 := slashing.Header_1.Header
	header2 := slashing.Header_2.Header

	if header1 == nil || header2 == nil {
		return fmt.Errorf("proposer slashing: one of the headers (or both) are nil")
	}

	// Verify header slots match
	if header1.Slot != header2.Slot {
		return fmt.Errorf("proposer slashing: slots not equal")
	}
	// Verify header proposer indices match
	if header1.ProposerIndex != header2.ProposerIndex {
		return fmt.Errorf("proposer slashing: proposer idx not equal")
	}
	// Verify the headers are different
	if core.BlockHeaderEqual(header1, header2) {
		return fmt.Errorf("proposer slashing: block headers are equal")
	}
	// Verify the proposer is slashable
	proposer := shared.GetBlockProducer(state, header1.ProposerIndex)
	if proposer == nil {
		return fmt.Errorf("proposer slashing: block producer not found")
	}
	if !shared.IsSlashableBp(proposer, shared.GetCurrentEpoch(state)) {
		return fmt.Errorf("proposer slashing: BP not slashable at epoch %d", shared.GetCurrentEpoch(state))
	}
	// Verify signatures
	for _, sig := range []*core.SignedPoolBlockHeader{slashing.Header_1, slashing.Header_2} {
		domain, err := shared.GetDomain(state, params.ChainConfig.DomainBeaconProposer, shared.ComputeEpochAtSlot(sig.Header.Slot))
		if err != nil {
			return err
		}
		root, err := shared.ComputeSigningRoot(sig.Header, domain)
		if err != nil {
			return err
		}
		res, err := shared.VerifySignature(root[:], proposer.PubKey, sig.Signature)
		if err != nil {
			return err
		}
		if !res {
			return fmt.Errorf("proposer slashing: sig not verified for proposer %d", proposer.Id)
		}
	}
	return nil
}

func ProcessAttesterSlashings(state *core.State, slashings []*core.AttesterSlashing) error {
	for _, s := range slashings {
		if err := ProcessAttesterSlashing(state, s); err != nil {
			return err
		}
	}
	return nil
}

/**
def process_attester_slashing(state: BeaconState, attester_slashing: AttesterSlashing) -> None:
    attestation_1 = attester_slashing.attestation_1
    attestation_2 = attester_slashing.attestation_2
    assert is_slashable_attestation_data(attestation_1.data, attestation_2.data)
    assert is_valid_indexed_attestation(state, attestation_1)
    assert is_valid_indexed_attestation(state, attestation_2)

    slashed_any = False
    indices = set(attestation_1.attesting_indices).intersection(attestation_2.attesting_indices)
    for index in sorted(indices):
        if is_slashable_validator(state.validators[index], get_current_epoch(state)):
            slash_validator(state, index)
            slashed_any = True
    assert slashed_any
*/
func ProcessAttesterSlashing(state *core.State, slashing *core.AttesterSlashing) error {
	attestation1 := slashing.Attestation_1
	attestation2 := slashing.Attestation_2

	// asserts
	if !shared.IsSlashableAttestationData(attestation1.Data, attestation2.Data) {
		return fmt.Errorf("attester slashing: attestation data not eqal")
	}
	if res, err := shared.IsValidIndexedAttestation(state, slashing.Attestation_1); !res || err != nil {
		if err != nil {
			return fmt.Errorf("attester slashing: att. 1 not valid %s", err.Error())
		}
		return fmt.Errorf("attester slashing: att. 1 not valid")
	}
	if res, err := shared.IsValidIndexedAttestation(state, slashing.Attestation_2); !res || err != nil {
		if err != nil {
			return fmt.Errorf("attester slashing: att. 2 not valid %s", err.Error())
		}
		return fmt.Errorf("attester slashing: att. 2 not valid")
	}


	slashedAny := false
	indices := slashableAttesterIndices(slashing)
	for _, index := range indices {
		bp := shared.GetBlockProducer(state, index)
		if bp == nil {
			return fmt.Errorf("attester slashing: BP %d not found", index)
		}
		if shared.IsSlashableBp(bp, shared.GetCurrentEpoch(state)) {
			if err := shared.SlashBlockProducer(state, index); err != nil {
				return err
			}
			slashedAny = true
		}
	}
	if !slashedAny {
		return fmt.Errorf("attester slashing: unable to slash any validator despite confirmed attester slashing")
	}

	return nil
}

func slashableAttesterIndices(slashing *core.AttesterSlashing) []uint64 {
	if slashing == nil || slashing.Attestation_1 == nil || slashing.Attestation_2 == nil {
		return nil
	}
	indices1 := slashing.Attestation_1.AttestingIndices
	indices2 := slashing.Attestation_2.AttestingIndices
	return shared.IntersectionUint64(indices1, indices2)
}