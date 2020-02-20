package zhanio

import (
	"github.com/libp2p/go-reuseport"
	"net"
	"syscall"
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
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs); err != nil {
		return err
	}
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs)
}
