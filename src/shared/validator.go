package shared

import (
	"bytes"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

/**
	def is_active_validator(validator: Validator, epoch: Epoch) -> bool:
		"""
		Check if ``validator`` is active.
		"""
		return validator.activation_epoch <= epoch < validator.exit_epoch
 */
func IsActiveBP(bp *core.BlockProducer, epoch uint64) bool {
	return bp.ActivationEpoch <= epoch && epoch < bp.ExitEpoch
}

/**
	def is_eligible_for_activation_queue(validator: Validator) -> bool:
		"""
		Check if ``validator`` is eligible to be placed into the activation queue.
		"""
		return (
			validator.activation_eligibility_epoch == FAR_FUTURE_EPOCH
			and validator.effective_balance == MAX_EFFECTIVE_BALANCE
		)
*/
func IsEligibleForActivationQueue(bp *core.BlockProducer) bool {
	return bp.ActivationEligibilityEpoch == params.ChainConfig.FarFutureEpoch && bp.EffectiveBalance == params.ChainConfig.MaxEffectiveBalance
}

/**
	def is_eligible_for_activation(state: BeaconState, validator: Validator) -> bool:
		"""
		Check if ``validator`` is eligible for activation.
		"""
		return (
			# Placement in queue is finalized
			validator.activation_eligibility_epoch <= state.finalized_checkpoint.epoch
			# Has not yet been activated
			and validator.activation_epoch == FAR_FUTURE_EPOCH
		)
 */
func IsEligibleForActivation(state *core.State, bp *core.BlockProducer) bool {
	return bp.ActivationEligibilityEpoch <= state.FinalizedCheckpoint.Epoch && // Placement in queue is finalized
					bp.ActivationEpoch == params.ChainConfig.FarFutureEpoch // Has not yet been activated
}

/**
	def is_slashable_validator(validator: Validator, epoch: Epoch) -> bool:
		"""
		Check if ``validator`` is slashable.
		"""
		return (not validator.slashed) and (validator.activation_epoch <= epoch < validator.withdrawable_epoch)
 */
func IsSlashableBp(bp *core.BlockProducer, epoch uint64) bool {
	return !bp.Slashed && (bp.ActivationEpoch <= epoch && epoch < bp.WithdrawableEpoch)
}

/**
def compute_proposer_index(state: BeaconState, indices: Sequence[ValidatorIndex], seed: Bytes32) -> ValidatorIndex:
    """
    Return from ``indices`` a random index sampled by effective balance.
    """
    assert len(indices) > 0
    MAX_RANDOM_BYTE = 2**8 - 1
    i = uint64(0)
    total = uint64(len(indices))
    while True:
        candidate_index = indices[compute_shuffled_index(i % total, total, seed)]
        random_byte = hash(seed + uint_to_bytes(uint64(i // 32)))[i % 32]
        effective_balance = state.validators[candidate_index].effective_balance
        if effective_balance * MAX_RANDOM_BYTE >= MAX_EFFECTIVE_BALANCE * random_byte:
            return candidate_index
        i += 1
 */
func ComputeProposerIndex(state *core.State, indices []uint64, seed []byte) (uint64, error) {
	if len(indices) == 0 {
		return 0, fmt.Errorf("couldn't compute proposer, indices list empty")
	}
	maxRandomByte := uint64(2^8-1)
	i := uint64(0)
	total := uint64(len(indices))
	for {
		idx, err := computeShuffledIndex(i % total, total, SliceToByte32(seed), true,10) // TODO - shuffle round via config
		if err != nil {
			return 0, err
		}

		candidateIndex := indices[idx]
		b := append(seed[:], Bytes8(i / 32)...)
		randomByte := Hash(b)[i%32]

		bp := GetBlockProducer(state, candidateIndex)
		if bp == nil {
			return 0, fmt.Errorf("could not find shuffled BP index %d", candidateIndex)
		}
		effectiveBalance := bp.EffectiveBalance

		if effectiveBalance * maxRandomByte >= params.ChainConfig.MaxEffectiveBalance * uint64(randomByte) {
			return candidateIndex, nil
		}
	}
}

/**
def compute_activation_exit_epoch(epoch: Epoch) -> Epoch:
    """
    Return the epoch during which validator activations and exits initiated in ``epoch`` take effect.
    """
    return Epoch(epoch + 1 + MAX_SEED_LOOKAHEAD)
 */
func ComputeActivationExitEpoch(epoch uint64) uint64 {
	return epoch + 1 + params.ChainConfig.MaxSeedLookahead
}

/**
def get_active_validator_indices(state: BeaconState, epoch: Epoch) -> Sequence[ValidatorIndex]:
    """
    Return the sequence of active validator indices at ``epoch``.
    """
    return [ValidatorIndex(i) for i, v in enumerate(state.validators) if is_active_validator(v, epoch)]
 */
func GetActiveBlockProducers(state *core.State, epoch uint64) []uint64 {
	var activeBps []uint64
	for _, bp := range state.BlockProducers {
		if IsActiveBP(bp, epoch) {
			activeBps = append(activeBps, bp.GetId())
		}
	}
	return activeBps
}

/**
def get_validator_churn_limit(state: BeaconState) -> uint64:
    """
    Return the validator churn limit for the current epoch.
    """
    active_validator_indices = get_active_validator_indices(state, get_current_epoch(state))
    return max(MIN_PER_EPOCH_CHURN_LIMIT, uint64(len(active_validator_indices)) // CHURN_LIMIT_QUOTIENT)
 */
func GetBPChurnLimit(state *core.State) uint64 {
	activeBPs := GetActiveBlockProducers(state, GetCurrentEpoch(state))
	churLimit := uint64(len(activeBPs)) / params.ChainConfig.ChurnLimitQuotient
	if churLimit < params.ChainConfig.MinPerEpochChurnLimit {
		churLimit = params.ChainConfig.MinPerEpochChurnLimit
	}
	return churLimit
}

/**
def get_beacon_proposer_index(state: BeaconState) -> ValidatorIndex:
    """
    Return the beacon proposer index at the current slot.
    """
    epoch = get_current_epoch(state)
    seed = hash(get_seed(state, epoch, DOMAIN_BEACON_PROPOSER) + uint_to_bytes(state.slot))
    indices = get_active_validator_indices(state, epoch)
    return compute_proposer_index(state, indices, seed)
 */
func GetBlockProposerIndex(state *core.State) (uint64, error) {
	epoch := GetCurrentEpoch(state)
	seed := GetSeed(state, epoch, params.ChainConfig.DomainBeaconProposer)
	SeedWithSlot := append(seed[:], Bytes8(state.CurrentSlot)...)
	hash := Hash(SeedWithSlot)

	bps := GetActiveBlockProducers(state, epoch)
	return ComputeProposerIndex(state, bps, hash[:])
}

/**
def increase_balance(state: BeaconState, index: ValidatorIndex, delta: Gwei) -> None:
    """
    Increase the validator balance at index ``index`` by ``delta``.
    """
    state.balances[index] += delta
 */
func IncreaseBalance(state *core.State, index uint64, delta uint64) {
	if bp := GetBlockProducer(state, index); bp != nil {
		bp.Balance += delta
	}
}

/**
def decrease_balance(state: BeaconState, index: ValidatorIndex, delta: Gwei) -> None:
    """
    Decrease the validator balance at index ``index`` by ``delta``, with underflow protection.
    """
    state.balances[index] = 0 if delta > state.balances[index] else state.balances[index] - delta
*/
func DecreaseBalance(state *core.State, index uint64, delta uint64) {
	if bp := GetBlockProducer(state, index); bp != nil {
		if delta > bp.Balance {
			bp.Balance = 0
		} else {
			bp.Balance -= delta
		}
	}
}

/**
def initiate_validator_exit(state: BeaconState, index: ValidatorIndex) -> None:
    """
    Initiate the exit of the validator with index ``index``.
    """
    # Return if validator already initiated exit
    validator = state.validators[index]
    if validator.exit_epoch != FAR_FUTURE_EPOCH:
        return

    # Compute exit queue epoch
    exit_epochs = [v.exit_epoch for v in state.validators if v.exit_epoch != FAR_FUTURE_EPOCH]
    exit_queue_epoch = max(exit_epochs + [compute_activation_exit_epoch(get_current_epoch(state))])
    exit_queue_churn = len([v for v in state.validators if v.exit_epoch == exit_queue_epoch])
    if exit_queue_churn >= get_validator_churn_limit(state):
        exit_queue_epoch += Epoch(1)

    # Set validator exit epoch and withdrawable epoch
    validator.exit_epoch = exit_queue_epoch
    validator.withdrawable_epoch = Epoch(validator.exit_epoch + MIN_VALIDATOR_WITHDRAWABILITY_DELAY)
 */
func InitiateBlockProducerExit(state *core.State, index uint64) {
	bp := GetBlockProducer(state, index)
	if bp == nil {
		return
	}
	if bp.ExitEpoch != params.ChainConfig.FarFutureEpoch {
		return
	}

	// Compute exit queue epoch
	exitEpochs := []uint64{}
	for _, bp := range state.BlockProducers {
		if bp.ExitEpoch != params.ChainConfig.FarFutureEpoch {
			exitEpochs = append(exitEpochs, bp.ExitEpoch)
		}
	}
	exitEpochs = append(exitEpochs, ComputeActivationExitEpoch(GetCurrentEpoch(state)))

	// Obtain the exit queue epoch as the maximum number in the exit epochs array.
	exitQueueEpoch := uint64(0)
	for _, i := range exitEpochs {
		if exitQueueEpoch < i {
			exitQueueEpoch = i
		}
	}

	// We use the exit queue churn to determine if we have passed a churn limit.
	exitQueueChurn := 0
	for _, bp := range state.BlockProducers {
		if bp.ExitEpoch == exitQueueEpoch {
			exitQueueChurn ++
		}
	}

	// Set validator exit epoch and withdrawable epoch
	bp.ExitEpoch = exitQueueEpoch
	bp.WithdrawableEpoch = bp.ExitEpoch + params.ChainConfig.MinValidatorWithdrawabilityDelay
}

/**
def slash_validator(state: BeaconState,
                    slashed_index: ValidatorIndex,
                    whistleblower_index: ValidatorIndex=None) -> None:
    """
    Slash the validator with index ``slashed_index``.
    """
    epoch = get_current_epoch(state)
    initiate_validator_exit(state, slashed_index)
    validator = state.validators[slashed_index]
    validator.slashed = True
    validator.withdrawable_epoch = max(validator.withdrawable_epoch, Epoch(epoch + EPOCHS_PER_SLASHINGS_VECTOR))
    state.slashings[epoch % EPOCHS_PER_SLASHINGS_VECTOR] += validator.effective_balance
    decrease_balance(state, slashed_index, validator.effective_balance // MIN_SLASHING_PENALTY_QUOTIENT)

    # Apply proposer and whistleblower rewards
    proposer_index = get_beacon_proposer_index(state)
    if whistleblower_index is None:
        whistleblower_index = proposer_index
    whistleblower_reward = Gwei(validator.effective_balance // WHISTLEBLOWER_REWARD_QUOTIENT)
    proposer_reward = Gwei(whistleblower_reward // PROPOSER_REWARD_QUOTIENT)
    increase_balance(state, proposer_index, proposer_reward)
    increase_balance(state, whistleblower_index, Gwei(whistleblower_reward - proposer_reward))
 */
func SlashBlockProducer(state *core.State, slashedIndex uint64) error {
	epoch := GetCurrentEpoch(state)
	InitiateBlockProducerExit(state, slashedIndex)
	bp := GetBlockProducer(state, slashedIndex)
	if bp == nil {
		return fmt.Errorf("slash BP: block producer not found")
	}
	bp.Slashed = true
	bp.WithdrawableEpoch = Max(bp.WithdrawableEpoch, epoch + params.ChainConfig.EpochsPerSlashingVector)
	state.Slashings[epoch % params.ChainConfig.EpochsPerSlashingVector] += bp.EffectiveBalance
	DecreaseBalance(state, slashedIndex, bp.EffectiveBalance / params.ChainConfig.MinSlashingPenaltyQuotient)

	// Apply proposer and whistleblower rewards
	proposer, err := GetBlockProposerIndex(state)
	if err != nil {
		return err
	}
	whistleblowerIndex := proposer
	whistleblowerReward := bp.EffectiveBalance / params.ChainConfig.WhitstleblowerRewardQuotient
	proposerReward := whistleblowerReward / params.ChainConfig.ProposerRewardQuotient
	IncreaseBalance(state, proposer, proposerReward)
	IncreaseBalance(state, whistleblowerIndex, whistleblowerReward)
	return nil
}

func BPByPubkey(state *core.State, pk []byte) *core.BlockProducer {
	// TODO - BPByPubkey optimize with some kind of map
	for _, bp := range state.BlockProducers {
		if bytes.Equal(pk, bp.PubKey) {
			return bp
		}
	}
	return nil
}