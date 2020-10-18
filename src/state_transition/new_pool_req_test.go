package state_transition

import (
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestCreatedNewPoolReq(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	req := []*core.CreateNewPoolRequest{
		{
			Id:                  13,
			Status:              1, // completed
			StartEpoch:          1,
			EndEpoch:            2,
			LeaderBlockProducer: 1,
			CreatePubKey:        toByte("a3b9110ec26cbb02e6182fab4dcb578d17411f26e41f16aad99cfce51e9bc76ce5e7de00a831bbcadd1d7bc0235c945d"), // priv: 3ef5411174c7d9672652bf4ffc342af3720cc23e52c377b95927871645435f41
			Participation:       bitfield.Bitlist{43,12},
		},
	}

	st := NewStateTransition()

	err := st.ProcessNewPoolRequests(state, req)
	require.NoError(t, err)

	// check created
	require.Equal(t, 13, len(state.Pools))

	// check rewards
	participation := bitfield.Bitlist{43,12}
	committee, err := shared.GetVaultCommittee(state, 13, 1)
	sort.Slice(committee, func(i int, j int) bool {
		return committee[i] < committee[j]
	})
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize ; i++ {
		bp := shared.GetBlockProducer(state, committee[i])
		if participation.BitAt(i) {
			require.EqualValues(t, 2000, bp.CDTBalance)
		} else {
			require.EqualValues(t, 0, bp.CDTBalance)
		}
	}

	// leader reward
	bp := shared.GetBlockProducer(state, 1)
	require.EqualValues(t, 4000, bp.CDTBalance)

	// pool data
	pool := shared.GetPool(state, 13)
	require.NotNil(t, pool)
	require.EqualValues(t, toByte("a3b9110ec26cbb02e6182fab4dcb578d17411f26e41f16aad99cfce51e9bc76ce5e7de00a831bbcadd1d7bc0235c945d"), pool.PubKey)
	require.EqualValues(t, committee, pool.SortedCommittee)
}

func TestNotCreatedNewPoolReq(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	req := []*core.CreateNewPoolRequest{
		{
			Id:                  13,
			Status:              2, // cancelled
			StartEpoch:          1,
			EndEpoch:            2,
			LeaderBlockProducer: 1,
			CreatePubKey:        toByte("a3b9110ec26cbb02e6182fab4dcb578d17411f26e41f16aad99cfce51e9bc76ce5e7de00a831bbcadd1d7bc0235c945d"), // priv: 3ef5411174c7d9672652bf4ffc342af3720cc23e52c377b95927871645435f41
			Participation:       bitfield.Bitlist{43, 12},
		},
	}

	st := NewStateTransition()

	err := st.ProcessNewPoolRequests(state, req)
	require.NoError(t, err)

	// check not created
	require.Equal(t, 12, len(state.Pools))

	// check penalties
	committee, err := shared.GetVaultCommittee(state, 13, 1)
	sort.Slice(committee, func(i int, j int) bool {
		return committee[i] < committee[j]
	})
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize ; i++ {
		bp := shared.GetBlockProducer(state, committee[i])
		require.EqualValues(t, 0, bp.CDTBalance)
	}

	// leader reward
	bp := shared.GetBlockProducer(state, 1)
	require.EqualValues(t, 1000, bp.CDTBalance)
}

func TestCreatedNewPoolReqWithExistingId(t *testing.T) {
	t.Skipf("create pool not supported yet")
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	req := []*core.CreateNewPoolRequest{
		{
			Id:                  12,
			Status:              2, // completed
			StartEpoch:          1,
			EndEpoch:            2,
			LeaderBlockProducer: 1,
			CreatePubKey:        toByte("a3b9110ec26cbb02e6182fab4dcb578d17411f26e41f16aad99cfce51e9bc76ce5e7de00a831bbcadd1d7bc0235c945d"), // priv: 3ef5411174c7d9672652bf4ffc342af3720cc23e52c377b95927871645435f41
			Participation:       bitfield.Bitlist{43,12,89},
		},
	}

	st := NewStateTransition()

	err := st.ProcessNewPoolRequests(state, req)
	require.Error(t, err, "new pool id == req id, this is already exists")
}