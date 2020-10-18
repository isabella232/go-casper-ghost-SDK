package state_transition

import (
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandaoRevealMix(t *testing.T) {
	state := generateTestState(t, 3)
	proposer, err := shared.BlockProposer(state, 2)
	require.NoError(t, err)
	// get data
	data, domain, err := RANDAOSigningData(state)
	require.NoError(t, err)
	// sign
	sig, err := shared.SignRandao(data, domain, []byte(fmt.Sprintf("%d", proposer)))
	require.NoError(t, err)
	require.NoError(t, processRANDAONoVerify(state,
		&core.PoolBlock{
			Slot:                 2,
			Proposer:             0,
			ParentRoot:           nil,
			StateRoot:            nil,
			Body:                 &core.PoolBlockBody{
				RandaoReveal: sig.Serialize(),
				Attestations: nil,
			},
		}),
	)
	require.EqualValues(t,
		toByte("ff6d5266a95330ff6b1cf9bfbf9525244ece2f31a9440604f20a63ec8f92c575"),
		state.Randao[3].Bytes,
		fmt.Sprintf("randao actual: %s\n", hex.EncodeToString(state.Randao[3].Bytes)))
}
