package twelve

import (
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
	"sync"
)

type txNode struct {
	tx  *Tx
	pre *txNode
	nxt *txNode
}

func txNodeOf(txM *Tx) *txNode {
	return &txNode{
		tx:  txM,
		pre: nil,
		nxt: nil,
	}
}

func txNodeHead() *txNode {

}

func (txN *txNode) Nxt(txM *Tx) *txNode {
	newNode := &txNode{
		tx:  txM,
		pre: txN,
		nxt: nil,
	}
	txN.nxt = newNode
	return newNode
}

func (txN *txNode) Backward(callback func(txM *Tx)) {
	cur := txN
	for cur.pre != nil {
		callback(cur.tx)
	}
}

func (txN *txNode) Forward(callback func(txM *Tx)) {
	cur := txN
	for cur.nxt != nil {
		callback(cur.tx)
	}
}

type Rope struct {
	chain   *kt.Chain
	tail    *kt.Knot
	bufHead *txNode
	bufTail *txNode
	txCang  *TxCang
	lock    sync.Mutex
}

func NewRope(chain *kt.Chain) (*Rope, *errors.Error) {
	r := &Rope{
		chain:   chain,
		tail:    kt.Genesis(chain),
		bufTail: HeadTx(),
	}
	var err *errors.Error
	r.txCang, err = NewTxCang(chain.Simple(), ".twelve/"+r.chain.Vn.S())
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Rope) Tail(callback func(tail *kt.Knot)) {
	r.lock.Lock()
	defer r.lock.Unlock()
	callback(r.tail)
}

func (r *Rope) BufExists(hash kt.Hash, callback func(exists bool)) *errors.Error {
	r.lock.Lock()
	defer r.lock.Unlock()
	b, err := r.txCang.QCollection(hash).Exists(hash)
	if err != nil {
		return err
	}
	callback(b)
	return nil
}

func (r *Rope) BufMustGet(hash kt.Hash, callback func(txM *Tx)) *errors.Error {
	r.lock.Lock()
	defer r.lock.Unlock()
	var txM Tx
	err := r.txCang.QCollection(hash).MustGet(hash, &txM)
	if err != nil {
		gLogger.Error("r.txCang.QCollection(hash).MustGet(hash, &txM) error",
			zap.String("hash", hash))
		return err
	}
	callback(&txM)
	return nil
}

func (r *Rope) PostTie(letter *Letter, callback func(txM *Tx)) *errors.Error {
	r.lock.Lock()
	r.lock.Unlock()
	if !r.bufTail.Before(letter.Signature.Timestamp) {

	}
	txM := r.bufTail.Nxt(letter)
	exists, err := r.txCang.QCollection(txM.Hash).Exists(txM.Hash)
	if err != nil {
		gLogger.Error("r.txCang.QCollection(txM.Hash).Exists(txM.Hash) error",
			zap.String("hash", txM.Hash))
		return err
	}
	if exists {
		gLogger.Error("PostTie Repeated", zap.String("hash", txM.Hash))
		return errors.Sys("repeated append: " + txM.Hash)
	}
	if sys.RunMode.IsRd() {
		sys.Info("Post Tie [ Chain:", r.chain.S(), ", Hash:", txM.Hash, ", Pre: ", txM.Pre, " ]")
	}
	err = r.txCang.QCollection(txM.Hash).Put(txM.Hash, txM)
	if err != nil {
		gLogger.Error("r.txCang.QCollection(txM.Hash).Put(txM.Hash, txM) error",
			zap.String("hash", txM.Hash))
		return err
	}
	r.bufTail = txM
	callback(txM)
	return nil
}

func (r *Rope) Confirm(hash kt.Hash, callback func(knot *kt.Knot)) *errors.Error {
	r.lock.Lock()
	r.lock.Unlock()
	var txM *Tx
	err := r.BufMustGet(hash, func(gTxM *Tx) {
		txM = gTxM
	})
	if err != nil {
		return err
	}
	tx.Confirm()
	err = q.txCang.QCollection(tx.ImmutableHash()).Put(tx.ImmutableHash(), tx)
	if err != nil {
		return nil, err
	}
	immutableTx := tx.Immutable()
	gLogger.Info("confirm_tx", zap.String("hash", immutableTx.Hash), zap.String(" pre: ", immutableTx.Pre))
	sys.Info("confirm tx ", immutableTx.Hash, " pre: ", immutableTx.Pre)
	gChainLogger.Info("MING KE", zap.Any("TX", immutableTx))
	err = q.txCang.CCollection(immutableTx.Hash).Put(immutableTx.Hash, immutableTx)
	if err != nil {
		return nil, err
	}
	q.immutableTail = immutableTx
	return immutableTx, nil
}
