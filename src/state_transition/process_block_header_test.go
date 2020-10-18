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

func TestProcessBlockHeader(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))
	tests := []struct{
		name              string
		block             *core.PoolBlock
		signerBP          uint64
		expectedError     error
	}{
		{
			name: "valid sig",
			block: &core.PoolBlock{
				Proposer:        13,
				Slot:            2,
				Body: &core.PoolBlockBody{
					RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
				},
				ParentRoot: toByte("71dcfc4567f947c7c396f293a615b3e46554a83595703399107d1b87d6b6ae3c"),
			},
			signerBP:          13,
			expectedError:     nil,
		},
		{
			name: "invalid sig",
			block: &core.PoolBlock{
				Proposer:        13,
				Slot:            2,
				Body: &core.PoolBlockBody{
					RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
				},
				ParentRoot: toByte("71dcfc4567f947c7c396f293a615b3e46554a83595703399107d1b87d6b6ae3c"),
			},
			signerBP: 12,
			expectedError: fmt.Errorf("block sig not verified"),
		},
		{
			name: "wrong proposer",
			block: &core.PoolBlock{
				Proposer:        2,
				Slot:            2,
				Body: &core.PoolBlockBody{
					RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
				},
				ParentRoot: toByte("332863d85bdafc9e5ccaeec92d12f00452bd9e3d71b80af4a0cab9df35c5e56f"),
			},
			signerBP: 2,
			expectedError: fmt.Errorf("block expectedProposer is worng, expected 13 but received 2"),
		},
		{
			name: "invalid proposer",
			block: &core.PoolBlock{
				Proposer:        4550000000,
				Slot:            2,
				Body: &core.PoolBlockBody{
					RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
				},
				ParentRoot: toByte("332863d85bdafc9e5ccaeec92d12f00452bd9e3d71b80af4a0cab9df35c5e56f"),
			},
			signerBP: 2,
			expectedError: fmt.Errorf("block expectedProposer is worng, expected 13 but received 4550000000"),
		},
		//{ // TODO ?
		//	name: "invalid block root",
		//	block: &core.PoolBlock{
		//		Proposer:        13,
		//		Slot:            2,
		//		Body: &core.PoolBlockBody{
		//			RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
		//		},
		//		ParentRoot: toByte("332863d85bdafc9e5ccaeec92d12f00452bd9e3d71b80af4a0cab9df35c5e56f"),
		//	},
		//	signerBP: 13,
		//	expectedError: fmt.Errorf("signed block root does not match block root"),
		//	useCorretBodyRoot: false,
		//},
		//{ // TODO - when randao processing done
		//	name: "RANDAO too small",
		//	block: &core.PoolBlock{
		//		Proposer:        13,
		//		Slot:            2,
		//		Body: &core.PoolBlockBody{
		//			RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6"),
		//		},
		//		ParentRoot: toByte("332863d85bdafc9e5ccaeec92d12f00452bd9e3d71b80af4a0cab9df35c5e56f"),
		//	},
		//	signerBP: 13,
		//	expectedError: fmt.Errorf("RANDAO should be 32 byte"),
		//},
		//{ // TODO - when randao processing done
		//	name: "RANDAO too big",
		//	block: &core.PoolBlock{
		//		Proposer:        13,
		//		Slot:            2,
		//		Body: &core.PoolBlockBody{
		//			RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6ddd"),
		//		},
		//		ParentRoot: toByte("332863d85bdafc9e5ccaeec92d12f00452bd9e3d71b80af4a0cab9df35c5e56f"),
		//	},
		//	signerBP: 13,
		//	expectedError: fmt.Errorf("RANDAO should be 32 byte"),
		//},
		{
			name: "invalid parent block root",
			block: &core.PoolBlock{
				Proposer:        13,
				Slot:            2,
				Body: &core.PoolBlockBody{
					RandaoReveal:         toByte("97c4116516e77c522344aa3c3c223db0c14bad05aa005be63aadd19341e0cc6d"),
				},
				ParentRoot: toByte("75141b2e032f1b045ab9c7998dfd7238044e40eed0b2c526c33340643e871e42"),
			},
			signerBP: 13,
			expectedError: fmt.Errorf("parent block root doesn't match, expected 71dcfc4567f947c7c396f293a615b3e46554a83595703399107d1b87d6b6ae3c"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := generateTestState(t, 3)

			// sign
			sk := []byte(fmt.Sprintf("%d", test.signerBP))
			blockDomain, err := shared.GetDomain(0, params.ChainConfig.DomainBeaconProposer, state.GenesisValidatorsRoot)
			require.NoError(t, err)
			sig, err := shared.SignBlock(test.block, sk, blockDomain) // TODO - dynamic domain
			require.NoError(t, err)


			// header
			signed := &core.SignedPoolBlock{
				Block:                test.block,
				Signature:            sig.Serialize(),
			}

			if test.expectedError != nil {
				require.EqualError(t, processBlockHeader(state, signed), test.expectedError.Error())
			} else {
				require.NoError(t, processBlockHeader(state, signed))
			}
		})
	}
}