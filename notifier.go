package twelve

import (
	"github.com/hootuu/utils/errors"
	"github.com/hootuu/utils/sys"
	"go.uber.org/zap"
)

const (
	NotifierMsgBufSize = 1024
)

type IListener interface {
	OnRequest(msg *Message)
	OnPrepare(msg *Message)
	OnCommitted(msg *Message)
	OnConfirmed(msg *Message)
}

type Notifier struct {
	node     ITwelveNode
	listener IListener
	buf      chan *Message
}

func NewNotifier(node ITwelveNode) *Notifier {
	if node == nil {
		sys.Error("require node")
	}
	ntf := &Notifier{
		node: node,
		buf:  make(chan *Message, NotifierMsgBufSize),
	}
	node.Register(ntf)
	return ntf
}

func (n *Notifier) BindListener(listener IListener) {
	n.listener = listener
}

func (n *Notifier) On(msg *Message) *errors.Error {
	if sys.RunMode.IsRd() {
		gLogger.Info("Notifier.On", zap.Any("msg", msg))
	}
	if msg == nil {
		return errors.Verify("require message, it is nil")
	}
	select {
	case n.buf <- msg:
		return nil
	default:
		return errors.Sys("The buffer is full")
	}
}

func (n *Notifier) Notify(msg *Message) *errors.Error {
	return n.node.Notify(msg)
}

func (n *Notifier) Listening() {
	go func() {
		for {
			sys.Info("========>>>>> Listening.....")
			msg := <-n.buf
			if sys.RunMode.IsRd() {
				gLogger.Info("Notifier.Listening", zap.Any("msg", msg))
			}
			switch msg.Type {
			case RequestMessage:
				sys.Info("========>>>>> Listening.OnRequest.1....")
				n.listener.OnRequest(msg)
				sys.Info("========>>>>> Listening.OnRequest.2....")
			case PrepareMessage:
				sys.Info("========>>>>> Listening.OnPrepare.1....")
				n.listener.OnPrepare(msg)
				sys.Info("========>>>>> Listening.OnPrepare.2....")
			case CommittedMessage:
				sys.Info("========>>>>> Listening.OnCommitted.1....")
				n.listener.OnCommitted(msg)
				sys.Info("========>>>>> Listening.OnCommitted.2....")
			case ConfirmedMessage:
				sys.Info("========>>>>> Listening.OnConfirmed.1....")
				n.listener.OnConfirmed(msg)
				sys.Info("========>>>>> Listening.OnConfirmed.2....")
			}
		}
	}()
}
