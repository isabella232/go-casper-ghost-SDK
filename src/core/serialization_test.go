package core

import (
	"encoding/hex"
)

func toByte(str string) []byte {
	ret, _ := hex.DecodeString(str)
	return ret
}

//func TestBlockBodySSZ(t *testing.T) {
//	tests := []struct{
//		testName string
//		body *BlockBody
//		expected []byte
//	}{
//		{
//			testName: "full SSZ",
//			body: &BlockBody{
//				Proposer:           12,
//				Epoch:              5,
//				ExecutionSummaries: []*ExecutionSummary{
//					&ExecutionSummary{
//						PoolId:        12,
//						Epoch:         5,
//						Duties:        []*BeaconDuty {
//							&BeaconDuty{
//								Type:          0, // attestation
//								Committee:     12,
//								Slot:         342,
//								Finalized:     true,
//								Participation: []byte{1,3,88,12,43,12,89,35,1,0,99,16,63,13,33,0},
//							},
//							&BeaconDuty{
//								Type:          1, // proposal
//								Committee:     0,
//								Slot:         343,
//								Finalized:     true,
//								Participation: []byte{},
//							},
//						},
//					},
//				},
//				NewPoolReq:         []*CreateNewPoolRequest{
//					&CreateNewPoolRequest{
//						Id:                  3,
//						Status:              0, // started
//						StartEpoch:          5,
//						EndEpoch:            6,
//						LeaderBlockProducer: 15,
//						CreatePubKey:        toByte("public key"),
//						Participation:       []byte{43,12,89,35,99,16,63,13,33,0,1,3,88,12,43,1},
//					},
//				},
//				ParentBlockRoot:    toByte("parent block root parent block root parent block root parent block root"),
//			},
//			expected:toByte("2a36efe0b9c926c269f77eb22bbe62216c7a518676ad02f1e64a27177e5a8ca2"),
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.testName, func(t *testing.T) {
//			root,err := ssz.HashTreeRoot(test.body)
//			require.NoError(t, err)
//			require.EqualValues(t, test.expected, root[:])
//		})
//	}
//}