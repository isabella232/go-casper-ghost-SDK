package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"sort"
)

func ProcessNewPoolRequests(state *core.State, requests []*core.CreateNewPoolRequest) error {
	for _, req := range requests {
		leader := shared.GetBlockProducer(state, req.StartEpoch)
		if leader == nil {
			return fmt.Errorf("could not find new pool req leader")
		}

		// verify leader is correct
		if req.LeaderBlockProducer != leader.Id {
			return fmt.Errorf("new pool req leader incorrect")
		}
		if shared.GetPool(state, req.Id) != nil {
			return fmt.Errorf("new pool id == req id, this is already exists")
		}
		// TODO - check that network has enough capitalization
		// TODO - check leader is not part of DKG Committee

		// TODO - what if a BP doesn't have enough CDT for penalties?

		// get DKG participants
		committee, err := shared.GetVaultCommittee(state, req.Id, req.StartEpoch)
		if err != nil {
			return err
		}
		sort.Slice(committee, func(i int, j int) bool {
			return committee[i] < committee[j]
		})

		switch req.GetStatus() {
		case 0:
			// TODO if i'm the DKG leader act uppon it
		case 1: // successful
			// get committee
			committee, err := shared.GetVaultCommittee(state, req.Id, req.StartEpoch)
			sort.Slice(committee, func(i int, j int) bool {
				return committee[i] < committee[j]
			})

			state.Pools = append(state.Pools, &core.Pool{
				Id:              req.Id,
				PubKey:          req.GetCreatePubKey(),
				SortedCommittee: committee,
			})
			if err != nil {
				return err
			}

			// reward/ penalty
			for i := 0 ; i < len(committee) ; i ++ {
				bp := shared.GetBlockProducer(state, committee[i])
				if bp == nil {
					return fmt.Errorf("could not find BP %d", committee[i])
				}
				partic := req.GetParticipation()
				if partic[:].BitAt(uint64(i)) {
					shared.IncreaseBalance(state, bp.Id, params.ChainConfig.DKGReward)
				} else {
					shared.DecreaseBalance(state, bp.Id, params.ChainConfig.DKGReward)
				}
			}

			// special reward for leader
			shared.IncreaseBalance(state, leader.Id, 3*params.ChainConfig.DKGReward)
		case 2: // un-successful
			// TODO - better define how the un-successful status is assigned.
			for i := 0 ; i < len(committee) ; i ++ {
				bp := shared.GetBlockProducer(state, committee[i])
				if bp == nil {
					return fmt.Errorf("could not find BP %d", committee[i])
				}
				shared.DecreaseBalance(state, bp.Id, params.ChainConfig.DKGReward)
			}
		}
	}
	return nil
}
