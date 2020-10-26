package state_transition

import (
	"encoding/hex"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/prysm/shared/mathutil"
	"log"
	"sort"
)

//def process_epoch(state: BeaconState) -> None:
//	process_justification_and_finalization(state)
//	process_rewards_and_penalties(state)
//	process_registry_updates(state)
//	process_slashings(state)
//	process_final_updates(state)
func processEpoch(state *core.State) error {
	if err := processJustificationAndFinalization(state); err != nil {
		return err
	}
	if err := ProcessRewardsAndPenalties(state); err != nil {
		return err
	}
	if err := ProcessRegistryUpdates(state); err != nil {
		return err
	}
	if err := ProcessSlashings(state); err != nil {
		return err
	}
	if err := ProcessFinalUpdates(state); err != nil {
		return err
	}
	return nil
}



/**
def process_justification_and_finalization(state: BeaconState) -> None:
    # Initial FFG checkpoint values have a `0x00` stub for `root`.
    # Skip FFG updates in the first two epochs to avoid corner cases that might result in modifying this stub.
    if get_current_epoch(state) <= GENESIS_EPOCH + 1:
        return

    previous_epoch = get_previous_epoch(state)
    current_epoch = get_current_epoch(state)
    old_previous_justified_checkpoint = state.previous_justified_checkpoint
    old_current_justified_checkpoint = state.current_justified_checkpoint

    # Process justifications
    state.previous_justified_checkpoint = state.current_justified_checkpoint
    state.justification_bits[1:] = state.justification_bits[:JUSTIFICATION_BITS_LENGTH - 1]
    state.justification_bits[0] = 0b0
    matching_target_attestations = get_matching_target_attestations(state, previous_epoch)  # Previous epoch
    if get_attesting_balance(state, matching_target_attestations) * 3 >= get_total_active_balance(state) * 2:
        state.current_justified_checkpoint = Checkpoint(epoch=previous_epoch,
                                                        root=get_block_root(state, previous_epoch))
        state.justification_bits[1] = 0b1
    matching_target_attestations = get_matching_target_attestations(state, current_epoch)  # Current epoch
    if get_attesting_balance(state, matching_target_attestations) * 3 >= get_total_active_balance(state) * 2:
        state.current_justified_checkpoint = Checkpoint(epoch=current_epoch,
                                                        root=get_block_root(state, current_epoch))
        state.justification_bits[0] = 0b1

    # Process finalizations
    bits = state.justification_bits
    # The 2nd/3rd/4th most recent epochs are justified, the 2nd using the 4th as source
    if all(bits[1:4]) and old_previous_justified_checkpoint.epoch + 3 == current_epoch:
        state.finalized_checkpoint = old_previous_justified_checkpoint
    # The 2nd/3rd most recent epochs are justified, the 2nd using the 3rd as source
    if all(bits[1:3]) and old_previous_justified_checkpoint.epoch + 2 == current_epoch:
        state.finalized_checkpoint = old_previous_justified_checkpoint
    # The 1st/2nd/3rd most recent epochs are justified, the 1st using the 3rd as source
    if all(bits[0:3]) and old_current_justified_checkpoint.epoch + 2 == current_epoch:
        state.finalized_checkpoint = old_current_justified_checkpoint
    # The 1st/2nd most recent epochs are justified, the 1st using the 2nd as source
    if all(bits[0:2]) and old_current_justified_checkpoint.epoch + 1 == current_epoch:
        state.finalized_checkpoint = old_current_justified_checkpoint
 */
func processJustificationAndFinalization(state *core.State) error {
	if shared.GetCurrentEpoch(state) <= params.ChainConfig.GenesisEpoch + 1 {
		return nil
	}

	currentEpoch := shared.GetCurrentEpoch(state)
	previousEpoch := shared.GetPreviousEpoch(state)

	oldPrevJustificationPoint := state.PreviousJustifiedCheckpoint
	oldCurrentJustificationPoint := state.CurrentJustifiedCheckpoint

	// process justifications
	state.PreviousJustifiedCheckpoint = state.CurrentJustifiedCheckpoint
	newBits := state.JustificationBits
	newBits.Shift(1)
	state.JustificationBits = newBits


	totalActive := shared.GetTotalActiveBalance(state)

	// Calculate previous epoch attestations justifications.
	matchingTargetAttestations, err := shared.GetMatchingTargetAttestations(state, previousEpoch)
	if err != nil {
		return err
	}
	prevAttestingBalance, err := shared.GetAttestingBalances(state, matchingTargetAttestations)
	if err != nil {
		return err
	}
	log.Printf("Prev epoch %d participation rate: %f\n", previousEpoch, float64(prevAttestingBalance) / float64(totalActive))
	if prevAttestingBalance * 3 >= totalActive * 2 {
		root, err := shared.GetBlockRoot(state, previousEpoch)
		if err != nil {
			return err
		}
		state.CurrentJustifiedCheckpoint = &core.Checkpoint{
			Epoch:                previousEpoch,
			Root:                 root,
		}
		newBits.SetBitAt(1, true)
		state.JustificationBits = newBits
		log.Printf("Justified epoch %d with root %s", previousEpoch, hex.EncodeToString(root))
	}

	// Calculate current epoch attestations justifications.
	matchingTargetAttestations, err = shared.GetMatchingTargetAttestations(state, currentEpoch)
	if err != nil {
		return err
	}
	currentAttestingBalance, err := shared.GetAttestingBalances(state, matchingTargetAttestations)
	if err != nil {
		return err
	}
	log.Printf("Current epoch %d participation rate: %f\n", currentEpoch, float64(currentAttestingBalance) / float64(totalActive))
	if currentAttestingBalance * 3 >= totalActive * 2 {
		root, err := shared.GetBlockRoot(state, currentEpoch)
		if err != nil {
			return err
		}
		state.CurrentJustifiedCheckpoint = &core.Checkpoint{
			Epoch:                currentEpoch,
			Root:                 root,
		}
		newBits.SetBitAt(0, true)
		state.JustificationBits = newBits
		log.Printf("Justified epoch %d with root %s", currentEpoch, hex.EncodeToString(root))
	}

	// process finalization
	justification := state.JustificationBits.Bytes()[0]

	// 2nd/3rd/4th (0b1110) most recent epochs are justified, the 2nd using the 4th as source.
	if justification&0x0E == 0x0E && (oldPrevJustificationPoint.Epoch+3) == currentEpoch {
		state.FinalizedCheckpoint = oldPrevJustificationPoint
		log.Printf("Finalized epoch %d with root %s", state.FinalizedCheckpoint.Epoch, hex.EncodeToString(state.FinalizedCheckpoint.Root))
	}

	// 2nd/3rd (0b0110) most recent epochs are justified, the 2nd using the 3rd as source.
	if justification&0x06 == 0x06 && (oldPrevJustificationPoint.Epoch+2) == currentEpoch {
		state.FinalizedCheckpoint = oldPrevJustificationPoint
		log.Printf("Finalized epoch %d with root %s", state.FinalizedCheckpoint.Epoch, hex.EncodeToString(state.FinalizedCheckpoint.Root))
	}

	// 1st/2nd/3rd (0b0111) most recent epochs are justified, the 1st using the 3rd as source.
	if justification&0x07 == 0x07 && (oldCurrentJustificationPoint.Epoch+2) == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustificationPoint
		log.Printf("Finalized epoch %d with root %s", state.FinalizedCheckpoint.Epoch, hex.EncodeToString(state.FinalizedCheckpoint.Root))
	}

	// The 1st/2nd (0b0011) most recent epochs are justified, the 1st using the 2nd as source
	if justification&0x03 == 0x03 && (oldCurrentJustificationPoint.Epoch+1) == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustificationPoint
		log.Printf("Finalized epoch %d with root %s", state.FinalizedCheckpoint.Epoch, hex.EncodeToString(state.FinalizedCheckpoint.Root))
	}
	return nil
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
	if shared.GetCurrentEpoch(state) == params.ChainConfig.GenesisEpoch {
		return nil
	}

	rewards, penalties, err := shared.GetAttestationDeltas(state)
	if err != nil {
		return err
	}

	for index := range state.Validators {
		shared.IncreaseBalance(state, uint64(index), rewards[uint64(index)])
		shared.DecreaseBalance(state, uint64(index), penalties[uint64(index)])
	}
	return nil
}

/**
def process_registry_updates(state: BeaconState) -> None:
    # Process activation eligibility and ejections
    for index, validator in enumerate(state.validators):
        if is_eligible_for_activation_queue(validator):
            validator.activation_eligibility_epoch = get_current_epoch(state) + 1

        if is_active_validator(validator, get_current_epoch(state)) and validator.effective_balance <= EJECTION_BALANCE:
            initiate_validator_exit(state, ValidatorIndex(index))

    # Queue validators eligible for activation and not yet dequeued for activation
    activation_queue = sorted([
        index for index, validator in enumerate(state.validators)
        if is_eligible_for_activation(state, validator)
        # Order by the sequence of activation_eligibility_epoch setting and then index
    ], key=lambda index: (state.validators[index].activation_eligibility_epoch, index))
    # Dequeued validators for activation up to churn limit
    for index in activation_queue[:get_validator_churn_limit(state)]:
        validator = state.validators[index]
        validator.activation_epoch = compute_activation_exit_epoch(get_current_epoch(state))
*/
func ProcessRegistryUpdates(state *core.State) error {
	for index, bp := range state.Validators {
		if shared.IsEligibleForActivationQueue(bp) {
			bp.ActivationEligibilityEpoch = shared.GetCurrentEpoch(state) + 1
		}

		if shared.IsActiveValidator(bp, shared.GetCurrentEpoch(state)) && bp.EffectiveBalance <= params.ChainConfig.EjectionBalance {
			shared.InitiateValidatorExit(state, uint64(index))
		}
	}

	// Queue validators eligible for activation and not yet dequeued for activation
	activationQueue := []uint64{}
	for index, bp := range state.Validators {
		if shared.IsEligibleForActivation(state, bp) {
			activationQueue = append(activationQueue, uint64(index))
		}
	}
	// Order by the sequence of activation_eligibility_epoch setting and then index
	sort.SliceStable(activationQueue, func(i, j int) bool {
		if state.Validators[i].ActivationEligibilityEpoch == state.Validators[j].ActivationEligibilityEpoch {
			return i < j
		}
		return state.Validators[i].ActivationEligibilityEpoch < state.Validators[j].ActivationEligibilityEpoch
	})

	// Dequeued validators for activation up to churn limit
	for index := range activationQueue[:mathutil.Min(uint64(len(activationQueue)), shared.GetValidatorChurnLimit(state))] {
		bp := state.Validators[index]
		bp.ActivationEpoch = shared.ComputeActivationExitEpoch(shared.GetCurrentEpoch(state))
	}
	return nil
}

/**
def process_slashings(state: BeaconState) -> None:
    epoch = get_current_epoch(state)
    total_balance = get_total_active_balance(state)
    adjusted_total_slashing_balance = min(sum(state.slashings) * PROPORTIONAL_SLASHING_MULTIPLIER, total_balance)
    for index, validator in enumerate(state.validators):
        if validator.slashed and epoch + EPOCHS_PER_SLASHINGS_VECTOR // 2 == validator.withdrawable_epoch:
            increment = EFFECTIVE_BALANCE_INCREMENT  # Factored out from penalty numerator to avoid uint64 overflow
            penalty_numerator = validator.effective_balance // increment * adjusted_total_slashing_balance
            penalty = penalty_numerator // total_balance * increment
            decrease_balance(state, ValidatorIndex(index), penalty)
 */
func ProcessSlashings(state *core.State) error {
	epoch := shared.GetCurrentEpoch(state)
	totalBalance := shared.GetTotalActiveBalance(state)
	adjustedTotalSlashingBalance := mathutil.Min(
			shared.SumSlashings(state) * params.ChainConfig.ProportionalSlashingMultiplier,
			totalBalance,
		)

	for index, bp := range state.Validators {
		if bp.Slashed && epoch + params.ChainConfig.EpochsPerSlashingVector / 2 == bp.WithdrawableEpoch {
			increment := params.ChainConfig.EffectiveBalanceIncrement // Factored out from penalty numerator to avoid uint64 overflow
			penaltyNumerator := bp.EffectiveBalance / increment * adjustedTotalSlashingBalance
			penalty := penaltyNumerator / totalBalance * increment
			shared.DecreaseBalance(state, uint64(index), penalty)
		}
	}
	return nil
}

/**
def process_final_updates(state: BeaconState) -> None:
    current_epoch = get_current_epoch(state)
    next_epoch = Epoch(current_epoch + 1)
    # Reset eth1 data votes
    if next_epoch % EPOCHS_PER_ETH1_VOTING_PERIOD == 0:
        state.eth1_data_votes = []
    # Update effective balances with hysteresis
    for index, validator in enumerate(state.validators):
        balance = state.balances[index]
        HYSTERESIS_INCREMENT = uint64(EFFECTIVE_BALANCE_INCREMENT // HYSTERESIS_QUOTIENT)
        DOWNWARD_THRESHOLD = HYSTERESIS_INCREMENT * HYSTERESIS_DOWNWARD_MULTIPLIER
        UPWARD_THRESHOLD = HYSTERESIS_INCREMENT * HYSTERESIS_UPWARD_MULTIPLIER
        if (
            balance + DOWNWARD_THRESHOLD < validator.effective_balance
            or validator.effective_balance + UPWARD_THRESHOLD < balance
        ):
            validator.effective_balance = min(balance - balance % EFFECTIVE_BALANCE_INCREMENT, MAX_EFFECTIVE_BALANCE)
    # Reset slashings
    state.slashings[next_epoch % EPOCHS_PER_SLASHINGS_VECTOR] = Gwei(0)
    # Set randao mix
    state.randao_mixes[next_epoch % EPOCHS_PER_HISTORICAL_VECTOR] = get_randao_mix(state, current_epoch)
    # Set historical root accumulator
    if next_epoch % (SLOTS_PER_HISTORICAL_ROOT // SLOTS_PER_EPOCH) == 0:
        historical_batch = HistoricalBatch(block_roots=state.block_roots, state_roots=state.state_roots)
        state.historical_roots.append(hash_tree_root(historical_batch))
    # Rotate current/previous epoch attestations
    state.previous_epoch_attestations = state.current_epoch_attestations
    state.current_epoch_attestations = []
 */
func ProcessFinalUpdates(state *core.State) error {
	currentEpoch := shared.GetCurrentEpoch(state)
	nextEpoch := currentEpoch + 1

	// Reset eth1 data votes
	if nextEpoch % params.ChainConfig.EpochsPerETH1VotingPeriod == 0 {
		state.Eth1DataVotes = []*core.ETH1Data{}
	}

	// Update effective balances with hysteresis
	for index, bp := range state.Validators {
		balance := state.Balances[index]

		hysteresisIncrement := params.ChainConfig.EffectiveBalanceIncrement / params.ChainConfig.HysteresisQuotient
		downwardThreshold := hysteresisIncrement * params.ChainConfig.HysteresisDownwardMultiplier
		upwardThreshold := hysteresisIncrement * params.ChainConfig.HysteresisUpwardMultiplier
		if balance + downwardThreshold < bp.EffectiveBalance || bp.EffectiveBalance + upwardThreshold < balance {
			bp.EffectiveBalance = mathutil.Min(balance - balance % params.ChainConfig.EffectiveBalanceIncrement, params.ChainConfig.MaxEffectiveBalance)
		}
	}

	// Reset slashings
	state.Slashings[nextEpoch % params.ChainConfig.EpochsPerSlashingVector] = 0

	// Set randao mix
	state.RandaoMixes[nextEpoch % params.ChainConfig.EpochsPerHistoricalVector] = shared.GetRandaoMix(state, currentEpoch)

	// Set historical root accumulator
	if nextEpoch % (params.ChainConfig.SlotsPerHistoricalRoot / params.ChainConfig.SlotsInEpoch) == 0 {
		hBatch := &core.HistoricalBatch{
			BlockRoots:           state.BlockRoots,
			StateRoots:           state.StateRoots,
		}
		root, err := hBatch.HashTreeRoot()
		if err != nil {
			return err;
		}
		state.HistoricalRoots = append(state.HistoricalRoots, root[:])
	}

	// Rotate current/previous epoch attestations
	state.PreviousEpochAttestations = state.CurrentEpochAttestations
	state.CurrentEpochAttestations = []*core.PendingAttestation{}

	return nil
}