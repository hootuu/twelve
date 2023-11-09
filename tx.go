package twelve

import (
	"encoding/json"
	"github.com/hootuu/utils/errors"
)

type State int8

const (
	Committed State = 1
	Pending   State = 2
	Confirmed State = 9
)

func (s State) IsConfirmed() bool {
	return s == Confirmed
}

type Tx struct {
	Hash    string
	State   State
	Message *Message
	Pre     string
}

type ImmutableTx struct {
	Hash    string
	Message *Message
	Pre     string
}

const (
	HeadTxHash = "$"
)

func HeadTx() *Tx {
	return &Tx{
		Hash:    HeadTxHash,
		State:   Confirmed,
		Message: nil,
		Pre:     HeadTxHash,
	}
}

func IsHead(hash string) bool {
	return hash == HeadTxHash
}

func (tx *Tx) IsHead() bool {
	return IsHead(tx.Hash)
}

func (tx *Tx) Immutable() *ImmutableTx {
	return &ImmutableTx{
		Hash:    tx.Hash,
		Message: tx.Message,
		Pre:     tx.Pre,
	}
}

func (tx *Tx) Confirm() {
	tx.State = Confirmed
}

func (tx *Tx) Nxt(msg *Message) *Tx {
	txM := &Tx{
		Hash:    msg.ID,
		State:   Committed,
		Message: msg,
		Pre:     tx.Hash,
	}
	if !tx.State.IsConfirmed() {
		txM.State = Pending
	}
	return txM
}

func TxOf(byteData []byte) (*Tx, *errors.Error) {
	var txM Tx
	nErr := json.Unmarshal(byteData, &txM)
	if nErr != nil {
		return nil, errors.Verify("invalid tx byte data: " + nErr.Error())
	}
	return &txM, nil
}
