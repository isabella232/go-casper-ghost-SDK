package spec_tests

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/bloxapp/go-casper-ghost-SDK/src/state_transition"
	ssz "github.com/ferranbt/fastssz"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var testToSSZObject = map[string]string{
	"attestation": "attestation",
	"attester_slashing": "attester_slashing",
	"block_header": "block",
	"deposit": "deposit",
	"proposer_slashing": "proposer_slashing",
	"voluntary_exit": "voluntary_exit",
}

func TestSpecOperationsMainnet(t *testing.T) {
	params.UseMainnetConfig()
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder,"mainnet/phase0/operations")
	objectsToTest, err := ioutil.ReadDir(root)
	require.NoError(t, err)

	for _, testObj := range objectsToTest { // iterate between tests operations/[attestation, attester_slashing..]
		t.Run(testObj.Name(), func(tt *testing.T) {
			objDirsPath := path.Join(root, testObj.Name(),"pyspec_tests/")
			dirs, err := ioutil.ReadDir(objDirsPath)
			require.NoError(t, err)

			for _, dir := range dirs { // iterate scenarios, e.g, operations/attestation/after_epoch_slots
				t.Run(testObj.Name() + "/" + dir.Name(), func(ttt *testing.T) {
					sszObj, ok := testToSSZObject[testObj.Name()]
					if !ok {
						ttt.Skip("could not find ssz object")
						return
					}
					subDir := path.Join(objDirsPath, dir.Name())
					if objFunc, ok := nameToObject[sszObj]; ok && objFunc != nil {
						if dir.Name() == "default_exit_epoch_subsequent_exit" {
							fmt.Printf("")
						}

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

						// unmarshal object
						obj := objFunc()
						byts, err := ioutil.ReadFile(path.Join(subDir, fmt.Sprintf("%s.ssz",sszObj)))
						require.NoError(ttt, err)
						require.NoError(ttt, obj.(ssz.Unmarshaler).UnmarshalSSZ(byts))

						// apply
						ok, err := applyObject(pre, obj)
						require.True(ttt, ok, "apply object not found")

						// verify
						if post != nil {
							targetPostRoot, err := post.HashTreeRoot()
							require.NoError(ttt, err)

							actualPostRoot, err := pre.HashTreeRoot()
							require.NoError(ttt, err)

							require.EqualValues(ttt, targetPostRoot, actualPostRoot)
						} else {
							require.NotNil(ttt, err)
						}
					} else {
						ttt.Skip("no object function")
					}
				})
			}
		})


	}


}


func applyObject(preState *core.State, obj interface{}) (bool, error) {
	if v, ok := obj.(*core.Attestation); ok {
		return true, state_transition.ProcessBlockAttestations(preState, []*core.Attestation{v})
	}
	if v, ok := obj.(*core.AttesterSlashing); ok {
		return true, state_transition.ProcessAttesterSlashings(preState, []*core.AttesterSlashing{v})
	}
	if v, ok := obj.(*core.Block); ok {
		proposer := shared.GetValidator(preState, v.Proposer)
		if proposer == nil {
			return false, fmt.Errorf("block proposer not found")
		}
		return true, state_transition.ProcessBlockHeader(preState, v)
	}
	if v, ok := obj.(*core.Deposit); ok {
		return true, state_transition.ProcessDeposits(preState, []*core.Deposit{v})
	}
	if v, ok := obj.(*core.ProposerSlashing); ok {
		return true, state_transition.ProcessProposerSlashings(preState, []*core.ProposerSlashing{v})
	}
	if v, ok := obj.(*core.SignedVoluntaryExit); ok {
		return true, state_transition.ProcessVoluntaryExit(preState, v)
	}
	return false, nil
}