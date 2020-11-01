package shared

import (
	"bytes"
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-bitfield"
	"sort"
)

/**
	def is_slashable_attestation_data(data_1: AttestationData, data_2: AttestationData) -> bool:
		"""
		Check if ``data_1`` and ``data_2`` are slashable according to Casper FFG rules.
		"""
		return (
			# Double vote
			(data_1 != data_2 and data_1.target.epoch == data_2.target.epoch) or
			# Surround vote
			(data_1.source.epoch < data_2.source.epoch and data_2.target.epoch < data_1.target.epoch)
		)
 */
func IsSlashableAttestationData (att1 *core.AttestationData, att2 *core.AttestationData) bool {
	return (!core.AttestationDataEqual(att1, att2) && att1.Target.Epoch == att2.Target.Epoch) || // double
		(																						// surround
			(att1.Source.Epoch < att2.Source.Epoch && att2.Target.Epoch < att1.Target.Epoch) ||
				(att2.Source.Epoch < att1.Source.Epoch && att1.Target.Epoch < att2.Target.Epoch))

}

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
// TODO - is_valid_indexed_attestation
 */
func IsValidIndexedAttestation(state *core.State, attestation *core.IndexedAttestation) (bool, error) {
	validateIndices := func (indices []uint64) error {
		if len(indices) == 0 {
			return fmt.Errorf("indices length 0")
		}
		if uint64(len(indices)) > params.ChainConfig.MaxValidatorsPerCommittee {
			return fmt.Errorf("committee indices count larger than MaxValidatorsPerCommittee")
		}
		for i := 1; i < len(indices); i++ {
			if indices[i-1] >= indices[i] {
				return fmt.Errorf("attesting indices is not uniquely sorted")
			}
		}
		return nil
	}

	// Verify indices are sorted and unique
	indices := attestation.AttestingIndices
	if err := validateIndices(indices); err != nil {
		return false, err
	}

	// Verify aggregate signature
	pks := []bls.PublicKey{}
	for _, index := range indices {
		validator := GetValidator(state, index)
		if validator == nil {
			return false, fmt.Errorf("BP not found")
		}
		pk := bls.PublicKey{}
		err := pk.Deserialize(validator.PublicKey)
		if err != nil {
			return false, err
		}
		pks = append(pks, pk)
	}

	domain, err := GetDomain(state, params.ChainConfig.DomainBeaconAttester, attestation.Data.Target.Epoch)
	if err != nil {
		return false, err
	}
	root, err :=  ComputeSigningRoot(attestation.Data, domain)
	if err != nil {
		return false, err
	}

	sig := &bls.Sign{}
	err = sig.Deserialize(attestation.Signature)
	if err != nil {
		return false, err
	}

	// verify sig
	res := sig.FastAggregateVerify(pks, root[:])
	if !res {
		return false, fmt.Errorf("indexed attestation signature not vrified")
	}
	return true, nil
}

/**
def compute_committee(indices: Sequence[ValidatorIndex],
                      seed: Bytes32,
                      index: uint64,
                      count: uint64) -> Sequence[ValidatorIndex]:
    """
    Return the committee corresponding to ``indices``, ``seed``, ``index``, and committee ``count``.
    """
    start = (len(indices) * index) // count
    end = (len(indices) * uint64(index + 1)) // count
    return [indices[compute_shuffled_index(uint64(i), uint64(len(indices)), seed)] for i in range(start, end)]
 */
func ComputeCommittee(indices []uint64, seed [32]byte, index uint64, count uint64) ([]uint64, error) {
	start := uint64(len(indices)) * index / count
	end := uint64(len(indices)) * uint64(index + 1) / count

	unshuffled, err := UnshuffleList(indices, seed)
	if err != nil {
		return nil, err
	}
	return unshuffled[start:end], nil

	//ret := []uint64{}
	//for i := start ; i < end ; i++ {
	//	idx, err := computeShuffledIndex(i, uint64(len(indices)), seed, true, params.ChainConfig.ShuffleRoundCount)
	//	if err != nil {
	//		return []uint64{}, err
	//	}
	//
	//	ret = append(ret, idx)
	//}
	//return ret, nil
}

/**
def get_committee_count_per_slot(state: BeaconState, epoch: Epoch) -> uint64:
    """
    Return the number of committees in each slot for the given ``epoch``.
    """
    return max(uint64(1), min(
        MAX_COMMITTEES_PER_SLOT,
        uint64(len(get_active_validator_indices(state, epoch))) // SLOTS_PER_EPOCH // TARGET_COMMITTEE_SIZE,
    ))
 */
func GetCommitteeCountPerSlot(state *core.State, slot uint64) uint64 {
	epoch := ComputeEpochAtSlot(slot)
	validators := GetActiveValidators(state, epoch)
	committeePerSlot := uint64(len(validators)) / params.ChainConfig.SlotsInEpoch / params.ChainConfig.TargetCommitteeSize

	if committeePerSlot > params.ChainConfig.MaxCommitteesPerSlot {
		return params.ChainConfig.MaxCommitteesPerSlot
	}
	if committeePerSlot == 0 {
		return 1
	}
	return committeePerSlot
}

/**
def get_beacon_committee(state: BeaconState, slot: Slot, index: CommitteeIndex) -> Sequence[ValidatorIndex]:
    """
    Return the beacon committee at ``slot`` for ``index``.
    """
    epoch = compute_epoch_at_slot(slot)
    committees_per_slot = get_committee_count_per_slot(state, epoch)
    return compute_committee(
        indices=get_active_validator_indices(state, epoch),
        seed=get_seed(state, epoch, DOMAIN_BEACON_ATTESTER),
        index=(slot % SLOTS_PER_EPOCH) * committees_per_slot + index,
        count=committees_per_slot * SLOTS_PER_EPOCH,
    )
 */
func GetBeaconCommittee(state *core.State, slot uint64, index uint64) ([]uint64, error) {
	epoch := ComputeEpochAtSlot(slot)
	committeesPerSlot := GetCommitteeCountPerSlot(state, slot)
	seed := GetSeed(state, epoch, params.ChainConfig.DomainBeaconAttester)

	active := GetActiveValidators(state, epoch)
	if false {
		return active, nil
	}

	return ComputeCommittee(
			GetActiveValidators(state, epoch),
			seed,
			(slot % params.ChainConfig.SlotsInEpoch) * committeesPerSlot + index,
			committeesPerSlot * params.ChainConfig.SlotsInEpoch,
		)
}

/**
def get_indexed_attestation(state: BeaconState, attestation: Attestation) -> IndexedAttestation:
    """
    Return the indexed attestation corresponding to ``attestation``.
    """
    attesting_indices = get_attesting_indices(state, attestation.data, attestation.aggregation_bits)

    return IndexedAttestation(
        attesting_indices=sorted(attesting_indices),
        data=attestation.data,
        signature=attestation.signature,
    )
 */
func GetIndexedAttestation(state *core.State, attestation *core.Attestation) (*core.IndexedAttestation, error) {
	indices, err := GetAttestingIndices(state, attestation.Data, attestation.AggregationBits)
	if err != nil {
		return nil, err
	}

	return &core.IndexedAttestation{
		AttestingIndices:     indices,
		Data:                 attestation.Data,
		Signature:            attestation.Signature,
	}, nil
}

/**
def get_attesting_indices(state: BeaconState,
                          data: AttestationData,
                          bits: Bitlist[MAX_VALIDATORS_PER_COMMITTEE]) -> Set[ValidatorIndex]:
    """
    Return the set of attesting indices corresponding to ``data`` and ``bits``.
    """
    committee = get_beacon_committee(state, data.slot, data.index)
    return set(index for i, index in enumerate(committee) if bits[i])
 */
func GetAttestingIndices(state *core.State, data *core.AttestationData, bits bitfield.Bitlist) ([]uint64, error) {
	committee, err := GetBeaconCommittee(state, data.Slot, data.CommitteeIndex)
	if err != nil {
		return nil, err
	}
	ret := make([]uint64,0,bits.Count())
	for _, idx := range bits.BitIndices() {
		if idx < len(committee) {
			ret = append(ret, committee[idx])
		}
	}
	return ret, nil
}

/**
def get_unslashed_attesting_indices(state: BeaconState,
                                    attestations: Sequence[PendingAttestation]) -> Set[ValidatorIndex]:
    output = set()  # type: Set[ValidatorIndex]
    for a in attestations:
        output = output.union(get_attesting_indices(state, a.data, a.aggregation_bits))
    return set(filter(lambda index: not state.validators[index].slashed, output))
 */
func GetUnslashedAttestingIndices(state *core.State, attestations []*core.PendingAttestation) ([]uint64, error) {
	output := make([]uint64,0)
	seen := make(map[uint64]bool)

	for _, a := range attestations {
		indices, err := GetAttestingIndices(state, a.Data, a.AggregationBits)
		if err != nil {
			return nil, err
		}
		for _, i := range indices {
			if !seen[i] {
				output = append(output, i)
			}
			seen[i] = true
		}
	}

	// Sort the attesting set indices by increasing order.
	sort.Slice(output, func(i, j int) bool {
		return output[i] < output[j]
	})

	// Remove slashed validator indices.
	ret := make([]uint64, 0)
	for i := range output {
		val := GetValidator(state, output[i])
		if val != nil && !val.Slashed {
			ret = append(ret, output[i])
		}
	}
	return ret, nil
}

/**
def get_matching_source_attestations(state: BeaconState, epoch: Epoch) -> Sequence[PendingAttestation]:
    assert epoch in (get_previous_epoch(state), get_current_epoch(state))
    return state.current_epoch_attestations if epoch == get_current_epoch(state) else state.previous_epoch_attestations
 */
func GetMatchingSourceAttestations(state *core.State, epoch uint64) ([]*core.PendingAttestation, error) {
	if epoch != GetPreviousEpoch(state) && epoch != GetCurrentEpoch(state) {
		return nil, fmt.Errorf("epoch not current nor previous")
	}
	if epoch == GetCurrentEpoch(state) {
		return state.CurrentEpochAttestations, nil
	} else {
		return state.PreviousEpochAttestations, nil
	}
}


/**
def get_matching_target_attestations(state: BeaconState, epoch: Epoch) -> Sequence[PendingAttestation]:
    return [
        a for a in get_matching_source_attestations(state, epoch)
        if a.data.target.root == get_block_root(state, epoch)
    ]
 */
func GetMatchingTargetAttestations(state *core.State, epoch uint64) ([]*core.PendingAttestation, error) {
	source, err := GetMatchingSourceAttestations(state, epoch)
	if err != nil {
		return nil, err
	}

	ret := make([]*core.PendingAttestation,0)

	if len(source) == 0 {
		return ret, nil
	}

	targetRoot, err := GetBlockRoot(state, epoch)
	if err != nil {
		return nil, err
	}
	for _, att := range source {
		if bytes.Equal(att.Data.Target.Root, targetRoot) {
			ret = append(ret, att)
		}
	}
	return ret, nil
}

/**
def get_matching_head_attestations(state: BeaconState, epoch: Epoch) -> Sequence[PendingAttestation]:
    return [
        a for a in get_matching_target_attestations(state, epoch)
        if a.data.beacon_block_root == get_block_root_at_slot(state, a.data.slot)
    ]
 */
func GetMatchingHeadAttestations(state *core.State, epoch uint64) ([]*core.PendingAttestation, error) {
	source, err := GetMatchingTargetAttestations(state, epoch)
	if err != nil {
		return nil, err
	}

	ret := []*core.PendingAttestation{}
	for _, att := range source {
		root, err := GetBlockRootAtSlot(state, att.Data.Slot)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(att.Data.BeaconBlockRoot, root) {
			ret = append(ret, att)
		}
	}
	return ret, nil
}