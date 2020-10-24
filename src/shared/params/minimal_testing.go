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

	ret.TargetCommitteeSize = 4
	ret.SlotsInEpoch = 4
	ret.MinGenesisActiveValidatorCount = ret.TargetCommitteeSize * ret.SlotsInEpoch

	return ret
}

func UseMinimalTestConfig() {
	ChainConfig = minimalTestingConfig()
}