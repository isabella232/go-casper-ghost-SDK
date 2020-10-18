package state_transition

import (
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/core"
	"github.com/bloxapp/eth2-staking-pools-research/go-spec/src/shared"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEpochJustification(t *testing.T) {
	require.NoError(t, bls.Init(bls.BLS12_381))
	require.NoError(t, bls.SetETHmode(bls.EthModeDraft07))

	state := generateTestState(t, 95)

	err := populateJustificationAndFinalization(state, 2, 95, 1, &core.Checkpoint{
		Epoch:                2,
		Root:                 shared.GetBlockRoot(state, 2).Bytes,
	})

	require.NoError(t, err)
	require.NoError(t, processJustificationAndFinalization(state))
	require.EqualValues(t,
		toByte("6084f1d26031b7f4736769b1bc5751e00e8d08aa4c4d8dfe46648c44778d105b"),
		state.CurrentJustifiedCheckpoint.Root,
		fmt.Sprintf("actual current justified checkpoint root: %s\n", hex.EncodeToString(state.CurrentJustifiedCheckpoint.Root)))
}
