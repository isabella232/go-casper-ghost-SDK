package params

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/ulule/deepcopier"
	"log"
)

func minimalTestingConfig() *core.ChainConfig {
	ret := &core.ChainConfig{}
	if err := deepcopier.Copy(mainnetConfig()).To(ret); err != nil {
		log.Fatal(err)
	}

	ret.MaxCommitteesPerSlot = 4
	ret.TargetCommitteeSize = 4
	ret.SlotsInEpoch = 10
	ret.MinGenesisActiveValidatorCount = 1280

	return ret
}

func UseMinimalTestConfig() {
	ChainConfig = minimalTestingConfig()
}