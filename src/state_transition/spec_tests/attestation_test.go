package spec_tests

import (
	"encoding/hex"
	"fmt"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"path"
	"testing"
)

func TestSpecAttestationMinimal(t *testing.T) {
	testPath := "minimal/phase0/operations/attestation/pyspec_tests"

	SpecTestsFromRootFolder(t, testPath, "preState","attestation")
}

func TestSpecSSZStaticState(t *testing.T) {
	testPath := "mainnet/phase0/ssz_static/BeaconState/ssz_random"
	specTests := SpecTestsFromRootFolder(t, testPath, "serializedState")

	type SSZRoots struct {
		Root        string `json:"root"`
		SigningRoot string `json:"signing_root"`
	}

	for _, test := range specTests {
		// read expected root
		byts, err := ioutil.ReadFile(path.Join(test.folderPath, "roots.yaml"))
		require.NoError(t, err)
		rootsYaml := &SSZRoots{}
		require.NoError(t, testutil.UnmarshalYaml(byts, rootsYaml))

		expectedRoot, err := hex.DecodeString(rootsYaml.Root[2:])
		require.NoError(t, err)

		// hash root state
		stateRoot, err := test.preState.HashTreeRoot()
		fmt.Printf("root: %s\n", hex.EncodeToString(stateRoot[:]))

		require.NoError(t, err)
		require.EqualValues(t, expectedRoot, stateRoot[:])
	}
}