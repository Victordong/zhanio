package zhanio

import (
	"golang.org/x/sys/unix"
	"syscall"
)

type Poll struct {
	fd     int
	wfd    int
	wfdBuf []byte
	queue  AsyncQueue
}

func OpenPoll() (*Poll, error) {
	poll := new(Poll)
	epollFD, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	poll.fd = epollFD
	r0, _, errno := unix.Syscall(unix.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		unix.Close(epollFD)
		return nil, errno
	}
	poll.wfd = int(r0)
	poll.wfdBuf = make([]byte, 8)
	if err = poll.AddRead(poll.fd); err != nil {
		unix.Close(poll.wfd)
		unix.Close(poll.fd)
		return nil, err
	}
	return poll, nil
}

func (p *Poll) ClosePoll() error {
	if err := unix.Close(p.fd); err != nil {
		return err
	}
	return unix.Close(p.wfd)
}

func (p *Poll) Wait(handler func(int, func() error) error) error {
	events := make([]unix.EpollEvent, 0)
	var note bool
	for {
		n, err := unix.EpollWait(p.fd, events, -1)
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			if fd := int(events[i].Fd); fd != p.wfd {
				if err := handler(fd, nil); err != nil {
					return err
				}
			} else {
				if _, err := unix.Read(p.wfd, p.wfdBuf); err != nil {
					return err
				}
				note = true
			}
			if note {
				note = false
				if err := p.queue.ForEach(func(job func() error) error {
					return handler(0, job)
				}); err != nil {
					return err
				}
			}
		}
	}
}

func (p *Poll) execute(fd int, op int, events uint32) error {
	if err := unix.EpollCtl(p.fd, unix.EPOLL_CTL_ADD, fd,
		&unix.EpollEvent{
			Events: events,
			Fd:     int32(fd),
		}); err != nil {
		return err
	}
	return nil
}

func (p *Poll) AddReadWrite(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_ADD, unix.EPOLLIN|unix.EPOLLOUT)
}

func (p *Poll) AddRead(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_ADD, unix.EPOLLIN)
}

func (p *Poll) AddWrite(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_ADD, unix.EPOLLOUT)
}

func (p *Poll) ModReadWrite(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_MOD, unix.EPOLLIN|unix.EPOLLOUT)
}

func (p *Poll) ModRead(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_MOD, unix.EPOLLIN)

}

func (p *Poll) ModWrite(fd int) error {
	return p.execute(fd, unix.EPOLL_CTL_MOD, unix.EPOLLOUT)
}

func (p *Poll) Delete(fd int) error {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}
	return nil
}
