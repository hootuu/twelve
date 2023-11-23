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
	Hash   string
	State  State
	Letter *Letter
	Pre    string
	Height int64
}

type ImmutableTx struct {
	Hash   string
	Letter *Letter
	Pre    string
	Height int64
}

const (
	HeadTxHash = "$"
)

func HeadTx() *Tx {
	return &Tx{
		Hash:   HeadTxHash,
		State:  Confirmed,
		Letter: nil,
		Pre:    HeadTxHash,
		Height: 0,
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
		Hash:   tx.Hash,
		Letter: tx.Letter,
		Pre:    tx.Pre,
		Height: tx.Height,
	}
}

func (tx *Tx) Confirm() {
	tx.State = Confirmed
}

func (tx *Tx) Nxt(letter *Letter) *Tx {
	txM := &Tx{
		Hash:   letter.Signature.Hash,
		State:  Committed,
		Letter: letter,
		Pre:    tx.Hash,
		Height: tx.Height + 1,
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
