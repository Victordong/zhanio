package zhanio

import (
	"golang.org/x/sys/unix"
	"math"
	"math/rand"
	"net"
	"sync/atomic"
	"syscall"
	"time"
)

type loop struct {
	idx     int
	poll    *Poll
	fdconns map[int]*conn
	server  *server
	count   int64
	packet  []byte
	codec   Codec
}

func (lp *loop) subReactor() {
	defer func() {
		lp.server.signalShutdown()
	}()

	handler := func(fd int, event uint32) error {
		if c, ok := lp.fdconns[fd]; ok {
			switch {
			case c.action != None:
				return lp.loopAction(c)
			case c.outBuf.Length() > 0 && event&writeEvents != 0:
				return lp.loopWrite(c)
			case event&readEvents != 0:
				return lp.loopRead(c)
			}
		}
		return nil
	}

	lp.poll.Wait(handler)
}

func (lp *loop) mainReactor() {
	defer func() {
		lp.server.signalShutdown()
	}()
	handler := func(fd int, event uint32) error {
		if event&readEvents != 0 {
			return lp.loopAccept(fd)
		}
		return nil
	}
	lp.poll.Wait(handler)
}

func (lp *loop) tickReactor() {
	defer func() {
		lp.server.signalShutdown()
	}()
	handler := func(fd int, event uint32) error {
		return nil
	}
	lp.poll.Wait(handler)
}

func (lp *loop) loopAccept(fd int) error {
	ln := lp.server.ln
	if ln.fd == fd {

		var cur *loop
		switch lp.server.balance {
		case Random:
			idx := rand.Intn(len(lp.server.loops))
			cur = lp.server.loops[idx]
		case LeastConnections:
			n := int64(math.MaxInt64)
			for _, l := range lp.server.loops {
				if atomic.LoadInt64(&l.count) < n {
					cur = l
				}
			}
		case RoundRobin:
			idx := int(atomic.LoadUint64(&lp.server.accepted)) % len(lp.server.loops)
			cur = lp.server.loops[idx]
			atomic.AddUint64(&lp.server.accepted, 1)
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

		if cur != nil {
			if err := cur.poll.AddRead(nfd); err == nil {
				c := &conn{fd: nfd, sa: sa, loop: cur, inBuf: NewBuffer(InitSize), outBuf: NewBuffer(InitSize)}
				cur.fdconns[c.fd] = c
				atomic.AddInt64(&cur.count, 1)
				return cur.loopOpen(c)
			}
		}
	}
	return nil
}

func (lp *loop) loopAction(c *conn) error {
	switch c.action {
	default:
		c.action = None
		return nil
	case Close:
		c.action = None
		return lp.loopCloseConn(c)
	case Shutdown:
		c.action = None
		lp.loopWrite(c)
		return errServerShutdown
	}
	return nil
}

func (lp *loop) loopOpen(c *conn) error {
	c.status = Opend
	c.localAddr = lp.server.ln.lnaddr
	c.remoteAddr = SockaddrToTCPOrUnixAddr(c.sa)

	if lp.server.opts.KeepAlive > 0 {
		if _, ok := lp.server.ln.ln.(*net.TCPListener); ok {
			err := SetKeepAlive(c.fd, lp.server.opts.KeepAlive)
			return err
		}
	}

	out, action := lp.server.eventHandler.Opened(c)
	c.action = action
	if len(out) > 0 {
		c.outBuf.Write(out)
	}

	if c.outBuf.Length() > 0 {
		err := lp.poll.ModWrite(c.fd)
		return err
	}

	return lp.loopAction(c)
}

func (lp *loop) loopRead(c *conn) (err error) {
	n, err := unix.Read(c.fd, lp.packet)
	if n == 0 || err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return lp.loopCloseConn(c)
	}
	c.inBuf.Write(lp.packet[:n])
	for inFrame, _ := c.read(); inFrame != nil; inFrame, _ = c.read() {
		go lp.server.eventHandler.Data(c, inFrame)
		if err := lp.loopAction(c); err != nil {
			return err
		}
		if c.status != Opend {
			return nil
		}
	}

	return nil
}

func (lp *loop) loopWrite(c *conn) error {
	begin, end := c.outBuf.ReadRaw()

	n, err := syscall.Write(c.fd, begin)
	if err != nil {
		if err == syscall.EAGAIN {
			return nil
		}
		return lp.loopCloseConn(c)
	}
	c.outBuf.ClearN(n)

	if n == len(begin) && end != nil {
		n, err := syscall.Write(c.fd, end)
		if err != nil {
			if err == syscall.EAGAIN {
				return nil
			}
			return lp.loopCloseConn(c)
		}
		c.outBuf.ClearN(n)
	}

	if c.outBuf.IsEmpty() {
		lp.poll.ModRead(c.fd)
	}
	return lp.loopAction(c)
}

func (lp *loop) loopCloseConn(c *conn) error {
	if lp.poll.Delete(c.fd) == nil && syscall.Close(c.fd) == nil {
		delete(lp.fdconns, c.fd)
		atomic.AddInt64(&lp.count, -1)
		switch lp.server.eventHandler.Closed(c) {
		case None:
		case Shutdown:
			return errClosing
		}
	}
	return nil
}

func (lp *loop) loopTicker() {
	for {
		lp.poll.Trigger(func() (err error) {
			delay, action := lp.server.eventHandler.Tick()
			lp.server.ticktock <- delay
			switch action {
			case None:
			case Shutdown:
				err = errServerShutdown
			}
			return nil
		})
		select {
		case delay := <-lp.server.ticktock:
			time.Sleep(delay)
		}
	}
}
