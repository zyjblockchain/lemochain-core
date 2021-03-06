package p2p

import (
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"net"
	"sync/atomic"
	"time"
)

const (
	dialTimeout = 3 * time.Second
)

type HandleConnFunc func(fd net.Conn, nodeID *NodeID) error

type IDialManager interface {
	Start() error
	Stop() error
	runDialTask(node string) int
}

type DialManager struct {
	handleConn HandleConnFunc
	discover   *DiscoverManager
	state      int32
}

func NewDialManager(handleConn HandleConnFunc, discover *DiscoverManager) *DialManager {
	return &DialManager{
		handleConn: handleConn,
		discover:   discover,
		state:      0,
	}
}

// Start
func (m *DialManager) Start() error {
	if atomic.LoadInt32(&m.state) == 1 {
		log.Info("Dial manager has already started")
		return ErrHasStared
	}
	atomic.StoreInt32(&m.state, 1)
	go m.loop()
	log.Infof("Dial manager start")
	return nil
}

// Stop
func (m *DialManager) Stop() error {
	if atomic.LoadInt32(&m.state) < 1 {
		log.Info("Dial manager not start")
		return ErrNotStart
	}
	atomic.StoreInt32(&m.state, -1)
	log.Infof("Dial manager stop")
	return nil
}

// runDialTask Run dial task
func (m *DialManager) runDialTask(node string) int {
	// check
	nodeID, endpoint := ParseNodeString(node)
	if nodeID == nil {
		log.Warnf("Dial: invalid node. node: %s", node)
		return -1
	}
	// black node
	if m.discover.IsBlackNode(nodeID) {
		log.Warnf("Dial: this node is black node. node: %s", node)
		return -1
	}
	// dial
	conn, err := net.DialTimeout("tcp", endpoint, dialTimeout)
	if err != nil {
		log.Warnf("Dial node error: %s", err.Error())
		if err = m.discover.SetConnectResult(nodeID, false); err != nil {
			log.Errorf("SetConnectResult failed: %v", err)
		}
		return -1
	}
	// handle connection
	if err = m.handleConn(conn, nodeID); err != nil {
		log.Warnf("Node first connect error: %s", err.Error())
		return -1
	}
	return 0
}

// loop
func (m *DialManager) loop() {
	for {
		list := m.discover.connectingNodes()
		for _, n := range list {
			log.Debugf("Start dial: %s", n[:16])
			if atomic.LoadInt32(&m.state) == -1 {
				return
			}
			m.runDialTask(n)
		}
		if atomic.LoadInt32(&m.state) == -1 {
			return
		}
		time.Sleep(3 * time.Second)
	}
}
