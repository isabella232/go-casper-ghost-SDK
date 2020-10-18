package shared

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

/**
def get_total_balance(state: BeaconState, indices: Set[ValidatorIndex]) -> Gwei:
    """
    Return the combined effective balance of the ``indices``.
    ``EFFECTIVE_BALANCE_INCREMENT`` Gwei minimum to avoid divisions by zero.
    Math safe up to ~10B ETH, afterwhich this overflows uint64.
    """
    return Gwei(max(EFFECTIVE_BALANCE_INCREMENT, sum([state.validators[index].effective_balance for index in indices])))
 */
func GetTotalStake(state *core.State, indices []uint64) uint64 {
	sum := uint64(0)
	for _, index := range indices {
		bp := GetBlockProducer(state, index)
		if bp != nil {
			sum += bp.EffectiveBalance
		}
	}

	if sum < params.ChainConfig.EffectiveBalanceIncrement {
		return params.ChainConfig.EffectiveBalanceIncrement
	}
	return sum
}

/**
def get_total_active_balance(state: BeaconState) -> Gwei:
    """
    Return the combined effective balance of the active validators.
    Note: ``get_total_balance`` returns ``EFFECTIVE_BALANCE_INCREMENT`` Gwei minimum to avoid divisions by zero.
    """
    return get_total_balance(state, set(get_active_validator_indices(state, get_current_epoch(state))))
 */
func GetTotalActiveStake(state *core.State) uint64 {
	indices := GetActiveBlockProducers(state, GetCurrentEpoch(state))
	return GetTotalStake(state, indices)
}

/**
def get_base_reward(state: BeaconState, index: ValidatorIndex) -> Gwei:
    total_balance = get_total_active_balance(state)
    effective_balance = state.validators[index].effective_balance
    return Gwei(effective_balance * BASE_REWARD_FACTOR // integer_squareroot(total_balance) // BASE_REWARDS_PER_EPOCH)
 */
func GetBaseReward(state *core.State, index uint64) (uint64, error) {
	totalBalance := GetTotalActiveStake(state)
	if bp := GetBlockProducer(state, index); bp != nil {
		effectiveBalance := bp.EffectiveBalance
		return effectiveBalance * params.ChainConfig.BaseRewardFactor / IntegerSquareRoot(totalBalance), nil
	} else {
		return 0, fmt.Errorf("could not find BP %d", index)
	}
}

/**
def get_proposer_reward(state: BeaconState, attesting_index: ValidatorIndex) -> Gwei:
    return Gwei(get_base_reward(state, attesting_index) // PROPOSER_REWARD_QUOTIENT)
 */
func GetProposerReward(state *core.State, attestingIndex uint64)(uint64, error) {
	base, err := GetBaseReward(state, attestingIndex)
	if err != nil {
		return 0, err
	}
	return base / params.ChainConfig.ProposerRewardQuotient, nil
}

/**
def get_finality_delay(state: BeaconState) -> uint64:
    return get_previous_epoch(state) - state.finalized_checkpoint.epoch
 */
func GetFinalityDelay(state *core.State) uint64 {
	return GetPreviousEpoch(state) - state.FinalizedCheckpoint.Epoch
}

/**
def is_in_inactivity_leak(state: BeaconState) -> bool:
    return get_finality_delay(state) > MIN_EPOCHS_TO_INACTIVITY_PENALTY
 */
func IsInInactivityLeak(state *core.State) bool {
	return GetFinalityDelay(state) > params.ChainConfig.MinEpochsToInactivityPenalty
}

/**
def get_eligible_validator_indices(state: BeaconState) -> Sequence[ValidatorIndex]:
    previous_epoch = get_previous_epoch(state)
    return [
        ValidatorIndex(index) for index, v in enumerate(state.validators)
        if is_active_validator(v, previous_epoch) or (v.slashed and previous_epoch + 1 < v.withdrawable_epoch)
    ]
 */
func GetEligibleBpIndices(state *core.State) []uint64 {
	ret := []uint64{}
	prevEpoch := GetPreviousEpoch(state)
	for _, bp := range state.BlockProducers {
		if IsActiveBP(bp, prevEpoch) || (bp.Slashed && prevEpoch + 1 < bp.WithdrawableEpoch) {
			ret = append(ret, bp.Id)
		}
	}
	return ret
}

/**
def get_attestation_component_deltas(state: BeaconState,
                                     attestations: Sequence[PendingAttestation]
                                     ) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Helper with shared logic for use by get source, target, and head deltas functions
    """
    rewards = [Gwei(0)] * len(state.validators)
    penalties = [Gwei(0)] * len(state.validators)
    total_balance = get_total_active_balance(state)
    unslashed_attesting_indices = get_unslashed_attesting_indices(state, attestations)
    attesting_balance = get_total_balance(state, unslashed_attesting_indices)
    for index in get_eligible_validator_indices(state):
        if index in unslashed_attesting_indices:
            increment = EFFECTIVE_BALANCE_INCREMENT  # Factored out from balance totals to avoid uint64 overflow
            if is_in_inactivity_leak(state):
                # Since full base reward will be canceled out by inactivity penalty deltas,
                # optimal participation receives full base reward compensation here.
                rewards[index] += get_base_reward(state, index)
            else:
                reward_numerator = get_base_reward(state, index) * (attesting_balance // increment)
                rewards[index] += reward_numerator // (total_balance // increment)
        else:
            penalties[index] += get_base_reward(state, index)
    return rewards, penalties
 */
func GetAttestationComponentDeltas(state *core.State, attestations []*core.PendingAttestation) ([]uint64, []uint64, error) {
	rewards := uint64ZeroArray(uint64(len(state.BlockProducers)))
	penalties := uint64ZeroArray(uint64(len(state.BlockProducers)))
	totalStake := GetTotalActiveStake(state)
	unslashedAttestingIndices, err := GetUnslashedAttestingIndices(state, attestations)
	if err != nil {
		return nil, nil, err
	}
	unslashedAttestingIndicesMap := make(map[uint64]bool)
	for _, i := range unslashedAttestingIndices {
		unslashedAttestingIndicesMap[i] = true
	}

	attestingBalance := GetTotalStake(state, unslashedAttestingIndices)

	for _, index := range GetEligibleBpIndices(state) {
		if unslashedAttestingIndicesMap[index] {
			increment := params.ChainConfig.EffectiveBalanceIncrement
			base, err := GetBaseReward(state, index)
			if err != nil {
				return nil, nil, err
			}
			if IsInInactivityLeak(state) {
				rewards[index] += base
			} else {
				rewardNumerator := base * (attestingBalance / increment)
				rewards[index] += rewardNumerator / (totalStake / increment)
			}
		} else {
			base, err := GetBaseReward(state, index)
			if err != nil {
				return nil, nil, err
			}
			penalties[index] += base
		}
	}
	return rewards, penalties, nil
}

/**
def get_source_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return attester micro-rewards/penalties for source-vote for each validator.
    """
    matching_source_attestations = get_matching_source_attestations(state, get_previous_epoch(state))
    return get_attestation_component_deltas(state, matching_source_attestations)
 */
func GetSourceDeltas(state *core.State) ([]uint64, []uint64, error) {
	matchingSourceAttestations, err := GetMatchingSourceAttestations(state, GetPreviousEpoch(state))
	if err != nil {
		return nil, nil, err
	}

	return GetAttestationComponentDeltas(state, matchingSourceAttestations)
}

/**
def get_target_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return attester micro-rewards/penalties for target-vote for each validator.
    """
    matching_target_attestations = get_matching_target_attestations(state, get_previous_epoch(state))
    return get_attestation_component_deltas(state, matching_target_attestations)
 */
func GetTargetDeltas(state *core.State) ([]uint64, []uint64, error) {
	matchingTargetAttestations, err := GetMatchingTargetAttestations(state, GetPreviousEpoch(state))
	if err != nil {
		return nil, nil, err
	}

	return GetAttestationComponentDeltas(state, matchingTargetAttestations)
}

/**
def get_head_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return attester micro-rewards/penalties for head-vote for each validator.
    """
    matching_head_attestations = get_matching_head_attestations(state, get_previous_epoch(state))
    return get_attestation_component_deltas(state, matching_head_attestations)
 */
func GetHeadDeltas(state *core.State) ([]uint64, []uint64, error) {
	matchingHeadAttestations, err := GetMatchingHeadAttestations(state, GetPreviousEpoch(state))
	if err != nil {
		return nil, nil, err
	}

	return GetAttestationComponentDeltas(state, matchingHeadAttestations)
}

/**
def get_inclusion_delay_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return proposer and inclusion delay micro-rewards/penalties for each validator.
    """
    rewards = [Gwei(0) for _ in range(len(state.validators))]
    matching_source_attestations = get_matching_source_attestations(state, get_previous_epoch(state))
    for index in get_unslashed_attesting_indices(state, matching_source_attestations):
        attestation = min([
            a for a in matching_source_attestations
            if index in get_attesting_indices(state, a.data, a.aggregation_bits)
        ], key=lambda a: a.inclusion_delay)
        rewards[attestation.proposer_index] += get_proposer_reward(state, index)
        max_attester_reward = Gwei(get_base_reward(state, index) - get_proposer_reward(state, index))
        rewards[index] += Gwei(max_attester_reward // attestation.inclusion_delay)

    # No penalties associated with inclusion delay
    penalties = [Gwei(0) for _ in range(len(state.validators))]
    return rewards, penalties
 */
func GetInclusionDelayDeltas(state *core.State) ([]uint64, []uint64, error) {
	rewards := uint64ZeroArray(uint64(len(state.BlockProducers)))
	matchingSourceAttestations, err := GetMatchingSourceAttestations(state, GetPreviousEpoch(state))
	if err != nil {
		return nil, nil, err
	}
	unslashed, err := GetUnslashedAttestingIndices(state, matchingSourceAttestations)
	if err != nil {
		return nil, nil, err
	}

	findMinInclusion := func(index uint64, matchingSourceAttestations []*core.PendingAttestation) (*core.PendingAttestation, error) {
		var ret *core.PendingAttestation
		for _, a := range matchingSourceAttestations {
			attIndices, err := GetAttestingIndices(state, a.Data, a.AggregationBits)
			if err != nil {
				return nil, err
			}
			for _, ai := range attIndices {
				if index == ai {
					if ret == nil {
						ret = a
					}
					if ret.InclusionDelay > a.InclusionDelay {
						ret = a
					}
				}
			}
		}
		return ret, nil
	}

	for _, index := range unslashed {
		min, err := findMinInclusion(index, matchingSourceAttestations)
		if err != nil {
			return nil, nil, err
		}
		rew, err := GetProposerReward(state, index)
		if err != nil {
			return nil, nil, err
		}
		base, err := GetBaseReward(state, index)
		if err != nil {
			return nil, nil, err
		}
		rewards[min.ProposerIndex] += rew
		maxAttesterReward := base - rew
		rewards[index] += maxAttesterReward / min.InclusionDelay
	}

	// No penalties associated with inclusion delay
	penalties := uint64ZeroArray(uint64(len(state.BlockProducers)))
	return rewards, penalties, nil
}

/**
def get_inactivity_penalty_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return inactivity reward/penalty deltas for each validator.
    """
    penalties = [Gwei(0) for _ in range(len(state.validators))]
    if is_in_inactivity_leak(state):
        matching_target_attestations = get_matching_target_attestations(state, get_previous_epoch(state))
        matching_target_attesting_indices = get_unslashed_attesting_indices(state, matching_target_attestations)
        for index in get_eligible_validator_indices(state):
            # If validator is performing optimally this cancels all rewards for a neutral balance
            base_reward = get_base_reward(state, index)
            penalties[index] += Gwei(BASE_REWARDS_PER_EPOCH * base_reward - get_proposer_reward(state, index))
            if index not in matching_target_attesting_indices:
                effective_balance = state.validators[index].effective_balance
                penalties[index] += Gwei(effective_balance * get_finality_delay(state) // INACTIVITY_PENALTY_QUOTIENT)

    # No rewards associated with inactivity penalties
    rewards = [Gwei(0) for _ in range(len(state.validators))]
    return rewards, penalties
 */
func GetInactivityPenaltyDeltas(state *core.State) ([]uint64, []uint64, error) {
	penalties := uint64ZeroArray(uint64(len(state.BlockProducers)))
	if IsInInactivityLeak(state) {
		matchingTargetAttestations, err := GetMatchingTargetAttestations(state, GetPreviousEpoch(state))
		if err != nil {
			return nil, nil, err
		}
		matchingTargetAttestingIndices, err := GetUnslashedAttestingIndices(state, matchingTargetAttestations)
		if err != nil {
			return nil, nil, err
		}
		matchingTargetAttestingIndicesMap := make(map[uint64]bool)
		for _, i := range matchingTargetAttestingIndices {
			matchingTargetAttestingIndicesMap[i] = true
		}
		for _, index := range GetEligibleBpIndices(state) {
			// If validator is performing optimally this cancels all rewards for a neutral balance
			base, err := GetBaseReward(state, index)
			if err != nil {
				return nil, nil, err
			}
			proposer, err := GetProposerReward(state, index)
			if err != nil {
				return nil, nil, err
			}
			penalties[index] += params.ChainConfig.BaseRewardsPerEpoch * base - proposer
			if !matchingTargetAttestingIndicesMap[index] {
				effectiveBalance := GetBlockProducer(state, index).EffectiveBalance
				penalties[index] += effectiveBalance * GetFinalityDelay(state) / params.ChainConfig.InactivityPenaltyQuotient
			}
		}
	}

	// No rewards associated with inactivity penalties
	rewards := uint64ZeroArray(uint64(len(state.BlockProducers)))
	return rewards, penalties, nil
}

/**
def get_attestation_deltas(state: BeaconState) -> Tuple[Sequence[Gwei], Sequence[Gwei]]:
    """
    Return attestation reward/penalty deltas for each validator.
    """
    source_rewards, source_penalties = get_source_deltas(state)
    target_rewards, target_penalties = get_target_deltas(state)
    head_rewards, head_penalties = get_head_deltas(state)
    inclusion_delay_rewards, _ = get_inclusion_delay_deltas(state)
    _, inactivity_penalties = get_inactivity_penalty_deltas(state)

    rewards = [
        source_rewards[i] + target_rewards[i] + head_rewards[i] + inclusion_delay_rewards[i]
        for i in range(len(state.validators))
    ]

    penalties = [
        source_penalties[i] + target_penalties[i] + head_penalties[i] + inactivity_penalties[i]
        for i in range(len(state.validators))
    ]

    return rewards, penalties
 */
func GetAttestationDeltas(state *core.State) ([]uint64, []uint64, error) {
	sourceRewards, sourcePenalties, err := GetSourceDeltas(state)
	if err != nil {
		return nil, nil, err
	}
	targetRewards, targetPenalties, err := GetTargetDeltas(state)
	if err != nil {
		return nil, nil, err
	}
	headRewards, headPenalties, err := GetHeadDeltas(state)
	if err != nil {
		return nil, nil, err
	}
	inclusioDelayRewards, _, err := GetInclusionDelayDeltas(state)
	if err != nil {
		return nil, nil, err
	}
	_, inactivityPenalties, err := GetInactivityPenaltyDeltas(state)
	if err != nil {
		return nil, nil, err
	}

	rewards := uint64ZeroArray(uint64(len(state.BlockProducers)))
	for i := range state.BlockProducers {
		rewards[i] = sourceRewards[i] + targetRewards[i] + headRewards[i] + inclusioDelayRewards[i]
	}

	penalties := uint64ZeroArray(uint64(len(state.BlockProducers)))
	for i := range state.BlockProducers {
		penalties[i] = sourcePenalties[i] + targetPenalties[i] + headPenalties[i] + inactivityPenalties[i]
	}

	return rewards, penalties, nil
}

/**
def process_rewards_and_penalties(state: BeaconState) -> None:
    # No rewards are applied at the end of `GENESIS_EPOCH` because rewards are for work done in the previous epoch
    if get_current_epoch(state) == GENESIS_EPOCH:
        return

    rewards, penalties = get_attestation_deltas(state)
    for index in range(len(state.validators)):
        increase_balance(state, ValidatorIndex(index), rewards[index])
        decrease_balance(state, ValidatorIndex(index), penalties[index])
 */
func ProcessRewardsAndPenalties(state *core.State) error {
	if GetCurrentEpoch(state) == params.ChainConfig.GenesisEpoch {
		return nil
	}

	rewards, penalties, err := GetAttestationDeltas(state)
	if err != nil {
		return err
	}

	for _, bp := range state.BlockProducers {
		IncreaseBalance(state, bp.Id, rewards[bp.Id])
		DecreaseBalance(state, bp.Id, penalties[bp.Id])
	}
	return nil
}

func uint64ZeroArray(len uint64) []uint64 {
	ret := make([]uint64, len)
	for i := len ; i < len ; i++ {
		ret[i] = 0
	}
	return ret
}