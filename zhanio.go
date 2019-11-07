package zhanio

import (
	"net"
	"os"
	"strings"
	"sync"
	"errors"
)

type LoadBalance int

const (
	Random LoadBalance = iota
	RoundRobin
	LeastConnections
)

type Action int

const (
	None Action = iota
	Detach
	Close
	Shutdown
)

var errClosing = errors.New("closing")
var errCloseConns = errors.New("close conns")

type EventHandler struct {
	Serving func() (action Action)
	Opened  func(c Conn) (out []byte, action Action)
	Closed  func(c Conn) (action Action)
	Data    func(c  Conn, in []byte) (out []byte, action Action)
	Tick    func() (action Action)
	Detached func() (action Action)
}

type addrOpts struct {
	reusePort bool
}

type server struct {
	eventHandler EventHandler
	loops        []*loop
	lns          []*listener
	wg           sync.WaitGroup
	cond         *sync.Cond
	balance      LoadBalance
	accepted     uintptr
}

type Server struct {
	Addrs    []net.Addr
	NumLoops int
}

func parseAddr(addr string) (network, address string) {
	network = "tcp"
	address = addr
	if strings.Contains(address, "://") {
		network = strings.Split(address, "://")[0]
		address = strings.Split(address, "://")[1]
	}
	if strings.HasSuffix(network, "-net") {
		network = network[:len(network)-4]
	}
	q := strings.Index(address, "?")
	if q != -1 {
		address = address[:q]
	}
	return
}

func Serve(eventHandler EventHandler, numLoops int, loadBalance LoadBalance, addrs ...string) error {
	var lns []*listener
	defer func() {
		for _, ln := range lns {
			ln.close()
		}
	}()
	for _, addr := range addrs {
		var ln listener
		ln.network, ln.addr = parseAddr(addr)
		if ln.network == "unix" {
			os.RemoveAll(ln.addr)
		}
		var err error
		if ln.network == "udp" {
			ln.pconn, err = net.ListenPacket(ln.network, ln.addr)
		} else {
			ln.ln, err = net.Listen(ln.network, ln.addr)
		}
		if err != nil {
			return err
		}
		if ln.pconn != nil {
			ln.lnaddr = ln.pconn.LocalAddr()
		} else {
			ln.lnaddr = ln.ln.Addr()
		}
		if err := ln.system(); err != nil {
			return err
		}
		lns = append(lns, &ln)
	}
	return serve(eventHandler, numLoops, loadBalance, lns)
}

func (s *server) signalShutdown() {
	s.cond.L.Lock()
	s.cond.Signal()
	s.cond.L.Unlock()
}

func (s *server) waitForShutDown() {
	s.cond.L.Lock()
	s.cond.Wait()
	s.cond.L.Unlock()
}
