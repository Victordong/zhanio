package zhanio

import (
	"github.com/libp2p/go-reuseport"
	"golang.org/x/sys/unix"
	"net"
)

type Options struct {
	ReusePort   bool
	NumLoops    int
	LoadBalance LoadBalance
	KeepAlive   int
}

// ReusePortListen returns a net.Listener for TCP.
func ReusePortListen(proto, addr string) (net.Listener, error) {
	return reuseport.Listen(proto, addr)
}

func SetKeepAlive(fd, secs int) error {
	if err := unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, secs)
}
