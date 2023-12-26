package twelve

import (
	"encoding/json"
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/tome/ki"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/utils/errors"
	"sync"
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

type TxNode struct {
	Tx    *kt.Tx  `json:"tx"`
	Pre   kt.Hash `json:"pre"`
	Nxt   kt.Hash `json:"nxt"`
	State State   `json:"state"`
}

const (
	maxMemTxNodeCount = 12 * 12
)

type TxLink struct {
	chain   *kt.Chain
	tail    *TxNode
	memTxDB sync.Map
	txCang  *TxCang
	lock    sync.Mutex
}

func NewTxLink(chain *kt.Chain) (*TxLink, *errors.Error) {
	link := &TxLink{
		tail: &TxNode{
			Tx:    kt.HeadTx(chain),
			Pre:   "",
			Nxt:   "",
			State: Confirmed,
		},
		chain:  chain,
		txCang: nil,
	}
	var err *errors.Error
	link.txCang, err = NewTxCang(chain.Simple(), ".twelve/"+chain.Vn.S())
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (link *TxLink) Get(hx kt.Hash, callback func(tn *TxNode)) *errors.Error {
	link.lock.Lock()
	defer link.lock.Unlock()
	val, ok := link.memTxDB.Load(hx)
	if !ok {
		tn, err := link.doGetFromCang(hx)
		if err != nil {
			return err
		}
		callback(tn)
		return nil
	}
	if tn, ok := val.(*TxNode); ok {
		callback(tn)
		return nil
	}
	return errors.Sys("Must Be TxNode")
}

func (link *TxLink) Inject(tx *kt.Tx) (TxNode, *errors.Error) {
	link.lock.Lock()
	defer link.lock.Unlock()
	//todo
	if !link.tail.Tx.Timestamp.Before(tx.Timestamp) {

	}
	//todo
	return TxNode{}, nil
}

func (link *TxLink) doGetFromCang(hx kt.Hash) (*TxNode, *errors.Error) {
	var txN TxNode
	err := link.txCang.QCollection(hx).MustGet(hx, &txN)
	if err != nil {
		return nil, err
	}
	return &txN, nil
}

type Tx struct {
	Hash       kt.Hash      `json:"hx"`
	State      State        `json:"s"`
	Chain      *kt.Chain    `json:"c"`
	Invariable bid.BID      `json:"i"`
	Timestamp  kt.Timestamp `json:"t"`
	Signer     ki.ADR       `json:"signer"`
	Signature  []byte       `json:"signature"`
	Pre        kt.KID       `json:"p"`
	Height     kt.Height    `json:"h"`

	previous *Tx
}

func txOf(letter *Letter) (*Tx, *errors.Error) {
	if letter.Arrow != RequestArrow {
		gLogger.Error("txOf(letter) With Letter.RequestArrow")
		return nil, errors.Verify("TxOf Must With Letter.RequestArrow")
	}
	return &Tx{
		Hash:       letter.Signature.Hash(),
		State:      Committed,
		Chain:      letter.Chain,
		Invariable: letter.Invariable,
		Timestamp:  letter.Signature.Timestamp,
		Signer:     letter.Signature.Signer,
		Signature:  letter.Signature.Signature,
		Pre:        "",
		Height:     0,
		previous:   nil,
	}, nil
}

func (tx *Tx) Link(letter *Letter) (*Tx, *errors.Error) {
	newTx, err := txOf(letter)
	if err != nil {
		return nil, err
	}
	return nil, nil
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
		Lock:   bid.GenesisBID,
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

func (tx *Tx) Before(t kt.Timestamp) bool {
	return tx.Letter.Signature.Timestamp < t
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
	nxtLock := tx.Lock //todo
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
