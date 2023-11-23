package twelve

import (
	"github.com/hootuu/tome/nd"
	"github.com/hootuu/utils/errors"
	"go.uber.org/zap"
)

type ITwelveListener interface {
	On(letter *Letter) *errors.Error
}

type ITwelveNode interface {
	Register(listener ITwelveListener)
	Notify(letter *Letter) *errors.Error
	Node() *nd.Node
}

type MemTwelveListenerBus struct {
	listeners []ITwelveListener
}

type MemTwelveNode struct {
	bus  *MemTwelveListenerBus
	node *nd.Node
}

func NewMemTwelveNode(bus *MemTwelveListenerBus, node *nd.Node) *MemTwelveNode {
	return &MemTwelveNode{
		bus:  bus,
		node: node,
	}
}

func (m *MemTwelveNode) Register(listener ITwelveListener) {
	m.bus.listeners = append(m.bus.listeners, listener)
}

func (m *MemTwelveNode) Notify(letter *Letter) *errors.Error {
	if len(m.bus.listeners) == 0 {
		return nil
	}
	//if sys.RunMode.IsRd() {
	//	gLogger.Info("notify msg: ", zap.Any("msg", msg))
	//}
	for _, listener := range m.bus.listeners {
		err := listener.On(letter)
		if err != nil {
			gLogger.Error("listener.On(msg) failed", zap.Error(err), zap.Any("letter", letter))
		}
	}
	return nil
}

func (m *MemTwelveNode) Node() *nd.Node {
	return m.node
}
