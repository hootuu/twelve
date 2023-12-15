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
	Lock   Lock
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
		Lock:   GenesisLock,
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

func (tx *Tx) ImmutableHash() string {
	return tx.Letter.Invariable.S()
}

func (tx *Tx) Confirm() {
	tx.State = Confirmed
}

func (tx *Tx) Nxt(letter *Letter) *Tx {
	nxtLock := tx.Lock.Nxt(tx.Height+1, letter.Invariable)
	txM := &Tx{
		Hash:   letter.Invariable.S(),
		State:  Committed,
		Letter: letter,
		Lock:   nxtLock,
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
