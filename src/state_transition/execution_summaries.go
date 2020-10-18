package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
)

func validateExecutionSummaries(state *core.State, summaries []*core.ExecutionSummary) error {
	// TODO - validate summaries epoch, should be in some range?
	return nil
}

func ProcessExecutionSummaries(state *core.State, summaries []*core.ExecutionSummary) error {
	if err := validateExecutionSummaries(state, summaries); err != nil {
		return err
	}

	// TODO - what if a BP doesn't have enough CDT for penalties?
	for _, summary := range summaries {
		pool := shared.GetPool(state, summary.GetPoolId())
		if pool == nil {
			return fmt.Errorf("could not find pool %d", summary.GetPoolId())
		}
		if !pool.Active {
			return fmt.Errorf("pool %d is not active", summary.GetPoolId())
		}

		executors := pool.GetSortedCommittee()

		for _, duty := range summary.GetDuties() {
			switch duty.GetType() {
			case 0: // attestation
				for i:=0 ; i < int(params.ChainConfig.VaultSize) ; i++ {
					bp := shared.GetBlockProducer(state, executors[i])
					if bp == nil {
						return fmt.Errorf("BP %d not found", executors[i])
					}

					if !duty.Finalized {
						shared.DecreaseBalance(state, bp.Id, 2*params.ChainConfig.BaseEth2DutyReward)
					} else {
						participation := duty.GetParticipation()
						if participation.BitAt(uint64(i)) {
							shared.IncreaseBalance(state, bp.Id, params.ChainConfig.BaseEth2DutyReward)
						} else {
							shared.DecreaseBalance(state, bp.Id, params.ChainConfig.BaseEth2DutyReward)
						}
					}
				}
			case 1: // proposal
				for i:=0 ; i < int(params.ChainConfig.VaultSize) ; i++ {
					bp := shared.GetBlockProducer(state, executors[i])
					if bp == nil {
						return fmt.Errorf("BP %d not found", executors[i])
					}

					if !duty.Finalized {
						shared.DecreaseBalance(state, bp.Id, 4*params.ChainConfig.BaseEth2DutyReward)
					} else {
						participation := duty.GetParticipation()
						if participation[:].BitAt(uint64(i)) {
							shared.IncreaseBalance(state, bp.Id, 2*params.ChainConfig.BaseEth2DutyReward)
						} else {
							shared.DecreaseBalance(state, bp.Id, 2*params.ChainConfig.BaseEth2DutyReward)
						}
					}
				}
			}
		}
	}
	return nil
}
