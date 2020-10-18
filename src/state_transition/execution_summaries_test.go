package state_transition

import (
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFinalizedAttestation(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	atts := generateAttestations(
		state,
		128,
		33,
		&core.Checkpoint{Epoch: 0, Root: []byte{}},
		&core.Checkpoint{Epoch: 1, Root: []byte{}},
		0,
		true,
		0, /* attestation */
	)
	st := NewStateTransition()

	err := st.ProcessExecutionSummaries(state, atts[0].Data.ExecutionSummaries)
	require.NoError(t, err)

	// check rewards
	participation := bitfield.Bitlist{1,3,88}
	committee := shared.GetPool(state, 3).SortedCommittee
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize; i++ { // pool id = 3
		bp := shared.GetBlockProducer(state, committee[i])
		if participation.BitAt(i) {
			require.EqualValues(t, 1100, bp.CDTBalance)
		} else {
			require.EqualValues(t, 900, bp.CDTBalance)
		}
	}
}

func TestNotFinalizedAttestation(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	atts := generateAttestations(
		state,
		128,
		33,
		&core.Checkpoint{Epoch: 0, Root: []byte{}},
		&core.Checkpoint{Epoch: 1, Root: []byte{}},
		0,
		false,
		0, /* attestation */
	)

	st := NewStateTransition()

	err := st.ProcessExecutionSummaries(state, atts[0].Data.ExecutionSummaries)
	require.NoError(t, err)

	// check rewards
	committee := shared.GetPool(state, 3).SortedCommittee
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize; i++ { // pool id = 3
		bp := shared.GetBlockProducer(state, committee[i])
		require.EqualValues(t, 800, bp.CDTBalance)
	}
}

func TestFinalizedProposal(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	att := generateAttestations(
		state,
		128,
		33,
		&core.Checkpoint{Epoch: 0, Root: []byte{}},
		&core.Checkpoint{Epoch: 1, Root: []byte{}},
		0,
		true,
		1, /* proposal */
	)
	st := NewStateTransition()

	err := st.ProcessExecutionSummaries(state, att[0].Data.ExecutionSummaries)
	require.NoError(t, err)

	// check rewards
	participation := bitfield.Bitlist{1,3,88}
	committee := shared.GetPool(state, 3).SortedCommittee
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize; i++ { // pool id = 3
		bp := shared.GetBlockProducer(state, committee[i])
		if participation.BitAt(i) {
			require.EqualValues(t, 1200, bp.CDTBalance)
		} else {
			require.EqualValues(t, 800, bp.CDTBalance)
		}
	}
}

func TestNotFinalizedProposal(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	att := generateAttestations(
		state,
		128,
		33,
		&core.Checkpoint{Epoch: 0, Root: []byte{}},
		&core.Checkpoint{Epoch: 1, Root: []byte{}},
		0,
		false,
		1, /* proposal */
	)

	st := NewStateTransition()

	err := st.ProcessExecutionSummaries(state, att[0].Data.ExecutionSummaries)
	require.NoError(t, err)

	// check rewards
	committee := shared.GetPool(state, 3).SortedCommittee
	require.NoError(t, err)

	// test penalties/ rewards
	for i := uint64(0) ; i < params.ChainConfig.VaultSize; i++ { // pool id = 3
		bp := shared.GetBlockProducer(state, committee[i])
		require.EqualValues(t, 600, bp.CDTBalance)
	}
}