package zhanio

import (
	"net"
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
}

type conn struct {
	fd         int
	inBuf      []byte
	outBuf     []byte
	sa         syscall.Sockaddr
	status     ConnStatus
	action     Action
	ctx        interface{}
	localAddr  net.Addr
	remoteAddr net.Addr
	loop       *loop
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

func (c *conn) Write() {

}

func (c *conn) Read() {
	
}