package state_transition

import (
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAttestationProcessing(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	thresholdSig := uint64(params.ChainConfig.MinAttestationCommitteeSize * 2 / 3 + 1)

	state := generateTestState(t, 34)
	tests := []struct{
		name string
		block *core.PoolBlock
		expectedError error
	}{
		{
			name: "valid block attestation epoch 0",
			block: &core.PoolBlock{
				Slot:                 2,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						1,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 0, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: nil,
		},
		{
			name: "valid block attestation epoch 1 with source epoch 0",
			block: &core.PoolBlock{
				Slot:                 33,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						32,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 1, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: nil,
		},
		{
			name: "threshold sig not achieved",
			block: &core.PoolBlock{
				Slot:                 32,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig - 1,
						32,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 1, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: fmt.Errorf("attestation did not pass threshold"),
		},
		{
			name: "target epoch invalid",
			block: &core.PoolBlock{
				Slot:                 32,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						32,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 5, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: fmt.Errorf("taregt not in current/ previous epoch"),
		},
		{
			name: "target slot not in the correct epoch",
			block: &core.PoolBlock{
				Slot:                 32,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						32,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 0, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: fmt.Errorf("target slot not in the correct epoch"),
		},
		{
			name: "min att. inclusion delay did not pass",
			block: &core.PoolBlock{
				Slot:                 32,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						33,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 1, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: fmt.Errorf("min att. inclusion delay did not pass"),
		},
		{
			name: "slot to submit att. has passed",
			block: &core.PoolBlock{
				Slot:                 32,
				Proposer:             0,
				ParentRoot:           nil,
				StateRoot:            nil,
				Body:                 &core.PoolBlockBody{
					Attestations:         generateAttestations(
						state,
						thresholdSig,
						0,
						&core.Checkpoint{Epoch: 0, Root: params.ChainConfig.ZeroHash},
						&core.Checkpoint{Epoch: 0, Root: []byte{}},
						0,
						true,
						0, /* attestation */
					),
				},
			},
			expectedError: fmt.Errorf("slot to submit att. has passed"),
		},
		//{ // TODO - complete
		//	name: "committee index out of range",
		//	blockBody: &core.BlockBody{
		//		Slot:                 32,
		//		Attestations:         generateAttestations(
		//			state,
		//			86,
		//			32,
		//			0,
		//			1,
		//			1000000,
		//			true,
		//			0, /* attestation */
		//		),
		//	},
		//	expectedError: fmt.Errorf("slot to submit att. has passed"),
		//},
	}

	for _, test := range tests {
		t.Run(test.name, func (t *testing.T) {
			require.Len(t, test.block.Body.Attestations, 1)

			stateCopy := shared.CopyState(state)
			st := NewStateTransition()

			if test.expectedError != nil {
				require.EqualError(t, st.ProcessBlockAttestations(stateCopy, test.block.Body.Attestations), test.expectedError.Error())
			} else {
				require.NoError(t, st.ProcessBlockAttestations(stateCopy, test.block.Body.Attestations))
			}
		})
	}
}