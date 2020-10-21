package shared

import (
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/ulule/deepcopier"
	"log"
)

// There are 21 fields in the beacon state.
const fieldCount = 21

var (
	leavesCache = make(map[string][][32]byte, fieldCount)
	layersCache = make(map[string][][][32]byte, fieldCount)
)

func CopyState(state *core.State) *core.State {
	if state == nil {
		return nil
	}

	ret := &core.State{}

	ret.Slot = state.Slot
	ret.GenesisTime = state.GenesisTime

	ret.BlockRoots = make([][]byte, len(state.BlockRoots))
	for i, r := range state.BlockRoots {
		ret.BlockRoots[i] = make([]byte, len(state.BlockRoots[i]))
		copy(ret.BlockRoots[i], r)
	}

	ret.StateRoots = make([][]byte, len(state.StateRoots))
	for i, r := range state.StateRoots {
		ret.StateRoots[i] = make([]byte, len(state.StateRoots[i]))
		copy(ret.StateRoots[i], r)
	}

	ret.RandaoMix = make([][]byte, len(state.RandaoMix))
	for i, r := range state.RandaoMix {
		ret.RandaoMix[i] = make([]byte, len(state.RandaoMix[i]))
		copy(ret.RandaoMix[i], r)
	}

	ret.HistoricalRoots = make([][]byte, len(state.HistoricalRoots))
	for i, r := range state.HistoricalRoots {
		ret.HistoricalRoots[i] = make([]byte, len(state.HistoricalRoots[i]))
		copy(ret.HistoricalRoots[i], r)
	}

	ret.Validators = make([]*core.Validator, len(state.Validators))
	for i, bp := range state.Validators {
		ret.Validators[i] = bp // TODO - state copy (validators) copying issue
		//ret.Validators[i] = &core.Validator{}
		//err := deepcopier.Copy(bp).To(ret.Validators[i])
		//if err != nil {
		//	log.Fatal(err)
		//}
	}

	ret.PreviousEpochAttestations = make([]*core.PendingAttestation, len(state.PreviousEpochAttestations))
	for i, pe := range state.PreviousEpochAttestations {
		ret.PreviousEpochAttestations[i] = &core.PendingAttestation{}
		err := deepcopier.Copy(pe).To(ret.PreviousEpochAttestations[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.CurrentEpochAttestations = make([]*core.PendingAttestation, len(state.CurrentEpochAttestations))
	for i, pe := range state.CurrentEpochAttestations {
		ret.CurrentEpochAttestations[i] = &core.PendingAttestation{}
		err := deepcopier.Copy(pe).To(ret.CurrentEpochAttestations[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.JustificationBits = make(bitfield.Bitvector4, len(state.JustificationBits))
	copy(ret.JustificationBits, state.JustificationBits)

	if state.PreviousJustifiedCheckpoint != nil {
		ret.PreviousJustifiedCheckpoint = &core.Checkpoint{}
		err := deepcopier.Copy(state.PreviousJustifiedCheckpoint).To(ret.PreviousJustifiedCheckpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.CurrentJustifiedCheckpoint = &core.Checkpoint{}
	err := deepcopier.Copy(state.CurrentJustifiedCheckpoint).To(ret.CurrentJustifiedCheckpoint)
	if err != nil {
		log.Fatal(err)
	}

	if state.FinalizedCheckpoint != nil {
		ret.FinalizedCheckpoint = &core.Checkpoint{}
		err := deepcopier.Copy(state.FinalizedCheckpoint).To(ret.FinalizedCheckpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	if state.LatestBlockHeader != nil {
		ret.LatestBlockHeader = &core.BlockHeader{}
		err := deepcopier.Copy(state.LatestBlockHeader).To(ret.LatestBlockHeader)
		if err != nil {
			log.Fatal(err)
		}
	}

	if state.Fork != nil {
		ret.Fork = &core.Fork{}
		err := deepcopier.Copy(state.Fork).To(ret.Fork)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.GenesisValidatorsRoot = make([]byte, len(state.GenesisValidatorsRoot))
	copy(ret.GenesisValidatorsRoot, state.GenesisValidatorsRoot)

	if state.Eth1Data != nil {
		ret.Eth1Data = &core.ETH1Data{}
		err := deepcopier.Copy(state.Eth1Data).To(ret.Eth1Data)
		if err != nil {
			log.Fatal(err)
		}
	}

	ret.Eth1DataVotes = make([]*core.ETH1Data, len(state.Eth1DataVotes))
	for i, v := range state.Eth1DataVotes {
		if v != nil { // TODO - can eth1Data be nil?
			ret.Eth1DataVotes[i] = &core.ETH1Data{}
			err := deepcopier.Copy(v).To(ret.Eth1DataVotes[i])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	ret.Eth1DepositIndex = state.Eth1DepositIndex

	ret.Slashings = make([]uint64, len(state.Slashings))
	copy(ret.Slashings, state.Slashings)

	return ret
}

// will return nil if not found or inactive
func GetValidator (state *core.State, id uint64) *core.Validator {
	if id < uint64(len(state.Validators)) {
		return state.Validators[id]
	}
	return nil
}

/**
def is_valid_genesis_state(state: BeaconState) -> bool:
    if state.genesis_time < MIN_GENESIS_TIME:
        return False
    if len(get_active_validator_indices(state, GENESIS_EPOCH)) < MIN_GENESIS_ACTIVE_VALIDATOR_COUNT:
        return False
    return True
 */
func IsValidGenesisState(state *core.State) bool {
	if state.GenesisTime < params.ChainConfig.MinGenesisTime {
		return false
	}
	if uint64(len(GetActiveValidators(state, params.ChainConfig.GenesisEpoch))) < params.ChainConfig.MinGenesisActiveValidatorCount {
		return false
	}
	return true
}

func SumSlashings(state *core.State) uint64 {
	totalSlashing := uint64(0)
	for _, slashing := range state.Slashings {
		totalSlashing += slashing
	}
	return totalSlashing
}

func StateHashTreeRoot (state *core.State) {

}

//func computeFieldRoots(state *core.State) ([][]byte, error) {
//	if state == nil {
//		return nil, fmt.Errorf("nil state")
//	}
//	hasher := hashutil.CustomSHA256Hasher()
//	fieldRoots := make([][]byte, fieldCount)
//
//	// Genesis time root.
//	genesisRoot := htrutils.Uint64Root(state.GenesisTime)
//	fieldRoots[0] = genesisRoot[:]
//
//	// Genesis validator root.
//	r := [32]byte{}
//	copy(r[:], state.GenesisValidatorsRoot)
//	fieldRoots[1] = r[:]
//
//	// Slot root.
//	slotRoot := htrutils.Uint64Root(state.Slot)
//	fieldRoots[2] = slotRoot[:]
//
//	// Fork data structure root.
//	forkHashTreeRoot, err := htrutils2.ForkRoot(state.Fork)
//	if err != nil {
//		return nil, fmt.Errorf("could not compute fork merkleization: %s", err.Error())
//	}
//	fieldRoots[3] = forkHashTreeRoot[:]
//
//	// BeaconBlockHeader data structure root.
//	headerHashTreeRoot, err := BlockHeaderRoot(state.LatestBlockHeader)
//	if err != nil {
//		return nil, fmt.Errorf("could not compute block header merkleization: %s", err.Error())
//	}
//	fieldRoots[4] = headerHashTreeRoot[:]
//
//	// BlockRoots array root.
//	blockRootsRoot, err := h.arraysRoot(state.BlockRoots, params.ChainConfig.SlotsPerHistoricalRoot, "BlockRoots")
//	if err != nil {
//		return nil, fmt.Errorf("could not compute block roots merkleization: %s", err.Error())
//	}
//	fieldRoots[5] = blockRootsRoot[:]
//
//	// StateRoots array root.
//	stateRootsRoot, err := h.arraysRoot(state.StateRoots, params.BeaconConfig().SlotsPerHistoricalRoot, "StateRoots")
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute state roots merkleization")
//	}
//	fieldRoots[6] = stateRootsRoot[:]
//
//	// HistoricalRoots slice root.
//	historicalRootsRt, err := htrutils.HistoricalRootsRoot(state.HistoricalRoots)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute historical roots merkleization")
//	}
//	fieldRoots[7] = historicalRootsRt[:]
//
//	// Eth1Data data structure root.
//	eth1HashTreeRoot, err := Eth1Root(hasher, state.Eth1Data)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute eth1data merkleization")
//	}
//	fieldRoots[8] = eth1HashTreeRoot[:]
//
//	// Eth1DataVotes slice root.
//	eth1VotesRoot, err := Eth1DataVotesRoot(state.Eth1DataVotes)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute eth1data votes merkleization")
//	}
//	fieldRoots[9] = eth1VotesRoot[:]
//
//	// Eth1DepositIndex root.
//	eth1DepositIndexBuf := make([]byte, 8)
//	binary.LittleEndian.PutUint64(eth1DepositIndexBuf, state.Eth1DepositIndex)
//	eth1DepositBuf := bytesutil.ToBytes32(eth1DepositIndexBuf)
//	fieldRoots[10] = eth1DepositBuf[:]
//
//	// Validators slice root.
//	validatorsRoot, err := h.validatorRegistryRoot(state.Validators)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute validator registry merkleization")
//	}
//	fieldRoots[11] = validatorsRoot[:]
//
//	// Balances slice root.
//	balancesRoot, err := ValidatorBalancesRoot(state.Balances)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute validator balances merkleization")
//	}
//	fieldRoots[12] = balancesRoot[:]
//
//	// RandaoMixes array root.
//	randaoRootsRoot, err := h.arraysRoot(state.RandaoMixes, params.BeaconConfig().EpochsPerHistoricalVector, "RandaoMixes")
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute randao roots merkleization")
//	}
//	fieldRoots[13] = randaoRootsRoot[:]
//
//	// Slashings array root.
//	slashingsRootsRoot, err := htrutils.SlashingsRoot(state.Slashings)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute slashings merkleization")
//	}
//	fieldRoots[14] = slashingsRootsRoot[:]
//
//	// PreviousEpochAttestations slice root.
//	prevAttsRoot, err := h.epochAttestationsRoot(state.PreviousEpochAttestations)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute previous epoch attestations merkleization")
//	}
//	fieldRoots[15] = prevAttsRoot[:]
//
//	// CurrentEpochAttestations slice root.
//	currAttsRoot, err := h.epochAttestationsRoot(state.CurrentEpochAttestations)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute current epoch attestations merkleization")
//	}
//	fieldRoots[16] = currAttsRoot[:]
//
//	// JustificationBits root.
//	justifiedBitsRoot := bytesutil.ToBytes32(state.JustificationBits)
//	fieldRoots[17] = justifiedBitsRoot[:]
//
//	// PreviousJustifiedCheckpoint data structure root.
//	prevCheckRoot, err := htrutils.CheckpointRoot(hasher, state.PreviousJustifiedCheckpoint)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute previous justified checkpoint merkleization")
//	}
//	fieldRoots[18] = prevCheckRoot[:]
//
//	// CurrentJustifiedCheckpoint data structure root.
//	currJustRoot, err := htrutils.CheckpointRoot(hasher, state.CurrentJustifiedCheckpoint)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute current justified checkpoint merkleization")
//	}
//	fieldRoots[19] = currJustRoot[:]
//
//	// FinalizedCheckpoint data structure root.
//	finalRoot, err := htrutils.CheckpointRoot(hasher, state.FinalizedCheckpoint)
//	if err != nil {
//		return nil, errors.Wrap(err, "could not compute finalized checkpoint merkleization")
//	}
//	fieldRoots[20] = finalRoot[:]
//	return fieldRoots, nil
//}
//
//func arraysRoot(input [][]byte, length uint64, fieldName string) ([32]byte, error) {
//	hashFunc := hashutil.CustomSHA256Hasher()
//	if _, ok := layersCache[fieldName]; !ok && h.rootsCache != nil {
//		depth := htrutils.GetDepth(length)
//		layersCache[fieldName] = make([][][32]byte, depth+1)
//	}
//
//	leaves := make([][32]byte, length)
//	for i, chunk := range input {
//		copy(leaves[i][:], chunk)
//	}
//	bytesProcessed := 0
//	changedIndices := make([]int, 0)
//	prevLeaves, ok := leavesCache[fieldName]
//	if len(prevLeaves) == 0 || h.rootsCache == nil {
//		prevLeaves = leaves
//	}
//
//	for i := 0; i < len(leaves); i++ {
//		// We check if any items changed since the roots were last recomputed.
//		notEqual := leaves[i] != prevLeaves[i]
//		if ok && h.rootsCache != nil && notEqual {
//			changedIndices = append(changedIndices, i)
//		}
//		bytesProcessed += 32
//	}
//	if len(changedIndices) > 0 && h.rootsCache != nil {
//		var rt [32]byte
//		var err error
//		// If indices did change since last computation, we only recompute
//		// the modified branches in the cached Merkle tree for this state field.
//		chunks := leaves
//
//		// We need to ensure we recompute indices of the Merkle tree which
//		// changed in-between calls to this function. This check adds an offset
//		// to the recomputed indices to ensure we do so evenly.
//		maxChangedIndex := changedIndices[len(changedIndices)-1]
//		if maxChangedIndex+2 == len(chunks) && maxChangedIndex%2 != 0 {
//			changedIndices = append(changedIndices, maxChangedIndex+1)
//		}
//		for i := 0; i < len(changedIndices); i++ {
//			rt, err = recomputeRoot(changedIndices[i], chunks, fieldName, hashFunc)
//			if err != nil {
//				return [32]byte{}, err
//			}
//		}
//		leavesCache[fieldName] = chunks
//		return rt, nil
//	}
//
//	res := h.merkleizeWithCache(leaves, length, fieldName, hashFunc)
//	if h.rootsCache != nil {
//		leavesCache[fieldName] = leaves
//	}
//	return res, nil
//}