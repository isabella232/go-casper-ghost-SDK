package spec_tests

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/bloxapp/go-casper-ghost-SDK/src/state_transition"
	ssz "github.com/ferranbt/fastssz"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestSpecOperationsMainnet(t *testing.T) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/operations/")

	t.Run("", func(ttt *testing.T) {
		subDirs := "attestation"
		subDir := path.Join(root, subDirs, "pyspec_tests/success")

		if objFunc, ok := nameToObject[subDirs]; ok && objFunc != nil {
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
			}

			// unmarshal object
			obj := objFunc()
			byts, err := ioutil.ReadFile(path.Join(subDir, fmt.Sprintf("%s.ssz",subDirs)))
			require.NoError(ttt, err)
			require.NoError(ttt, obj.(ssz.Unmarshaler).UnmarshalSSZ(byts))

			// apply
			applyObject(t, pre, obj)
		} else {
			ttt.Skip("no object function")
		}
	})
}


func applyObject(t *testing.T, preState *core.State, obj interface{}) {
	if v, ok := obj.(*core.Attestation); ok {
		require.NoError(t, state_transition.ProcessBlockAttestations(preState, []*core.Attestation{v}))
	}
}