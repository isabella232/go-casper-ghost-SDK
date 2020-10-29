package spec_tests

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/bloxapp/go-casper-ghost-SDK/src/state_transition"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSpecFinalityMainnet(t *testing.T) {
	baseFinalityTest(t, "finality")
}

type FinalityMeta struct {
	Blocks_count int `json:"blocks_count"`
}

func baseFinalityTest(t *testing.T, scenario string) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/finality", scenario)

	t.Run(scenario, func(tt *testing.T) {
		phaseDirsPath := path.Join(root,"pyspec_tests/")
		dirs, err := ioutil.ReadDir(phaseDirsPath)
		require.NoError(tt, err)

		for _, dir := range dirs { // iterate scenarios, e.g, rewards/basic/empty....
			t.Run(dir.Name(), func(ttt *testing.T) {
				subDir := path.Join(phaseDirsPath, dir.Name())

				// unmarshal pre state
				preByts, err := ioutil.ReadFile(path.Join(subDir, "pre.ssz"))
				require.NoError(ttt, err)
				pre := &core.State{}
				require.NoError(ttt, pre.UnmarshalSSZ(preByts))
				// unmarshal post state if exists
				postByts, err := ioutil.ReadFile(path.Join(subDir, "post.ssz"))
				post := &core.State{}
				require.NoError(ttt, post.UnmarshalSSZ(postByts))
				// load meta object
				metaByts, err := ioutil.ReadFile(path.Join(subDir, "meta.yaml"))
				require.NoError(t, err)
				meta := &FinalityMeta{}
				require.NoError(t, testutil.UnmarshalYaml(metaByts, meta))

				// load blocks
				blocks := make([]*core.SignedBlock, meta.Blocks_count)
				for i := 0 ; i < len(blocks) ; i++ {
					byts, err := ioutil.ReadFile(path.Join(subDir, fmt.Sprintf("blocks_%d.ssz", i)))
					require.NoError(t, err)
					blocks[i] = &core.SignedBlock{}
					require.NoError(t, blocks[i].UnmarshalSSZ(byts))
				}

				// execute blocks
				st := state_transition.NewStateTransition()
				for _, blk := range blocks {
					newState, err := st.ExecuteStateTransition(pre, blk, true)
					require.NoError(t, err)
					pre = newState
				}
			})
		}
	})
}