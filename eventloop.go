package zhanio

import (
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
)

type loop struct {
	idx     int
	poll    *Poll
	packet  []byte
	fdconns map[int]*conn
	server  *server
	count   int32
}

func serve(eventHandler EventHandler, numLoops int, loadBalance LoadBalance, listeners []*listener) error {
	if numLoops <= 0 {
		if numLoops == 0 {
			numLoops = 1
		} else {
			numLoops = runtime.NumCPU()
		}
	}

	s := &server{}
	s.eventHandler = eventHandler
	s.lns = listeners
	s.cond = sync.NewCond(&sync.Mutex{})
	s.balance = loadBalance

	if s.eventHandler.Serving != nil {
		var svr Server
		svr.NumLoops = numLoops
		svr.Addrs = make([]net.Addr, len(listeners))
		for i, ln := range listeners {
			svr.Addrs[i] = ln.lnaddr
		}
		action := s.eventHandler.Serving()
		switch action {
		case None:
		case Shutdown:
			return nil
		}
	}

	defer func() {
		s.waitForShutDown()
		for _, l := range s.loops {
			l.poll.Trigger(0)
		}
		s.wg.Wait()
		for _, l := range s.loops {
			for _, c := range l.fdconns {
				l.loopCloseConn(c)
			}
			l.poll.ClosePoll()
		}
	}()

	for i := 0; i < numLoops; i++ {
		poll := OpenPoll()
		l := &loop{
			idx:     i,
			poll:    OpenPoll(),
			packet:  make([]byte, 0xFFFF),
			fdconns: make(map[int]*conn),
			server:  s,
		}
		for _, ln := range listeners {
			l.poll.AddRead(ln.fd)
		}
		s.loops = append(s.loops, l)
	}

	s.wg.Add(len(s.loops))
	for _, l := range s.loops {
		go l.loopRun()
	}

	return nil
}

func (lp *loop) loopRun() {
	defer func() {
		lp.server.signalShutdown()
		lp.server.wg.Done()
	}()

	lp.poll.Wait(func(fd int) error {
		if fd == 0 {
			return nil
		}
		c := lp.fdconns[fd]
		switch {
		case c == nil:
			return lp.loopAccept(fd)
		case c.status == Closed:
			return lp.loopOpen(c)
		case len(c.outBuf) > 0:
			return lp.loopWrite(c)
		case c.action != None:
			return lp.loopAction(c)
		default:
			return lp.loopRead(c)
		}
	})

}

func (lp *loop) loopAccept(fd int) error {
	for _, ln := range lp.server.lns {
		if ln.fd == fd {
			if len(lp.server.loops) > 1 {
				switch lp.server.balance {
				case LeastConnections:
					n := atomic.LoadInt32(&lp.count)
					for _, l := range lp.server.loops {
						if l.idx != lp.idx {
							if atomic.LoadInt32(&l.count) < n {
								return nil
							}
						}
					}
				case RoundRobin:
					idx := int(atomic.LoadUintptr(&lp.server.accepted)) % len(lp.server.loops)
					if idx != lp.idx {
						return nil
					}
					atomic.AddUintptr(&lp.server.accepted, 1)
				}
			}
			if ln.pconn != nil {
				return lp.loopUdpRead()
			}
			nfd, sa, err := syscall.Accept(fd)
			if err != nil {
				if err == syscall.EAGAIN {
					return nil
				}
				return err
			}
			if err := syscall.SetNonblock(nfd, true); err != nil {
				return err
			}
			c := &conn{fd: nfd, sa: sa, loop: lp}
			lp.fdconns[c.fd] = c
			lp.poll.AddReadWrite(c.fd)
			atomic.AddInt32(&lp.count, 1)
			break
		}
	}
	return nil
}

func (lp *loop) loopAction(c *conn) error {
	switch c.action {
	default:
		c.action = None
	case Close:
		lp.loopCloseConn(c)
	case Detach:
		lp.loopDetachConn(c)
	}
	if len(c.outBuf) == 0 && c.action == None {
		lp.poll.ModRead(c.fd)
	}
	return nil
}

func (lp *loop) loopOpen(c *conn) error {
	c.status = Opend
	if lp.server.eventHandler.Opened != nil {
		out, action := lp.server.eventHandler.Opened(c)
		if len(out) > 0 {
			c.outBuf = append([]byte{}, out...)
		}
		c.action = action
	}
	if len(c.outBuf) == 0 && c.action == None {
		lp.poll.ModRead(c.fd)
	}
	return nil
}

func (lp *loop) loopRead(c *conn) error {
	var in []byte
	n, err := syscall.Read(c.fd, lp.packet)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return lp.loopCloseConn(c)
	}
	in = lp.packet[:n]
	if lp.server.eventHandler.Data != nil {
		out, action := lp.server.eventHandler.Data(c, in)
		c.action = action
		if len(out) > 0 {
			c.outBuf = append([]byte{}, out...)
		}
	}
	if len(c.outBuf) != 0 || c.action != None {
		lp.poll.ModReadWrite(c.fd)
	}
	return nil
}

func (lp *loop) loopWrite(c *conn) error {
	n, err := syscall.Write(c.fd, c.outBuf)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return lp.loopCloseConn(c)
	}
	if n == len(c.outBuf) {
		c.outBuf = nil
	} else {
		c.outBuf = c.outBuf[n:]
	}
	if len(c.outBuf) == 0 && c.action == None {
		lp.poll.ModRead(c.fd)
	}
	return nil
}

func (lp *loop) loopCloseConn(c *conn) error {
	atomic.AddInt32(&lp.count, -1)
	delete(lp.fdconns, c.fd)
	syscall.Close(c.fd)
	if lp.server.eventHandler.Closed != nil {
		switch lp.server.eventHandler.Closed(c) {
		case None:
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func (lp *loop) loopDetachConn(c *conn) error {
	lp.poll.ModDetach(c.fd)
	atomic.AddInt32(&lp.count, -1)
	delete(lp.fdconns, c.fd)
	if err := syscall.SetNonblock(c.fd, false); err != nil {
		return err
	}
	switch lp.server.eventHandler.Detached() {
	case None:
	case Shutdown:
		return nil
	}
	return nil
}

func (lp *loop) loopUdpRead() error {
	return nil
}

func (lp *loop) loopWake() error {
	return nil
}
