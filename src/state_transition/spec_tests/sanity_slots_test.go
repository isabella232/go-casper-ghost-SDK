package spec_tests

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/bloxapp/go-casper-ghost-SDK/src/state_transition"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestSpecSanitySlotsMainnet(t *testing.T) {
	baseSanitySlotsTest(t, "slots")
}

func baseSanitySlotsTest(t *testing.T, scenario string) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/sanity", scenario)

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
				metaByts, err := ioutil.ReadFile(path.Join(subDir, "slots.yaml"))
				require.NoError(t, err)
				fileStr := string(metaByts)
				slotsCount, err := strconv.Atoi(fileStr[:len(fileStr)-5])
				require.NoError(t, err)

				// execute process slots
				st := state_transition.NewStateTransition()
				require.NoError(t, st.ProcessSlots(pre, pre.Slot + uint64(slotsCount)))

				// compare roots
				expectedRoot, err := post.HashTreeRoot()
				require.NoError(t, err)
				actualRoot, err := pre.HashTreeRoot()
				require.NoError(t, err)
				require.EqualValues(t, expectedRoot[:], actualRoot[:])
			})
		}
	})
}
