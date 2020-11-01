package spec_tests

import (
	"encoding/hex"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/stretchr/testify/require"
	"github.com/wealdtech/go-bytesutil"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

type ShufflingMeta struct {
	Seed string
	Count int
	Mapping []uint64
}

func TestShuffling(t *testing.T) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/shuffling/core/shuffle")
	dirs, err := ioutil.ReadDir(root)
	require.NoError(t, err)

	for _, dir := range dirs { // iterate scenarios, e.g, rewards/basic/empty....
		t.Run(dir.Name(), func(ttt *testing.T) {
			subDir := path.Join(root, dir.Name())

			// load meta object
			metaByts, err := ioutil.ReadFile(path.Join(subDir, "mapping.yaml"))
			require.NoError(t, err)
			meta := &ShufflingMeta{}
			require.NoError(t, testutil.UnmarshalYaml(metaByts, meta))

			seed, err := hex.DecodeString(meta.Seed[2:])
			require.NoError(t, err)
			mapping, err := helpers.UnshuffleList(listFromLength(meta.Count), bytesutil.ToBytes32(seed))
			require.NoError(t, err)
			require.EqualValues(t, meta.Mapping, mapping)
		})
	}
}

func listFromLength(l int) []uint64 {
	ret := make([]uint64, l)
	for i := 0 ; i < l ; i++ {
		ret[i] = uint64(i)
	}
	return ret
}