package state_transition

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSanity(t *testing.T) {
	_, err := NewStateTestContext(
			params.ChainConfig,
			nil,
			0,
		)
	require.NoError(t, err)

}
