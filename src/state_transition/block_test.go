package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStateCopying(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)

	preRoot, err := ssz.HashTreeRoot(state)
	require.NoError(t, err)

	newState := shared.CopyState(state)
	require.NoError(t, err)

	// test new state and old state ssz
	newStateRoot, err := ssz.HashTreeRoot(newState)
	require.NoError(t, err)
	require.EqualValues(t, preRoot, newStateRoot)

	// test manipulating prams on new state copying
	bp := shared.GetBlockProducer(newState, 0)
	bp.CDTBalance = 100000
	require.NotEqualValues(t, shared.GetBlockProducer(state, 0).CDTBalance, 100000)

	shared.GetPool(newState, 0).Active = false
	require.NotEqualValues(t, shared.GetPool(state, 0).Active, shared.GetPool(newState, 0).Active)

	// test old state root not changed
	postRoot, err := ssz.HashTreeRoot(state)
	require.NoError(t, err)
	require.EqualValues(t, preRoot, postRoot)
}

func TestBlockApplyConsistency(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 3)
	block := &core.PoolBlock{
		Slot:                 2,
		Proposer:             13,
		ParentRoot:           toByte("71dcfc4567f947c7c396f293a615b3e46554a83595703399107d1b87d6b6ae3c"),
		StateRoot:            nil,
		Body:                 &core.PoolBlockBody{
			RandaoReveal:          toByte("b99d58464b006350d5348891225744c3e0c683598e27a2bc8088db6d068580a5aa53c63a55894803f0b0e189870d85d204ba1caf80ef102a012d04784e3ec1726adb234a01400b4e471715d13b43f6b336c8638be7f8ab4fb050d118161e9a36"),
			NewPoolReq:      nil,
			Attestations:         generateAttestations(
				state,
				128,
				1,
				&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
				&core.Checkpoint{Epoch: 0, Root: []byte{}},
				0,
				true,
				0, /* attestation */
			),
		},
	}

	// sign
	blockDomain, err := shared.GetDomain(0, params.ChainConfig.DomainBeaconProposer, state.GenesisValidatorsRoot)
	require.NoError(t, err)
	sig, err := shared.SignBlock(
		block,
		[]byte(fmt.Sprintf("%d", 13)),
		blockDomain)
	require.NoError(t, err)
	signed := &core.SignedPoolBlock{
		Block:                block,
		Signature:            sig.Serialize(),
	}


	preRoot,err := ssz.HashTreeRoot(state)
	require.NoError(t, err)

	var postRoot []byte
	for i := 0 ; i < 10 ; i++ {
		newState := shared.CopyState(state)
		st := NewStateTransition()
		err := st.ProcessBlock(newState, signed)
		require.NoError(t, err)

		if i != 0 {
			require.EqualValues(t, postRoot, shared.GetStateRoot(newState, 1))
		}

		postRoot = shared.GetStateRoot(newState, 1)
	}

	post,err := ssz.HashTreeRoot(state)
	require.NoError(t, err)
	require.EqualValues(t, preRoot, post)
}