package shared

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/prysm/shared/hashutil"
	"github.com/wealdtech/go-bytesutil"
)

/**
def get_randao_mix(state: BeaconState, epoch: Epoch) -> Bytes32:
    """
    Return the randao mix at a recent ``epoch``.
    """
    return state.randao_mixes[epoch % EPOCHS_PER_HISTORICAL_VECTOR]
 */
func GetRandaoMix(state *core.State, epoch uint64) []byte {
	return state.RandaoMixes[epoch % params.ChainConfig.EpochsPerHistoricalVector]
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

	seed := append(domainType[:], bytesutil.Bytes8(epoch)...)
	seed = append(seed, randaoMix...)

	return hashutil.Hash(seed)
}