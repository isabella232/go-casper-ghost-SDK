package state_transition

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSanity(t *testing.T) {
	ctx, err := NewStateTestContext(
			params.ChainConfig,
			nil,
			0,
		)
	require.NoError(t, err)

	ctx.PopulateGenesisValidator(params.ChainConfig.MinGenesisActiveValidatorCount)
	ctx.ProgressSlotsAndEpochs(128)

	fmt.Printf("")
}
