package state_transition

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"testing"
)

func TestSanity(t *testing.T) {
	ctx := NewStateTestContext(
			params.ChainConfig,
			nil,
			0,
		)
	ctx.PopulateGenesisValidator(params.ChainConfig.MinGenesisActiveValidatorCount)
	ctx.ProgressSlotsAndEpochs(128, 1, 2)

	fmt.Printf("")
}
