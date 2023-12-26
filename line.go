package twelve

import (
	"github.com/hootuu/tome/bk/bid"
	"github.com/hootuu/tome/kt"
	"github.com/hootuu/utils/errors"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Job struct {
	letter    *Letter
	done      bool
	timestamp time.Time
}

type JobDict struct {
	dict map[kt.Hash]*Job
}

func NewJobDict() *JobDict {
	return &JobDict{dict: make(map[kt.Hash]*Job)}
}

func (jd *JobDict) doHasDone() bool {
	for _, job := range jd.dict {
		if job.done {
			return true
		}
	}
	return false
}

func (jd *JobDict) doOncePut(letter *Letter) {
	jd.dict[letter.Signature.Hash()] = &Job{
		letter:    letter,
		done:      false,
		timestamp: time.Now(),
	}
}

func (jd *JobDict) doRun(runFunc func(letter *Letter) *errors.Error) *errors.Error {
	for _, job := range jd.dict {
		if job.done {
			continue
		}
		if err := runFunc(job.letter); err != nil {
			gLogger.Warn("job.dict.run letter failed, ignore it",
				zap.String("letter.hash", job.letter.Signature.Hash().S()))
			continue
		}
		job.done = true
		job.timestamp = time.Now()
	}
	return nil
}

type Line struct {
	inv       bid.BID
	arrows    [4]*JobDict
	timestamp time.Time
	lock      sync.Mutex
}

func NewLine(letter *Letter) *Line {
	ln := &Line{
		inv:    letter.Invariable,
		arrows: [4]*JobDict{NewJobDict(), NewJobDict(), NewJobDict(), NewJobDict()},
	}
	ln.doSet(letter)
	return ln
}

func (line *Line) RunOrRegister(letter *Letter, runFunc func(letter *Letter) *errors.Error) *errors.Error {
	idx := line.doGetIdx(letter.Arrow)
	if idx < 0 {
		return errors.Verify(ErrInvalidLetter)
	}
	line.lock.Lock()
	defer line.lock.Unlock()
	if line.arrows[idx] == nil {
		gLogger.Error("CODE ERROR", zap.Int("idx", idx))
		return nil
	}
	line.arrows[idx].doOncePut(letter)
	preOk := true
	for i := 0; i < len(line.arrows); i++ {
		if !preOk {
			return nil
		}
		if line.arrows[i] == nil {
			return nil
		}
		err := line.arrows[i].doRun(runFunc)
		if err != nil {
			return err
		}
		preOk = line.arrows[i].doHasDone()
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
	line.arrows[idx].doOncePut(letter)

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
