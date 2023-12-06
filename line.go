package twelve

import (
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/utils/errors"
	"sync"
	"time"
)

type Job struct {
	letter    *Letter
	done      bool
	timestamp time.Time
}

type Line struct {
	inv       bid.BID
	arrows    [4]*Job
	timestamp time.Time
	lock      sync.Mutex
}

func NewLine(letter *Letter) *Line {
	ln := &Line{
		inv:    letter.Invariable,
		arrows: [4]*Job{nil, nil, nil, nil},
	}
	ln.doSet(letter)
	return ln
}

//
//func (line *Line) IsDone() bool {
//	line.lock.Lock()
//	defer line.lock.Unlock()
//	for i := 0; i < len(line.arrows); i++ {
//		if line.arrows[i] == nil {
//			return false
//		}
//	}
//	return true
//}
//
//func (line *Line) IsRequest() bool {
//	return line.doIs(RequestArrow)
//}
//
//func (line *Line) IsPrepare() bool {
//	return line.doIs(PrepareArrow)
//}
//
//func (line *Line) IsCommitted() bool {
//	return line.doIs(CommittedArrow)
//}
//
//func (line *Line) IsConfirmed() bool {
//	return line.doIs(ConfirmedArrow)
//}
//
//func (line *Line) doIs(arrow Arrow) bool {
//	idx := line.doGetIdx(arrow)
//	if idx < 0 {
//		return false
//	}
//	return line.arrows[idx] != nil
//}

func (line *Line) RunOrRegister(letter *Letter, runFunc func(letter *Letter) *errors.Error) *errors.Error {
	idx := line.doGetIdx(letter.Arrow)
	if idx < 0 {
		return errors.Verify(ErrInvalidLetter)
	}
	line.lock.Lock()
	defer line.lock.Unlock()
	if line.arrows[idx] == nil {
		line.arrows[idx] = &Job{
			letter:    letter,
			done:      false,
			timestamp: time.Now(),
		}
		line.timestamp = time.Now()
	}
	preOk := true
	for i := 0; i < 4; i++ {
		if !preOk {
			return nil
		}
		if line.arrows[i] == nil {
			return nil
		}
		if line.arrows[i].done {
			preOk = true
			continue
		} else {
			err := runFunc(line.arrows[i].letter)
			if err != nil {
				return err
			}
			line.arrows[i].done = true
			line.arrows[i].timestamp = time.Now()
		}
		preOk = line.arrows[i].done
	}
	return nil
}

func (line *Line) doSet(letter *Letter) {
	idx := line.doGetIdx(letter.Arrow)
	if idx < 0 {
		return
	}
	line.lock.Lock()
	defer line.lock.Unlock()
	if line.arrows[idx] == nil {
		line.arrows[idx] = &Job{
			letter:    letter,
			done:      false,
			timestamp: time.Now(),
		}
		line.timestamp = time.Now()
	}
}

func (line *Line) doGetIdx(arrow Arrow) int {
	idx := -1
	switch arrow {
	case RequestArrow:
		idx = 0
	case PrepareArrow:
		idx = 1
	case CommittedArrow:
		idx = 2
	case ConfirmedArrow:
		idx = 3
	}
	return idx
}

type LineFactory struct {
	db   sync.Map
	lock sync.Mutex
}

func NewLineFactory() *LineFactory {
	return &LineFactory{}
}

func (lf *LineFactory) Exists(inv bid.BID) bool {
	_, ok := lf.db.Load(inv)
	return ok
}

func (lf *LineFactory) MustGet(letter *Letter) *Line {
	lf.lock.Lock()
	defer lf.lock.Unlock()
	itf, ok := lf.db.Load(letter.Invariable)
	if !ok {
		ln := NewLine(letter)
		lf.db.Store(letter.Invariable, ln)
		return ln
	}
	return itf.(*Line)
}
