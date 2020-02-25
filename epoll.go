package zhanio

import (
	"syscall"
)

const (
	InitEventSum    = 128
	readEvents      = syscall.EPOLLPRI | syscall.EPOLLIN
	writeEvents     = syscall.EPOLLOUT
	readWriteEvents = readEvents | writeEvents
	EPOLLET         = 1 << 31
)

type Poll struct {
	fd     int
	wfd    int
	wfdBuf []byte
	queue  AsyncQueue
}

type EpollHandler func(int, uint32) error

func OpenPoll() (*Poll, error) {
	poll := new(Poll)
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	poll.fd = epollFD
	r0, _, errno := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if errno != 0 {
		syscall.Close(epollFD)
		return nil, errno
	}
	poll.wfd = int(r0)
	poll.wfdBuf = make([]byte, 8)
	if err = poll.AddRead(poll.wfd); err != nil {
		syscall.Close(poll.wfd)
		syscall.Close(poll.fd)
		return nil, err
	}
	return poll, nil
}

func (p *Poll) ClosePoll() error {
	if err := syscall.Close(p.fd); err != nil {
		return err
	}
	return syscall.Close(p.wfd)
}

func (p *Poll) Wait(handler EpollHandler) error {
	events := make([]syscall.EpollEvent, InitEventSum)
	var runJob bool = false
	for {
		n, err := syscall.EpollWait(p.fd, events, -1)
		if err != nil && err != syscall.EINTR {
			return err
		}
		for i := 0; i < n; i++ {
			if fd := int(events[i].Fd); fd != p.wfd {
				if err := handler(fd, events[i].Events); err != nil {
					return err
				}
			} else {
				if _, err := syscall.Read(p.wfd, p.wfdBuf); err != nil {
					return err
				}
				runJob = true
			}
		}
		if runJob {
			runJob = false
			if err := p.queue.ForEach(); err != nil {
			}
		}
		if n == len(events) {
			events = make([]syscall.EpollEvent, len(events)*2)
		}
	}
}

func (p *Poll) execute(fd int, op int, events uint32) error {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd,
		&syscall.EpollEvent{
			Events: events,
			Fd:     int32(fd),
		}); err != nil {
		return err
	}
	return nil
}

func (p *Poll) AddReadWrite(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_ADD, readWriteEvents)
}

func (p *Poll) AddRead(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_ADD, readEvents)
}

func (p *Poll) AddWrite(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_ADD, writeEvents)
}

func (p *Poll) ModReadWrite(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_MOD, readWriteEvents)
}

func (p *Poll) ModRead(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_MOD, readEvents)

}

func (p *Poll) ModWrite(fd int) error {
	return p.execute(fd, syscall.EPOLL_CTL_MOD, writeEvents)
}

func (p *Poll) Mod(fd int, event uint32) error {
	return p.execute(fd, syscall.EPOLL_CTL_MOD, event)
}

func (p *Poll) Add(fd int, event uint32) error {
	return p.execute(fd, syscall.EPOLL_CTL_ADD, event)
}

func (p *Poll) Execute(fd int, op int, event uint32) error {
	return p.execute(fd, syscall.EPOLL_CTL_ADD, event)
}

func (p *Poll) Delete(fd int) error {
	if err := syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil); err != nil {
		return err
	}
	return nil
}

func (p *Poll) Trigger(job func() error) error {
	p.queue.locker.Lock()
	p.queue.jobs = append(p.queue.jobs, job)
	p.queue.locker.Unlock()
	_, err := syscall.Write(p.wfd, []byte{0, 0, 0, 0, 0, 0, 0, 1})
	if err != nil {
		return err
	}
	return nil
}
