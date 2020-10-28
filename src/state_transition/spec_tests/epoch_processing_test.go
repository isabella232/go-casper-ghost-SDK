package spec_tests

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/bloxapp/go-casper-ghost-SDK/src/state_transition"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSpecEpochProcessingMainnet(t *testing.T) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/epoch_processing")
	phaseToTest, err := ioutil.ReadDir(root)
	require.NoError(t, err)

	for _, phase := range phaseToTest { // iterate between tests epoch_processing/[final_updates, registry_updates..]
		t.Run(phase.Name(), func(tt *testing.T) {
			phaseDirsPath := path.Join(root, phase.Name(),"pyspec_tests/")
			dirs, err := ioutil.ReadDir(phaseDirsPath)
			require.NoError(tt, err)

			for _, dir := range dirs { // iterate scenarios, e.g, epoch_processing/final_updates/effective_balance_hyteresis...
				t.Run(phase.Name() + "/" + dir.Name(), func(ttt *testing.T) {
					subDir := path.Join(phaseDirsPath, dir.Name())

					// unmarshal pre state
					preByts, err := ioutil.ReadFile(path.Join(subDir, "pre.ssz"))
					require.NoError(ttt, err)
					pre := &core.State{}
					require.NoError(ttt, pre.UnmarshalSSZ(preByts))
					// unmarshal post state if exists
					postByts, err := ioutil.ReadFile(path.Join(subDir, "post.ssz"))
					post := &core.State{}
					if err == nil {
						require.NoError(ttt, post.UnmarshalSSZ(postByts))
					} else {
						post = nil
					}

					if dir.Name() == "attestations_some_slashed" {
						fmt.Printf("")
					}

					ok, err := applyPhase(pre, phase.Name())
					require.True(ttt, ok, "apply phase not found")


					// verify pre and post roots
					targetPostRoot, err := post.HashTreeRoot()
					require.NoError(ttt, err)

					actualPostRoot, err := pre.HashTreeRoot()
					require.NoError(ttt, err)

					require.EqualValues(ttt, targetPostRoot, actualPostRoot)
				})
			}
		})
	}
}

func applyPhase(state *core.State, phase string)(bool, error) {
	if phase == "final_updates" {
		return true, state_transition.ProcessFinalUpdates(state)
	}
	if phase == "justification_and_finalization" {
		return true, state_transition.ProcessJustificationAndFinalization(state)
	}
	if phase == "registry_updates" {
		return true, state_transition.ProcessRegistryUpdates(state)
	}
	if phase == "rewards_and_penalties" {
		return true, state_transition.ProcessRewardsAndPenalties(state)
	}
	if phase == "slashings" {
		return true, state_transition.ProcessSlashings(state)
	}
	return false, nil
}