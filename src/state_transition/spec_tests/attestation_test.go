package spec_tests

import (
	"fmt"
	"testing"
)

func TestSpecAttestationMinimal(t *testing.T) {
	testPath := "minimal/phase0/operations/attestation/pyspec_tests"

	SpecTestsFromRootFolder(t, testPath, "preState","attestation")
}

func TestSpecSSZStaticState(t *testing.T) {
	testPath := "mainnet/phase0/ssz_static/BeaconState/ssz_random"
	specTest := SpecTestsFromRootFolder(t, testPath, "serializedState")
	fmt.Sprintf("%s", specTest[0].folderPath)
}