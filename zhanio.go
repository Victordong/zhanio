package zhanio

import (
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const connRingBufferSize = 1024

type LoadBalance int

const (
	Random LoadBalance = iota
	RoundRobin
	LeastConnections
)

type Action int

const (
	// None indicates that no action should occur following an event.
	None Action = iota

	// Close closes the connection.
	Close

	// Shutdown shutdowns the server.
	Shutdown
)

var errClosing = errors.New("closing")
var errCloseConns = errors.New("close conns")

type EventHandler interface {
	Serving(s Server) (action Action)
	Opened(c Conn) (out []byte, action Action)
	Closed(c Conn) (action Action)
	Data(c Conn, frame []byte)
	Tick() (delay time.Duration, action Action)
}

type server struct {
	eventHandler EventHandler
	mainLoop     *loop
	loops        []*loop
	trigger      *loop
	ln           *listener
	wg           sync.WaitGroup
	opts         Options
	cond         *sync.Cond
	balance      LoadBalance
	codec        Codec
	tch          chan time.Duration
	ticktock     chan time.Duration
	accepted     uint64
}

type Server struct {
	NumLoops int
	Addr     net.Addr
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

func Serve(eventHandler EventHandler, addr string, opts Options) error {
	ln := new(listener)
	defer func() {
		ln.close()
	}()

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
	return serve(eventHandler, ln, opts)
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
