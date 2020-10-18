package shared

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

/**
def get_randao_mix(state: BeaconState, epoch: Epoch) -> Bytes32:
    """
    Return the randao mix at a recent ``epoch``.
    """
    return state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR]
 */
func GetRandaoMix(state *core.State, epoch uint64) []byte {
	return state.Randao[len(state.Randao) - 1].Bytes // TODO - GetRandaoMix make it as spec
}

func GetLatestRandaoMix(state *core.State) []byte {
	return state.Randao[len(state.Randao) - 1].Bytes
}

// Returns the seed after randao was applied on the last slot of the epoch
// will return error if not found
func GetEpochSeed(state *core.State, epoch uint64) ([]byte, error) { // TODO - should be replaced with GetRandaoMix
	if epoch == 0 {
		return params.ChainConfig.GenesisSeed, nil
	}

	targetSlot := epoch * params.ChainConfig.SlotsInEpoch - 1 + params.ChainConfig.SlotsInEpoch
	for _, d := range state.Randao {
		if d.Slot == targetSlot {
			return d.Bytes, nil
		}
	}
	return []byte{}, fmt.Errorf("seed for slot %d not found", targetSlot)
}

/**
def get_seed(state: BeaconState, epoch: Epoch, domain_type: DomainType) -> Bytes32:
    """
    Return the seed at ``epoch``.
    """
    mix = get_randao_mix(state, Epoch(epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD - 1))  # Avoid underflow
    return hash(domain_type + uint_to_bytes(epoch) + mix)
 */
func GetSeed(state *core.State, epoch uint64, domainType []byte) [32]byte {
	randaoMix := GetRandaoMix(state, epoch + params.ChainConfig.EpochsPerHistoricalVector - params.ChainConfig.MinSeedLookahead - 1)

	seed := append(domainType[:], Bytes8(epoch)...)
	seed = append(seed, randaoMix...)

	return Hash(domainType)
}