package state_transition

import (
	"bytes"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/prysmaticlabs/go-ssz"
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
	return nil
}

// https://github.com/ethereum/eth2.0-specs/blob/dev/specs/phase0/beacon-chain.md#justification-and-finalization
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

	prev, current, err := calculateAttestingBalances(state)
	if err != nil {
		return err
	}
	if prev.AttestingBalance * 3 >= prev.ActiveBalance * 2 {
		root, err := shared.GetBlockRoot(state, previousEpoch)
		if err != nil {
			return err
		}
		state.CurrentJustifiedCheckpoint = &core.Checkpoint{
			Epoch:                previousEpoch,
			Root:                 root.Bytes,
		}
		newBits.SetBitAt(1, true)
		state.JustificationBits = newBits
	}
	if current.AttestingBalance * 3 >= current.ActiveBalance * 2 {
		root, err := shared.GetBlockRoot(state, currentEpoch)
		if err != nil {
			return err
		}
		state.CurrentJustifiedCheckpoint = &core.Checkpoint{
			Epoch:                currentEpoch,
			Root:                 root.Bytes,
		}
		newBits.SetBitAt(0, true)
		state.JustificationBits = newBits
	}

	// process finalization
	justification := state.JustificationBits.Bytes()[0]

	// 2nd/3rd/4th (0b1110) most recent epochs are justified, the 2nd using the 4th as source.
	if justification&0x0E == 0x0E && (oldPrevJustificationPoint.Epoch+3) == currentEpoch {
		state.FinalizedCheckpoint = oldPrevJustificationPoint
	}

	// 2nd/3rd (0b0110) most recent epochs are justified, the 2nd using the 3rd as source.
	if justification&0x06 == 0x06 && (oldPrevJustificationPoint.Epoch+2) == currentEpoch {
		state.FinalizedCheckpoint = oldPrevJustificationPoint
	}

	// 1st/2nd/3rd (0b0111) most recent epochs are justified, the 1st using the 3rd as source.
	if justification&0x07 == 0x07 && (oldCurrentJustificationPoint.Epoch+2) == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustificationPoint
	}

	// The 1st/2nd (0b0011) most recent epochs are justified, the 1st using the 2nd as source
	if justification&0x03 == 0x03 && (oldCurrentJustificationPoint.Epoch+1) == currentEpoch {
		state.FinalizedCheckpoint = oldCurrentJustificationPoint
	}
	return nil
}

type Balances struct {
	Epoch uint64
	ActiveBalance uint64
	AttestingIndexes []uint64
	AttestingBalance uint64
}

func calculateAttestingBalances(state *core.State) (prev *Balances, current *Balances, err error) {
	// TODO - assert epoch in [currentEpoch, previousEpoch]

	calc := func(attestations []*core.PendingAttestation, epoch uint64) (*Balances,error) {
		// filter matching att. by target root
		matchingAtt := make([]*core.PendingAttestation, 0)
		for _, att := range attestations {
			root, err := shared.GetBlockRoot(state, epoch)
			if err != nil {
				return nil, err
			}
			if  root != nil {
				if bytes.Equal(att.Data.Target.Root, root.Bytes) {
					matchingAtt = append(matchingAtt, att)
				}
			} else {
				return nil, fmt.Errorf("could not find block root for epoch %d", epoch)
			}
		}

		ret := &Balances{
			Epoch:            epoch,
			ActiveBalance:    0,
			AttestingIndexes: []uint64{},
			AttestingBalance: 0,
		}

		// calculate attesting balance and indices
		for _, att := range matchingAtt {
			attestingIndices, err := shared.GetAttestingIndices(state, att.Data, att.AggregationBits)
			if err != nil {
				return nil, err
			}
			for _, idx := range attestingIndices {
				bp := shared.GetBlockProducer(state, idx)
				if bp != nil && !bp.Slashed {
					ret.AttestingIndexes = append(ret.AttestingIndexes, idx)
					ret.AttestingBalance += bp.EffectiveBalance
				}
			}
		}

		// get active balance
		activeBps := shared.GetActiveBlockProducers(state, shared.GetCurrentEpoch(state))
		for _, idx := range activeBps {
			bp := shared.GetBlockProducer(state, idx)
			if bp != nil {
				ret.ActiveBalance += bp.EffectiveBalance
			}
		}

		return ret, nil
	}

	currentEpoch := shared.GetCurrentEpoch(state)
	previousEpoch := shared.GetPreviousEpoch(state)
	prev, err = calc(state.PreviousEpochAttestations, previousEpoch)
	if err != nil {
		return nil, nil, err
	}
	current, err = calc(state.CurrentEpochAttestations, currentEpoch)
	if err != nil {
		return nil, nil, err
	}


	return prev, current, nil
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
func ProcessRegistryUpdates(state *core.State) {
	for _, bp := range state.BlockProducers {
		if shared.IsEligibleForActivationQueue(bp) {
			bp.ActivationEligibilityEpoch = shared.GetCurrentEpoch(state) + 1
		}

		if shared.IsActiveBP(bp, shared.GetCurrentEpoch(state)) && bp.EffectiveBalance <= params.ChainConfig.EjectionBalance {
			shared.InitiateBlockProducerExit(state, bp.Id)
		}
	}

	// Queue validators eligible for activation and not yet dequeued for activation
	activationQueue := []uint64{}
	for index, bp := range state.BlockProducers {
		if shared.IsEligibleForActivation(state, bp) {
			activationQueue = append(activationQueue, uint64(index))
		}
	}
	// Order by the sequence of activation_eligibility_epoch setting and then index
	sort.SliceStable(activationQueue, func(i, j int) bool {
		if state.BlockProducers[i].ActivationEligibilityEpoch == state.BlockProducers[j].ActivationEligibilityEpoch {
			return i < j
		}
		return state.BlockProducers[i].ActivationEligibilityEpoch < state.BlockProducers[j].ActivationEligibilityEpoch
	})

	// Dequeued validators for activation up to churn limit
	for index := range activationQueue[:shared.GetBPChurnLimit(state)] {
		bp := state.BlockProducers[index]
		bp.ActivationEpoch = shared.ComputeActivationExitEpoch(shared.GetCurrentEpoch(state))
	}
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
func ProcessSlashings(state *core.State) {
	epoch := shared.GetCurrentEpoch(state)
	totalBalance := shared.GetTotalActiveStake(state)
	adjustedTotalSlashingBalance := shared.Min(
			shared.SumSlashings(state) * params.ChainConfig.ProportionalSlashingMultiplier,
			totalBalance,
		)

	for _, bp := range state.BlockProducers {
		if bp.Slashed && epoch + params.ChainConfig.EpochsPerSlashingVector / 2 == bp.WithdrawableEpoch {
			increment := params.ChainConfig.EffectiveBalanceIncrement // Factored out from penalty numerator to avoid uint64 overflow
			penaltyNumerator := bp.EffectiveBalance / increment * adjustedTotalSlashingBalance
			penalty := penaltyNumerator / totalBalance * increment
			shared.DecreaseBalance(state, bp.Id, penalty)
		}
	}
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
	for _, bp := range state.BlockProducers {
		hysteresisIncrement := params.ChainConfig.EffectiveBalanceIncrement / params.ChainConfig.HysteresisQuotient
		downwardThreshold := hysteresisIncrement * params.ChainConfig.HysteresisDownwardMultiplier
		upwardThreshold := hysteresisIncrement * params.ChainConfig.HysteresisUpwardMultiplier
		if bp.Balance + downwardThreshold < bp.EffectiveBalance || bp.EffectiveBalance + upwardThreshold < bp.Balance {
			bp.EffectiveBalance = shared.Min(bp.Balance - bp.Balance % params.ChainConfig.EffectiveBalanceIncrement, params.ChainConfig.MaxEffectiveBalance)
		}
	}

	// Reset slashings
	state.Slashings[nextEpoch % params.ChainConfig.EpochsPerSlashingVector] = 0

	// Set randao mix
	state.RandaoMix[nextEpoch % params.ChainConfig.EpochsPerHistoricalVector] = shared.GetRandaoMix(state, currentEpoch)

	// Set historical root accumulator
	if nextEpoch % (params.ChainConfig.SlotsPerHistoricalRoot / params.ChainConfig.SlotsInEpoch) == 0 {
		root, err := ssz.HashTreeRoot(&core.HistoricalBatch{
			BlockRoots:           state.BlockRoots,
			StateRoots:           state.StateRoots,
		})
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