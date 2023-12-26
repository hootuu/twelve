package twelve

import (
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
)

const (
	NotifierMsgBufSize = 102400
)

type IListener interface {
	OnRequest(msg *Letter) *errors.Error
	OnPrepare(msg *Letter) *errors.Error
	OnCommitted(msg *Letter) *errors.Error
	OnConfirmed(msg *Letter) *errors.Error
	OnInvariable(msg *Letter) *errors.Error
}

type Notifier struct {
	node     ITwelveNode
	listener IListener
	buf      chan *Letter
	lf       *LineFactory
}

func NewNotifier(node ITwelveNode) *Notifier {
	if node == nil {
		sys.Error("require node")
	}
	ntf := &Notifier{
		node: node,
		buf:  make(chan *Letter, NotifierMsgBufSize),
		lf:   NewLineFactory(),
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
	if sys.RunMode.IsRd() {
		gLogger.Info("REV LETTER", zap.String("peer", letter.From.S()),
			zap.String("here", nd.Here().ID.S()))
	}
	//if letter.From.IsHere() {
	//	return nil
	//}
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
			if letter.Arrow == InvariableArrow {
				sys.Info("========>>>>> Listening.OnInvariable.1....")
				err := n.listener.OnInvariable(letter)
				sys.Info("========>>>>> Listening.OnInvariable.2....")
				if err != nil {
					gLogger.Error("Listening.OnInvariable error", zap.Error(err))
					sys.Error("Listening.OnInvariable error", err.Error())
				}
				continue
			}
			line := n.lf.MustGet(letter)
			err := line.RunOrRegister(letter, func(letter *Letter) *errors.Error {
				var innerErr *errors.Error
				switch letter.Arrow {
				case RequestArrow:
					sys.Info("========>>>>> Listening.OnRequest.1....")
					innerErr = n.listener.OnRequest(letter)
					sys.Info("========>>>>> Listening.OnRequest.2....")
				case PrepareArrow:
					sys.Info("========>>>>> Listening.OnPrepare.1....")
					innerErr = n.listener.OnPrepare(letter)
					sys.Info("========>>>>> Listening.OnPrepare.2....")
				case CommittedArrow:
					sys.Info("========>>>>> Listening.OnCommitted.1....")
					innerErr = n.listener.OnCommitted(letter)
					sys.Info("========>>>>> Listening.OnCommitted.2....")
				case ConfirmedArrow:
					sys.Info("========>>>>> Listening.OnConfirmed.1....")
					innerErr = n.listener.OnConfirmed(letter)
					sys.Info("========>>>>> Listening.OnConfirmed.2....")
				}
				if innerErr != nil {
					return innerErr
				}
				return nil
			})
			if err != nil {
				gLogger.Error("line.RunOrRegister error", zap.Error(err))
				sys.Error("line.RunOrRegister error", err.Error())
				//n.buf <- letter
				continue
			}
		}
	}()
}
