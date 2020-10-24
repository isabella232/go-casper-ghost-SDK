package shared

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/ulule/deepcopier"
	"log"
)

func CopyState(state *core.State) *core.State {
	if state == nil {
		return nil
	}

	ret := &core.State{}

	ret.Slot = state.Slot
	ret.GenesisTime = state.GenesisTime

	ret.BlockRoots = make([][]byte, len(state.BlockRoots))
	for i, r := range state.BlockRoots {
		ret.BlockRoots[i] = make([]byte, len(state.BlockRoots[i]))
		copy(ret.BlockRoots[i], r)
	}

	ret.StateRoots = make([][]byte, len(state.StateRoots))
	for i, r := range state.StateRoots {
		ret.StateRoots[i] = make([]byte, len(state.StateRoots[i]))
		copy(ret.StateRoots[i], r)
	}

	ret.RandaoMix = make([][]byte, len(state.RandaoMix))
	for i, r := range state.RandaoMix {
		ret.RandaoMix[i] = make([]byte, len(state.RandaoMix[i]))
		copy(ret.RandaoMix[i], r)
	}

	ret.HistoricalRoots = make([][]byte, len(state.HistoricalRoots))
	for i, r := range state.HistoricalRoots {
		ret.HistoricalRoots[i] = make([]byte, len(state.HistoricalRoots[i]))
		copy(ret.HistoricalRoots[i], r)
	}

	ret.Validators = make([]*core.Validator, len(state.Validators))
	for i, bp := range state.Validators {
		ret.Validators[i] = &core.Validator{}
		err := deepcopier.Copy(bp).To(ret.Validators[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.Balances = make([]uint64, len(state.Balances))
	copy(ret.Balances, state.Balances)

	ret.PreviousEpochAttestations = make([]*core.PendingAttestation, len(state.PreviousEpochAttestations))
	for i, pe := range state.PreviousEpochAttestations {
		ret.PreviousEpochAttestations[i] = &core.PendingAttestation{}
		err := deepcopier.Copy(pe).To(ret.PreviousEpochAttestations[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.CurrentEpochAttestations = make([]*core.PendingAttestation, len(state.CurrentEpochAttestations))
	for i, pe := range state.CurrentEpochAttestations {
		ret.CurrentEpochAttestations[i] = &core.PendingAttestation{}
		err := deepcopier.Copy(pe).To(ret.CurrentEpochAttestations[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.JustificationBits = make(bitfield.Bitvector4, len(state.JustificationBits))
	copy(ret.JustificationBits, state.JustificationBits)

	if state.PreviousJustifiedCheckpoint != nil {
		ret.PreviousJustifiedCheckpoint = &core.Checkpoint{}
		err := deepcopier.Copy(state.PreviousJustifiedCheckpoint).To(ret.PreviousJustifiedCheckpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.CurrentJustifiedCheckpoint = &core.Checkpoint{}
	err := deepcopier.Copy(state.CurrentJustifiedCheckpoint).To(ret.CurrentJustifiedCheckpoint)
	if err != nil {
		log.Fatal(err)
	}

	if state.FinalizedCheckpoint != nil {
		ret.FinalizedCheckpoint = &core.Checkpoint{}
		err := deepcopier.Copy(state.FinalizedCheckpoint).To(ret.FinalizedCheckpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	if state.LatestBlockHeader != nil {
		ret.LatestBlockHeader = &core.BlockHeader{}
		err := deepcopier.Copy(state.LatestBlockHeader).To(ret.LatestBlockHeader)
		if err != nil {
			log.Fatal(err)
		}
	}

	if state.Fork != nil {
		ret.Fork = &core.Fork{}
		err := deepcopier.Copy(state.Fork).To(ret.Fork)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.GenesisValidatorsRoot = make([]byte, len(state.GenesisValidatorsRoot))
	copy(ret.GenesisValidatorsRoot, state.GenesisValidatorsRoot)

	if state.Eth1Data != nil {
		ret.Eth1Data = &core.ETH1Data{}
		err := deepcopier.Copy(state.Eth1Data).To(ret.Eth1Data)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.Eth1DataVotes = make([]*core.ETH1Data, len(state.Eth1DataVotes))
	for i, v := range state.Eth1DataVotes {
		if v != nil { // TODO - can eth1Data be nil?
			ret.Eth1DataVotes[i] = &core.ETH1Data{}
			err := deepcopier.Copy(v).To(ret.Eth1DataVotes[i])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	ret.Eth1DepositIndex = state.Eth1DepositIndex

	ret.Slashings = make([]uint64, len(state.Slashings))
	copy(ret.Slashings, state.Slashings)

	return ret
}

// will return nil if not found or inactive
func GetValidator (state *core.State, id uint64) *core.Validator {
	if id < uint64(len(state.Validators)) {
		return state.Validators[id]
	}
	return nil
}

/**
def is_valid_genesis_state(state: BeaconState) -> bool:
    if state.genesis_time < MIN_GENESIS_TIME:
        return False
    if len(get_active_validator_indices(state, GENESIS_EPOCH)) < MIN_GENESIS_ACTIVE_VALIDATOR_COUNT:
        return False
    return True
 */
func IsValidGenesisState(state *core.State) bool {
	if state.GenesisTime < params.ChainConfig.MinGenesisTime {
		return false
	}
	if uint64(len(GetActiveValidators(state, params.ChainConfig.GenesisEpoch))) < params.ChainConfig.MinGenesisActiveValidatorCount {
		return false
	}
	return true
}

func SumSlashings(state *core.State) uint64 {
	totalSlashing := uint64(0)
	for _, slashing := range state.Slashings {
		totalSlashing += slashing
	}
	return totalSlashing
}