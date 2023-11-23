package twelve

import (
	"fmt"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/tome/vn"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"github.com/hootuu/utils/types/pagination"
	"go.uber.org/zap"
	"sync"
)

const (
	FirstHash = "#"
)

type Queue struct {
	tail          *Tx
	immutableTail *ImmutableTx
	txCang        *TxCang
	lock          sync.RWMutex
}

var gQueueIdx = 0 //todo will delete

func NewQueue(vnID vn.ID, chain kt.Chain) (*Queue, *errors.Error) {
	q := &Queue{
		tail:          HeadTx(),
		immutableTail: HeadTx().Immutable(),
		lock:          sync.RWMutex{},
	}
	var err *errors.Error
	q.txCang, err = NewTxCang(chain.S(), ".twelve/"+vnID.S()+"/"+fmt.Sprintf("%d", gQueueIdx))
	gQueueIdx++ //todo will delete
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (q *Queue) Tail() (*Tx, *errors.Error) {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if q.tail == nil {
		return nil, nil
	}
	return q.tail, nil
}

func (q *Queue) MustGet(hash string) (*Tx, *errors.Error) {
	var txM Tx
	err := q.txCang.QCollection(hash).MustGet(hash, &txM)
	if err != nil {
		return nil, err
	}
	return &txM, nil
}

func (q *Queue) ImmutableMustGet(hash string) (*ImmutableTx, *errors.Error) {
	var txM ImmutableTx
	err := q.txCang.CCollection(hash).MustGet(hash, &txM)
	if err != nil {
		return nil, err
	}
	return &txM, nil
}

func (q *Queue) Append(letter *Letter) (*Tx, *errors.Error) {
	q.lock.Lock()
	q.lock.Unlock()
	txM := q.tail.Nxt(letter)
	exists, err := q.txCang.QCollection(txM.Hash).Exists(txM.Hash)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Sys("repeated append: " + txM.Hash)
	}

	sys.Info("append tx ", txM.Hash, " pre: ", txM.Pre)
	err = q.txCang.QCollection(txM.Hash).Put(txM.Hash, txM)
	if err != nil {
		return nil, err
	}
	q.tail = txM
	return txM, nil
}

func (q *Queue) Confirm(hash string) (*ImmutableTx, *errors.Error) {
	q.lock.Lock()
	q.lock.Unlock()
	tx, err := q.MustGet(hash)
	if err != nil {
		return nil, err
	}
	tx.Confirm()
	err = q.txCang.QCollection(tx.Hash).Put(tx.Hash, tx)
	if err != nil {
		return nil, err
	}
	immutableTx := tx.Immutable()
	gLogger.Info("confirm_tx", zap.String("hash", immutableTx.Hash), zap.String(" pre: ", immutableTx.Pre))
	sys.Info("confirm tx ", immutableTx.Hash, " pre: ", immutableTx.Pre)
	err = q.txCang.CCollection(immutableTx.Hash).Put(immutableTx.Hash, immutableTx)
	if err != nil {
		return nil, err
	}
	q.immutableTail = immutableTx
	return immutableTx, nil
}

func (q *Queue) ImmutableList(lstHash string, limit pagination.Limit) ([]*ImmutableTx, string, *errors.Error) {
	if lstHash == FirstHash || len(lstHash) == 0 {
		lstHash = q.immutableTail.Hash
	}
	if IsHead(lstHash) {
		return []*ImmutableTx{HeadTx().Immutable()}, HeadTxHash, nil
	}
	var arr []*ImmutableTx
	current := lstHash
	err := limit.Iter(func(i int64) (bool, *errors.Error) {
		if IsHead(current) {
			arr = append(arr, HeadTx().Immutable())
			return true, nil
		}
		tx, innerErr := q.ImmutableMustGet(current)
		if innerErr != nil {
			return false, innerErr
		}
		arr = append(arr, tx)
		current = tx.Pre
		return false, nil
	}, false)
	if err != nil {
		return nil, "", err
	}

	return arr, current, nil
}

func (q *Queue) List(lstHash string, limit pagination.Limit) ([]*Tx, string, *errors.Error) {
	if lstHash == FirstHash || len(lstHash) == 0 {
		lstHash = q.tail.Hash
	}
	if IsHead(lstHash) {
		return []*Tx{HeadTx()}, HeadTxHash, nil
	}
	var arr []*Tx
	current := lstHash
	err := limit.Iter(func(i int64) (bool, *errors.Error) {
		tx, innerErr := q.MustGet(current)
		if innerErr != nil {
			return false, innerErr
		}
		arr = append(arr, tx)
		current = tx.Pre
		if IsHead(current) {
			return true, nil
		}
		return false, nil
	}, false)
	if err != nil {
		return nil, "", err
	}

	return arr, current, nil
}
