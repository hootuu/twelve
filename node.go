package twelve

import (
	"github.com/hootuu/tome/pr"
	"github.com/hootuu/utils/errors"
	"go.uber.org/zap"
)

type ITwelveListener interface {
	On(msg *Message) *errors.Error
}

type ITwelveNode interface {
	Register(listener ITwelveListener)
	Notify(msg *Message) *errors.Error
	Peer() *pr.Local
}

type MemTwelveListenerBus struct {
	listeners []ITwelveListener
}

type MemTwelveNode struct {
	bus  *MemTwelveListenerBus
	peer *pr.Local
}

func NewMemTwelveNode(bus *MemTwelveListenerBus, peer *pr.Local) *MemTwelveNode {
	return &MemTwelveNode{
		bus:  bus,
		peer: peer,
	}
}

func (m *MemTwelveNode) Register(listener ITwelveListener) {
	m.bus.listeners = append(m.bus.listeners, listener)
}

func (m *MemTwelveNode) Notify(msg *Message) *errors.Error {
	if len(m.bus.listeners) == 0 {
		return nil
	}
	//if sys.RunMode.IsRd() {
	//	gLogger.Info("notify msg: ", zap.Any("msg", msg))
	//}
	for _, listener := range m.bus.listeners {
		err := listener.On(msg)
		if err != nil {
			gLogger.Error("listener.On(msg) failed", zap.Error(err), zap.Any("msg", msg))
		}
	}
	return nil
}

func (m *MemTwelveNode) Peer() *pr.Local {
	return m.peer
}
