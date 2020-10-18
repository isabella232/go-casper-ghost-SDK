package params

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
)

// Bytes4 returns integer x to bytes in little-endian format, x.to_bytes(4, 'little').
// TODO - copied here for cyclic dependency issue
func Bytes4(x uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, x)
	return bytes[:4]
}

func testConfig() *core.PoolsChainConfig {
	genesisSeed,_ := hex.DecodeString("sdddseedseedseedseedseedseedseed")

	return &core.PoolsChainConfig{
		// Time
		SlotsInEpoch:                32,
		MinAttestationInclusionDelay: 1,
		MaxSeedLookahead: 2^2, // 4 epochs
		MinSeedLookahead: 1, // 1 epoch
		SlotsPerHistoricalRoot: 2 ^ 13, // ~27H
		MinValidatorWithdrawabilityDelay: 2^8, // 256 epochs, ~27 hours
		MinEpochsToInactivityPenalty: 2^2, // 4 epochs 25.6 min
		EpochsPerETH1VotingPeriod: 2^5, // 32 ~3.4 hours
		ShardCommitteePeriod: 2^8, // 256, ~27H

		// initial values

		// Misc
		MinAttestationCommitteeSize: 16,
		MaxAttestationCommitteeSize: 16,
		MaxCommitteesPerSlot:        2^6, // 64
		ChurnLimitQuotient:          2^16, // 65,536
		VaultSize:                   4,
		MinPerEpochChurnLimit:       4,
		MinGenesisTime:              1578009600, // Jan 3, 2020
		MinGenesisActiveBPCount:     2^8,        // 256
		ProportionalSlashingMultiplier: 3,
		HysteresisQuotient: 4,
		HysteresisDownwardMultiplier: 1,
		HysteresisUpwardMultiplier: 5,

		// constants
		FarFutureEpoch: 		2^64-1,
		ZeroHash: 				make([]byte, 32),
		GenesisSeed: 	       	genesisSeed,
		GenesisEpoch: 		   	0,
		BaseRewardsPerEpoch:   	4,
		DepositContractTreeDepth: 2^5, // 32

		// state list lengths
		EpochsPerHistoricalVector: 2 ^ 16, // ~36 days
		EpochsPerSlashingVector: 2^13, // 8,192, ~36 days

		// rewards and penalties
		BaseRewardFactor: 2^6, // 64
		BaseEth2DutyReward:    100,
		DKGReward:             1000,
		MinSlashingPenaltyQuotient: 2^5, // 32
		WhitstleblowerRewardQuotient: 2^9, // 512
		ProposerRewardQuotient: 2^3, // 8
		InactivityPenaltyQuotient: 2^24, // 16,777,216

		// domain
		DomainBeaconProposer: Bytes4(0),
		DomainBeaconAttester: Bytes4(1),
		DomainRandao: Bytes4(2),
		DomainDeposit: Bytes4(3),
		DomainVoluntaryExit: Bytes4(4),
		DomainSelectionProof:Bytes4(5),
		DomainAggregateAndProof:Bytes4(6),
		GenesisForkVersion: []byte{},

		// Gwei values
		MaxEffectiveBalance: 2^5 * 10^9, // 32 ETH
		EffectiveBalanceIncrement: 2^0 * 2^9, // 1 ETH
		EjectionBalance: 2^0 * 2^9, // 16 ETH

		// Max operations per block
		MaxProposerSlashings: 16,
		MaxAttesterSlashings: 2,
		MaxAttestations: 128,
		MaxDeposits: 16,
		MaxVoluntaryExits: 16,
	}
}

func UseMinimalTestConfig() {
	ChainConfig = testConfig()
}