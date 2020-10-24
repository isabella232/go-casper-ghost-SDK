package state_transition

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/bloxapp/go-casper-ghost-SDK/src/core"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared"
	"github.com/bloxapp/go-casper-ghost-SDK/src/shared/params"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/shared/trieutil"
	"github.com/ulule/deepcopier"
	"log"
	"time"
)

func toByte(str string) []byte {
	ret, _ := hex.DecodeString(str)
	return ret
}

func defaultEth1Data() (*core.ETH1Data, error) {
	trie, err := trieutil.NewTrie(int(params.ChainConfig.DepositContractTreeDepth))
	if err != nil {
		return nil, err
	}
	depositRoot := trie.Root()

	return &core.ETH1Data{
		DepositRoot:          depositRoot[:],
		DepositCount:         0,
		BlockHash:            toByte("8049cb6db44dc94a52a9702779c0b4d5d77164bc56a913da1203563de5193405"),
	}, nil
}

func initBlockHeader() (*core.BlockHeader, error) {
	root, err := ssz.HashTreeRoot(&core.BlockBody{})
	if err != nil {
		return nil, err
	}
	return &core.BlockHeader{
		Slot:                 0,
		ProposerIndex:        0,
		ParentRoot:           params.ChainConfig.ZeroHash,
		StateRoot:            params.ChainConfig.ZeroHash,
		BodyRoot:             root[:],
	}, nil
}

type StateTestContext struct {
	State *core.State
}

func NewStateTestContext(config *core.ChainConfig, eth1Data *core.ETH1Data, genesisTime uint64) *StateTestContext {
	start := time.Now()

	var err error
	if eth1Data == nil {
		eth1Data, err = defaultEth1Data()
		if err != nil {
			log.Fatal(err)
		}
	}

	initBlockHeader, err := initBlockHeader()
	if err != nil {
		log.Fatal(err)
	}

	randaoMixes := make([][]byte, params.ChainConfig.EpochsPerHistoricalVector)
	for i := range randaoMixes {
		randaoMixes[i] = eth1Data.BlockHash
	}

	blockRoots := make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot)
	for i := range blockRoots {
		blockRoots[i] = params.ChainConfig.ZeroHash
	}

	stateRoots := make([][]byte, params.ChainConfig.SlotsPerHistoricalRoot)
	for i := range stateRoots {
		stateRoots[i] = params.ChainConfig.ZeroHash
	}

	genesisValidatorRoot, err := ssz.HashTreeRoot([]*core.Validator{})
	if err != nil {
		log.Fatal(err)
	}

	ret := &StateTestContext{
		State: &core.State{
			GenesisTime:       genesisTime,
			Slot:              0,
			LatestBlockHeader: initBlockHeader,
			Fork:                        &core.Fork{
				PreviousVersion:      config.GenesisForkVersion,
				CurrentVersion:       config.GenesisForkVersion,
				Epoch:                config.GenesisEpoch,
			},
			BlockRoots:                blockRoots,
			StateRoots:                stateRoots,
			RandaoMix:                 randaoMixes,
			HistoricalRoots:           [][]byte{},
			GenesisValidatorsRoot:     genesisValidatorRoot[:],
			PreviousEpochAttestations: []*core.PendingAttestation{},
			CurrentEpochAttestations:  []*core.PendingAttestation{},
			JustificationBits:         []byte{0},
			PreviousJustifiedCheckpoint: &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			CurrentJustifiedCheckpoint:  &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			FinalizedCheckpoint:         &core.Checkpoint{
				Epoch:                0,
				Root:                 params.ChainConfig.ZeroHash,
			},
			Eth1Data:                    eth1Data,
			Eth1DataVotes:               []*core.ETH1Data{},
			Eth1DepositIndex:            0,
			Validators:                  []*core.Validator{},
			Balances: 					 []uint64{},
			Slashings:                   make([]uint64, params.ChainConfig.EpochsPerSlashingVector),
		},
	}

	end := time.Now()
	log.Printf("state ctx generate: %f\n", end.Sub(start).Seconds())

	return ret
}

func (c *StateTestContext) PopulateGenesisValidator(validatorIndexEnd uint64) *StateTestContext {
	if err := bls.Init(bls.BLS12_381); err != nil {
		log.Fatal(err)
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		log.Fatal(err)
	}

	start := time.Now()

	var leaves [][]byte
	deposits := make([]*core.Deposit, validatorIndexEnd)
	for i := uint64(0) ; i < validatorIndexEnd ; i++ {
		sk := &bls.SecretKey{}
		sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", uint64(i)))))

		cred, err := ssz.HashTreeRoot([]byte("test_withdrawal_cred"))
		if err != nil {
			log.Fatal(err)
		}

		depositMessage := &core.DepositMessage{
			PublicKey:             sk.GetPublicKey().Serialize(),
			WithdrawalCredentials: cred[:],
			Amount:                32 * 1e9, // gwei
		}
		root, err := ssz.HashTreeRoot(depositMessage)
		if err != nil {
			log.Fatal(err)
		}

		// insert into the trie
		depositData := &core.Deposit_DepositData{
			PublicKey:             depositMessage.PublicKey,
			WithdrawalCredentials: depositMessage.WithdrawalCredentials,
			Amount:                depositMessage.Amount,
			Signature:             sk.SignByte(root[:]).Serialize(),
		}
		deposits[i] = &core.Deposit{
			Proof:                nil,
			Data:                 depositData,
		}
		root, err = ssz.HashTreeRoot(depositData)
		if err != nil {
			log.Fatal(err)
		}
		leaves = append(leaves, root[:])
	}

	vals := time.Now()
	log.Printf("create vals: %f\n", vals.Sub(start).Seconds())

	var trie *trieutil.SparseMerkleTrie
	var err error
	if len(leaves) > 0 {
		trie, err = trieutil.GenerateTrieFromItems(leaves, int(params.ChainConfig.DepositContractTreeDepth))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		trie, err = trieutil.NewTrie(int(params.ChainConfig.DepositContractTreeDepth))
		if err != nil {
			log.Fatal(err)
		}
	}

	depositRoot := trie.Root()
	c.State.Eth1Data.DepositRoot = depositRoot[:]

	// process
	for _, deposit := range deposits {
		// Add validator and balance entries
		v := GetValidatorFromDeposit(c.State, deposit)
		v.ActivationEpoch = 0
		v.ActivationEligibilityEpoch = 0
		c.State.Validators = append(c.State.Validators, v)
		c.State.Balances = append(c.State.Balances, deposit.Data.Amount)
	}

	// update genesis root
	genesisValidatorRoot, err := ssz.HashTreeRoot(c.State.Validators)
	if err != nil {
		log.Fatal(err)
	}
	c.State.GenesisValidatorsRoot = genesisValidatorRoot[:]

	depo := time.Now()
	log.Printf("generate deposits: %f\n", depo.Sub(vals).Seconds())

	return c
}

// will generate and save blocks from slot 0 until maxBlocks
func (c *StateTestContext) ProgressSlotsAndEpochs(maxBlocks int, justifiedEpoch uint64, finalizedEpoch uint64) *StateTestContext {
	var previousBlockHeader *core.BlockHeader
	previousBlockHeader, err := initBlockHeader()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0 ; i < maxBlocks ; i++ {
		log.Printf("progressing block %d\n", i)

		start := time.Now()

		// get proposer
		stateCopy := shared.CopyState(c.State)
		cpy := time.Now()
		log.Printf("cpy: %f\n", cpy.Sub(start).Seconds())


		// we increment it before we gt proposer as this is what happens in real block processing
		// in slot > 0 the process slots is called before process block which increments slot by 1
		if i != 0 {
			stateCopy.Slot++
		}
		pID, err := shared.GetBlockProposerIndex(stateCopy)
		if err != nil {
			log.Fatal(err)
		}
		sk := []byte(fmt.Sprintf("%d", pID))

		// randao
		randaoReveal, err := signRandao(c.State, sk)
		if err != nil {
			log.Fatal(err)
		}

		// parent
		// replicates what the next process slot does
		if i != 0 {
			stateRoot,err := c.State.HashTreeRoot()
			if err != nil {
				log.Fatal(err)
			}
			previousBlockHeader.StateRoot =  stateRoot[:]
		}
		parentRoot,err := previousBlockHeader.HashTreeRoot()
		if err != nil {
			log.Fatal(err)
		}

		pre := time.Now()
		log.Printf("pre: %f\n", pre.Sub(start).Seconds())

		eth1Vote, err := defaultEth1Data() // TODO - block eth1 vote dynamic?
		if err != nil {
			log.Fatal(err)
		}

		block := &core.Block{
			Slot:                 uint64(i),
			Proposer:             pID,
			ParentRoot:           parentRoot[:],
			StateRoot:            params.ChainConfig.ZeroHash,
			Body:                 &core.BlockBody{
				RandaoReveal:         randaoReveal.Serialize(),
				Attestations:         []*core.Attestation{},
				Eth1Data: 			  eth1Vote,
			},
		}
		populateAttestations(c.State, block, uint64(i), justifiedEpoch, finalizedEpoch, block.ParentRoot)

		att := time.Now()
		log.Printf("att: %f\n", att.Sub(pre).Seconds())

		// process
		st := NewStateTransition()
		// compute state root
		root, err := st.ComputeStateRoot(c.State, &core.SignedBlock{
			Block:                block,
			Signature:            []byte{},
		})
		if err != nil {
			log.Fatal(err)
		}
		block.StateRoot = root[:]

		compu := time.Now()
		log.Printf("compu: %f\n", compu.Sub(att).Seconds())

		// sign
		blockDomain, err := shared.GetDomain(c.State, params.ChainConfig.DomainBeaconProposer, shared.GetCurrentEpoch(c.State))
		if err != nil {
			log.Fatal(err)
		}
		sig, err := shared.SignBlock(block, sk, blockDomain) // TODO - dynamic domain
		if err != nil {
			log.Fatal(err)
		}

		sign := time.Now()
		log.Printf("sign: %f\n", sign.Sub(compu).Seconds())

		// execute
		newState, err := st.ExecuteStateTransition(c.State, &core.SignedBlock{
			Block:                block,
			Signature:            sig.Serialize(),
		})
		if err != nil {
			log.Fatal(err)
		} else {
			c.State = newState
		}

		exe := time.Now()
		log.Printf("exe: %f\n", exe.Sub(sign).Seconds())

		// copy to previousBlockRoot
		previousBlockHeader = &core.BlockHeader{}
		deepcopier.Copy(c.State.LatestBlockHeader).To(previousBlockHeader)

		// Print epoch summary.
		if uint64(i + 1) % params.ChainConfig.SlotsInEpoch == 0 {
			str := "\n\n#########\nEpoch %d\n"
			str += "Pre justified epoch: %d\n"
			str += "Current justified epoch: %d\n"
			str += "Finalized epoch: %d\n"
			str += "#########\n\n"
			log.Printf(str, shared.ComputeEpochAtSlot(uint64(i)), c.State.PreviousJustifiedCheckpoint.Epoch, c.State.CurrentJustifiedCheckpoint.Epoch, c.State.FinalizedCheckpoint.Epoch)
		}
	}
	return c
}

func populateAttestations(state *core.State, block *core.Block, slot uint64, justifiedEpoch uint64, finalizedEpoch uint64, headRoot []byte) {
	if slot == 0 { // TODO - attestations at slot 0?
		return // start from slot 1 forward
	}

	// every block we collect attestations "broadcasted" in the previous slot
	slotEpoch := shared.ComputeEpochAtSlot(slot-1)

	nextStateCopy := shared.CopyState(state)
	nextStateCopy.Slot ++

	for i := uint64(0) ; i < shared.GetCommitteeCountPerSlot(nextStateCopy, slot-1) ; i++{
		// get attestation to sign
		var targetRoot []byte
		var err error
		targetRoot, err = shared.GetBlockRoot(nextStateCopy, slotEpoch)
		if err != nil {
			log.Fatalf("populateAttestations: %s", err.Error())
		}
		if bytes.Equal(targetRoot, params.ChainConfig.ZeroHash) {
			targetRoot = headRoot
		}

		data := &core.AttestationData{
			Slot:                 slot - 1,
			CommitteeIndex:       i,
			BeaconBlockRoot:      nextStateCopy.LatestBlockHeader.BodyRoot,
			Source:               &core.Checkpoint{},
			Target:               &core.Checkpoint{
				Epoch:                slotEpoch,
				Root:                 targetRoot,
			},
		}
		deepcopier.Copy(state.CurrentJustifiedCheckpoint).To(data.Source)

		// root
		root, err := ssz.HashTreeRoot(data)
		if err != nil {
			log.Fatalf("populateAttestations: %s", err.Error())
		}

		// sign
		indices, err := shared.GetAttestationCommittee(nextStateCopy, slot-1, i)
		if err != nil {
			log.Fatalf("populateAttestations: %s", err.Error())
		}
		var aggregatedSig *bls.Sign
		aggBits := make(bitfield.Bitlist, len(indices)) // for bytes
		signed := uint64(0)
		for aggIndex, index := range indices {
			sk := &bls.SecretKey{}
			sk.SetHexString(hex.EncodeToString([]byte(fmt.Sprintf("%d", index))))

			// sign
			if aggregatedSig == nil {
				aggregatedSig = sk.SignByte(root[:])
			} else {
				aggregatedSig.Add(sk.SignByte(root[:]))
			}
			aggBits.SetBitAt(uint64(aggIndex), true)
			signed ++

			if slotEpoch <= justifiedEpoch + 1 || slotEpoch <= finalizedEpoch + 1 {
				// vote @ 2/3
				if signed * 2 >= 2 * uint64(len(indices)) {
					break
				}
			} else {
				// vote @ 1/3
				if signed * 3 >= 1 * uint64(len(indices)) {
					break
				}
			}
		}

		block.Body.Attestations = append(block.Body.Attestations, &core.Attestation{
			AggregationBits:      aggBits,
			Data:                 data,
			Signature:            aggregatedSig.Serialize(),
		})
	}
}

func signRandao(state *core.State, sk []byte) (*bls.Sign, error) {
	// We bump the slot by one to accurately calculate the epoch as when this
	// randao reveal will be verified the state.Slot will be +1
	copyState := shared.CopyState(state)
	copyState.Slot++

	data, domain, err := RANDAOSigningData(copyState)
	if err != nil {
		return nil, err
	}
	return shared.SignRandao(data, domain, sk)
}