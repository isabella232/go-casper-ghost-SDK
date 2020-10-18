package params

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
)

// Bytes4 returns integer x to bytes in little-endian format, x.to_bytes(4, 'little').
// TODO - copied here for cyclic dependency issue
func Bytes4(x uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, x)
	return bytes[:4]
}

func mainnetConfig() *core.ChainConfig {
	genesisSeed,_ := hex.DecodeString("sdddseedseedseedseedseedseedseed")

	return &core.ChainConfig{
		// initial values
		GenesisForkVersion: Bytes4(0),

		// Time
		SlotsInEpoch:                32,
		MinAttestationInclusionDelay: 1,
		MaxSeedLookahead: 2^2, // 4 epochs
		MinSeedLookahead: 1, // 1 epoch
		SlotsPerHistoricalRoot: 1 << 13, // ~27H
		MinValidatorWithdrawabilityDelay: 1 << 8, // 256 epochs, ~27 hours
		MinEpochsToInactivityPenalty: 4, // 4 epochs 25.6 min
		EpochsPerETH1VotingPeriod: 32, // 32 ~3.4 hours
		ShardCommitteePeriod: 1 << 8, // 256, ~27H

		// initial values

		// Misc
		TargetCommitteeSize:		    128,
		MaxValidatorsPerCommittee:    	2048,
		MaxCommitteesPerSlot:           1 << 6, // 64
		ChurnLimitQuotient:             1 << 16, // 65,536
		VaultSize:                      4,
		MinPerEpochChurnLimit:          4,
		MinGenesisTime:                 1578009600, // Jan 3, 2020
		MinGenesisActiveValidatorCount: 1 << 14,        // 16,384
		ProportionalSlashingMultiplier: 3,
		HysteresisQuotient:             4,
		HysteresisDownwardMultiplier:   1,
		HysteresisUpwardMultiplier:     5,
		ShuffleRoundCount: 				90,

		// constants
		FarFutureEpoch: 		1 << 64-1,
		ZeroHash: 				make([]byte, 32),
		GenesisSeed: 	       	genesisSeed,
		GenesisEpoch: 		   	0,
		BaseRewardsPerEpoch:   	4,
		DepositContractTreeDepth: 1 << 5, // 32

		// state list lengths
		EpochsPerHistoricalVector: 1 << 16, // ~36 days
		EpochsPerSlashingVector: 1 << 13, // 8,192, ~36 days

		// rewards and penalties
		BaseRewardFactor: 1 << 6, // 64
		BaseEth2DutyReward:    100,
		DKGReward:             1000,
		MinSlashingPenaltyQuotient: 1 << 5, // 32
		WhitstleblowerRewardQuotient: 1 << 9, // 512
		ProposerRewardQuotient: 1 << 3, // 8
		InactivityPenaltyQuotient: 1 << 24, // 16,777,216

		// domain
		DomainBeaconProposer: Bytes4(0),
		DomainBeaconAttester: Bytes4(1),
		DomainRandao: Bytes4(2),
		DomainDeposit: Bytes4(3),
		DomainVoluntaryExit: Bytes4(4),
		DomainSelectionProof:Bytes4(5),
		DomainAggregateAndProof:Bytes4(6),

		// Gwei values
		MaxEffectiveBalance: 32 * 1e9, // 32 ETH
		EffectiveBalanceIncrement: 1 * 1e9, // 1 ETH
		EjectionBalance: 16 * 1e9, // 16 ETH

		// Max operations per block
		MaxProposerSlashings: 16,
		MaxAttesterSlashings: 2,
		MaxAttestations: 128,
		MaxDeposits: 16,
		MaxVoluntaryExits: 16,
	}
}

func UseMainnetConfig() {
	ChainConfig = mainnetConfig()
}