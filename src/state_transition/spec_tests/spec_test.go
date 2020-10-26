package spec_tests

import (
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var rootSpecTestsFolder = ".temp/tests"

type SpecTest struct {
	folderPath string
	t *testing.T

	preState *core.State
	postState *core.State

	// variables that could be nil
	attestation *core.Attestation
}

func SpecTestsFromRootFolder(t *testing.T, folderPath string, objects ...string) []*SpecTest {
	base, err := os.Getwd()
	require.NoError(t, err)

	root := path.Join(base, rootSpecTestsFolder, folderPath)
	ret := make([]*SpecTest, 0)


	files, err := ioutil.ReadDir(root)
	require.NoError(t, err)
	for _, file := range files {
		if file.IsDir() {
			ret = append(ret, SpecTestFromFolder(t, path.Join(root, file.Name()), objects...))
		}
	}
	return ret
}

func SpecTestFromFolder(t *testing.T, folderPath string, objects ...string) *SpecTest {
	ret := &SpecTest{
		folderPath:folderPath,
		t:t,
	}
	for _, obj := range objects {
		switch obj {
		case "pre":
		case "serialized":
			ret.readPreState(obj)
		case "attestation":
			ret.readAttestation()
		}
	}
	return ret
}

func(test *SpecTest) readPreState(fileName string) {
	preByts, err := ioutil.ReadFile(path.Join(test.folderPath, fmt.Sprintf("%s.ssz", fileName)))
	require.NoError(test.t, err)
	test.preState = &core.State{}
	require.NoError(test.t, test.preState.UnmarshalSSZ(preByts))
}

func(test *SpecTest) readAttestation() {
	attByts, err := ioutil.ReadFile(path.Join(test.folderPath, "attestation.ssz"))
	require.NoError(test.t, err)
	test.attestation = &core.Attestation{}
	require.NoError(test.t, test.attestation.UnmarshalSSZ(attByts))
}