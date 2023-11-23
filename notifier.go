package twelve

import (
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
)

const (
	NotifierMsgBufSize = 102400
)

type IListener interface {
	OnRequest(msg *Letter)
	OnPrepare(msg *Letter)
	OnCommitted(msg *Letter)
	OnConfirmed(msg *Letter)
}

type Notifier struct {
	node     ITwelveNode
	listener IListener
	buf      chan *Letter
}

func NewNotifier(node ITwelveNode) *Notifier {
	if node == nil {
		sys.Error("require node")
	}
	ntf := &Notifier{
		node: node,
		buf:  make(chan *Letter, NotifierMsgBufSize),
	}
	node.Register(ntf)
	return ntf
}

func (n *Notifier) BindListener(listener IListener) {
	n.listener = listener
}

func (n *Notifier) On(letter *Letter) *errors.Error {
	if letter == nil {
		return errors.Verify("require letter, it is nil")
	}
	select {
	case n.buf <- letter:
		return nil
	default:
		return errors.Sys("The buffer is full")
	}
}

func (n *Notifier) Notify(letter *Letter) *errors.Error {
	return n.node.Notify(letter)
}

func (n *Notifier) Listening() {
	go func() {
		for {
			sys.Info("========>>>>> Listening.....")
			letter := <-n.buf
			if sys.RunMode.IsRd() {
				gLogger.Info("Notifier.Listening", zap.Any("letter", letter))
			}
			switch letter.Arrow {
			case RequestArrow:
				sys.Info("========>>>>> Listening.OnRequest.1....")
				n.listener.OnRequest(letter)
				sys.Info("========>>>>> Listening.OnRequest.2....")
			case PrepareArrow:
				sys.Info("========>>>>> Listening.OnPrepare.1....")
				n.listener.OnPrepare(letter)
				sys.Info("========>>>>> Listening.OnPrepare.2....")
			case CommittedArrow:
				sys.Info("========>>>>> Listening.OnCommitted.1....")
				n.listener.OnCommitted(letter)
				sys.Info("========>>>>> Listening.OnCommitted.2....")
			case ConfirmedArrow:
				sys.Info("========>>>>> Listening.OnConfirmed.1....")
				n.listener.OnConfirmed(letter)
				sys.Info("========>>>>> Listening.OnConfirmed.2....")
			}
		}
	}()
}
