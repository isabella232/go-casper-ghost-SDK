package spec_tests

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSpecRewardsBasicMainnet(t *testing.T) {
	baseRewardsTest(t, "basic")
}

func TestSpecRewardsLeakMainnet(t *testing.T) {
	baseRewardsTest(t, "leak")
}

func TestSpecRewardsRandomMainnet(t *testing.T) {
	baseRewardsTest(t, "random")
}

type ReardDeltas struct {
	Rewards []uint64 `json:"rewards"`
	Penalties []uint64 `json:"penalties"`
}

func baseRewardsTest(t *testing.T, scenario string) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/rewards", scenario)

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

				headDeltas, err := loadDeltas(path.Join(subDir, "head_deltas.yaml"))
				require.NoError(t, err)
				inactivityDeltas, err := loadDeltas(path.Join(subDir, "inactivity_penalty_deltas.yaml"))
				require.NoError(t, err)
				inclusionDeltas, err := loadDeltas(path.Join(subDir, "inclusion_delay_deltas.yaml"))
				require.NoError(t, err)
				sourceDeltas, err := loadDeltas(path.Join(subDir, "source_deltas.yaml"))
				require.NoError(t, err)
				targetDeltas, err := loadDeltas(path.Join(subDir, "target_deltas.yaml"))
				require.NoError(t, err)



				actualHeadDeltasRewards, actualHeadDeltasPenalties, err := shared.GetHeadDeltas(pre)
				require.NoError(t, err)
				require.EqualValues(t, headDeltas.Rewards, actualHeadDeltasRewards)
				require.EqualValues(t, headDeltas.Penalties, actualHeadDeltasPenalties)

				actualInactivityDeltasRewards, actualInactivityDeltasPenalties, err := shared.GetInactivityPenaltyDeltas(pre)
				require.NoError(t, err)
				require.EqualValues(t, inactivityDeltas.Rewards, actualInactivityDeltasRewards)
				require.EqualValues(t, inactivityDeltas.Penalties, actualInactivityDeltasPenalties)

				actualInclusionDeltasRewards, actualInclusionDeltasPenalties, err := shared.GetInclusionDelayDeltas(pre)
				require.NoError(t, err)
				require.EqualValues(t, inclusionDeltas.Rewards, actualInclusionDeltasRewards)
				require.EqualValues(t, inclusionDeltas.Penalties, actualInclusionDeltasPenalties)

				actualSourceDeltasRewards, actualSourceDeltasPenalties, err := shared.GetSourceDeltas(pre)
				require.NoError(t, err)
				require.EqualValues(t, sourceDeltas.Rewards, actualSourceDeltasRewards)
				require.EqualValues(t, sourceDeltas.Penalties, actualSourceDeltasPenalties)

				actualTargetDeltasRewards, actualTargetDeltasPenalties, err := shared.GetTargetDeltas(pre)
				require.NoError(t, err)
				require.EqualValues(t, targetDeltas.Rewards, actualTargetDeltasRewards)
				require.EqualValues(t, targetDeltas.Penalties, actualTargetDeltasPenalties)
			})
		}
	})
}

func loadDeltas(path string) (*ReardDeltas, error) {
	byts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ret := &ReardDeltas{}
	if err := testutil.UnmarshalYaml(byts, ret); err != nil {
		return nil, err
	}
	return ret, nil
}