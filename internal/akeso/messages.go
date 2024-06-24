package akeso

import (
	"encoding/json"

	"github.com/etclab/art"
)

type GroupSetupMessage struct {
	EKeys       map[string][]byte `json:"EKeys`
	IKeys       map[string][]byte `json:"IKeys"`
	InPubKey    []byte            `json:"InPubKey`
	SetupMsg    art.SetupMessage  `json:"SetupMsg"`
	SetupMsgSig []byte            `json:"SetupMsgSig"`
}

type KeyUpdateMessage struct {
	UpdatedBy    string            `json:"UpdatedBy"`
	UpdateMsg    art.UpdateMessage `json:"UpdateMsg"`
	UpdateMsgMac []byte            `json:"UpdateMsgMac"`
}

func (sgMessage *GroupSetupMessage) UnmarshalJSON(b []byte) error {
	type Dup GroupSetupMessage

	tmp := struct {
		SetupMsg []byte `json:"SetupMsg"`
		*Dup
	}{
		Dup: (*Dup)(sgMessage),
	}

	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}

	err = json.Unmarshal(tmp.SetupMsg, &sgMessage.SetupMsg)
	if err != nil {
		return err
	}

	return nil
}

type KeyPairMessage struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}
