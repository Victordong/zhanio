package zhanio

import (
	"net"
	"sync"
	"syscall"
)

type ConnStatus int

const (
	Closed ConnStatus = iota
	Opend
)

type Conn interface {
	Context() interface{}
	SetContext(interface{})
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Read() (buf []byte)
	AsyncWrite(buf []byte)
	Close()
}

type conn struct {
	fd         int
	inBuf      *RingBuffer
	outBuf     *RingBuffer
	sa         syscall.Sockaddr
	status     ConnStatus
	action     Action
	ctx        interface{}
	localAddr  net.Addr
	remoteAddr net.Addr
	loop       *loop
	readLock   sync.Mutex
	writeLock  sync.Mutex
}

func (c *conn) write(buf []byte) error {
	var err error
	if !c.outBuf.IsEmpty() {
		c.writeLock.Lock()
		c.outBuf.Write(buf)
		c.writeLock.Unlock()
		return nil
	}
	n, err := syscall.Write(c.fd, buf)
	if err != nil {
	}
	if err != nil {
		if err == syscall.EAGAIN {
			c.writeLock.Lock()
			c.outBuf.Write(buf)
			c.writeLock.Unlock()
			err = c.loop.poll.ModReadWrite(c.fd)
			return err
		} else {
			return c.loop.loopCloseConn(c)
		}
	}
	if n != len(buf) {
		c.writeLock.Lock()
		c.outBuf.Write(buf[n:])
		c.writeLock.Unlock()
		err = c.loop.poll.ModReadWrite(c.fd)
		return err
	}
	return nil
}

func (c *conn) read() ([]byte, error) {
	return c.loop.server.codec.Decode(c)
}

func (c *conn) Context() interface{} {
	return c.ctx
}

func (c *conn) SetContext(ctx interface{}) {
	c.ctx = ctx
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) AsyncWrite(buf []byte) {
	if writeResult, err := c.loop.server.codec.Encode(c, buf); err == nil {
		c.loop.poll.Trigger(func() error {
			if c.status == Opend {
				err := c.write(writeResult)
				return err
			}
			return errCloseConns
		})
	}
}

func (c *conn) Read() []byte {
	if c.inBuf.isEmpty {
		return nil
	}
	c.readLock.Lock()
	head, tail := c.inBuf.ReadRaw()
	c.readLock.Unlock()
	result := make([]byte, c.inBuf.size)
	if head != nil {
		copy(result, head)
	}
	if tail != nil {
		copy(result[len(head):], tail)
	}
	c.inBuf.Reset()
	return result
}

func (c *conn) Close() {
	c.loop.poll.Trigger(func() error {
		return c.loop.loopCloseConn(c)
	})
}
