package zhanio

import "syscall"

type Poll struct {
	fd  int
	wfd int
}

func OpenPoll() *Poll {
	l := new(Poll)
	p, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(err)
	}
	l.fd = p
	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		syscall.Close(p)
		panic(err)
	}
	l.wfd = int(r0)
	l.AddRead(l.wfd)
	return l
}

func (p *Poll) ClosePoll() error {
	if err := syscall.Close(p.wfd); err != nil {
		return err
	}
	return syscall.Close(p.fd)
}

func (p *Poll) AddRead(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{
			Fd:     int32(fd),
			Events: syscall.EPOLLIN,
		}); err != nil {
		panic(err)
	}
}

func (p *Poll) AddWrite(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{
			Events: syscall.EPOLLOUT,
			Fd:     int32(fd),
		}); err != nil {
		panic(err)
	}

}

func (p *Poll) AddReadWrite(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{
			Events: syscall.EPOLLOUT | syscall.EPOLLIN,
			Fd:     int32(fd),
		}); err != nil {
		panic(err)
	}
}

func (p *Poll) ModRead(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd,
		&syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(fd),
		}); err != nil {
			panic(err)
	}
}

func (p *Poll) ModReadWrite(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd,
		&syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
			Fd:     int32(fd),
		}); err != nil {
		panic(err)
	}
}

func (p *Poll) ModDetach(fd int) {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd,
		&syscall.EpollEvent{
			Events: syscall.EPOLLIN | syscall.EPOLLOUT,
			Fd:     int32(fd),
		}); err != nil {
		panic(err)
	}
}

func (p *Poll) Wait(iter func(fd int) error) error{
	events := make([]syscall.EpollEvent, 64)
	for {
		n, err := syscall.EpollWait(p.fd, events, -1)
		if err != nil && err != syscall.EINTR {
			return err
		}
		for i:=0;i<n;i++ {
			if fd:= int(events[i].Fd);fd!=p.wfd {
				if err := iter(fd); err != nil {}
				return err
			}
		}
	}
}

func (p *Poll) Trigger(note interface{}) error {
	_, err := syscall.Write(p.wfd, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	return err
}