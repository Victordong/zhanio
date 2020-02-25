package zhanio

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func (s *server) start(loopNum int) error {
	for i := 0; i < loopNum; i++ {
		if p, err := OpenPoll(); err == nil {
			l := &loop{
				idx:     i,
				poll:    p,
				fdconns: make(map[int]*conn),
				server:  s,
				count:   0,
				packet:  make([]byte, 0x10000),
				codec:   s.codec,
			}
			s.loops = append(s.loops, l)
		} else {
			return err
		}
	}
	for i := 0; i < loopNum; i++ {
		s.wg.Add(1)
		go func(index int) {
			defer s.wg.Done()
			fmt.Println(s.loops[index].idx, s.ln.fd, s.loops[index].poll.fd)
			s.loops[index].subReactor()
		}(i)
	}
	if p, err := OpenPoll(); err == nil {
		mainLoop := &loop{
			idx:    -1,
			poll:   p,
			server: s,
		}
		err = mainLoop.poll.ModRead(s.ln.fd)
		s.mainLoop = mainLoop
		s.wg.Add(1)
		go func() {
			defer func() {
				s.wg.Done()
			}()
			fmt.Println(mainLoop.idx, s.ln.fd, s.mainLoop.poll.fd, s.mainLoop.poll.wfd)
			mainLoop.mainReactor()
		}()
		if s.opts.Tick {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				mainLoop.loopTicker()
			}()
		}
	} else {
		fmt.Println(err)
		return err
	}
	return nil
}

func (s *server) stop() {
	s.waitForShutDown()
	for _, loop := range s.loops {
		loop.poll.Trigger(func() error {
			return errServerShutdown
		})
	}
	s.ln.close()
	s.mainLoop.poll.Trigger(func() error {
		return errServerShutdown
	})
	s.wg.Wait()
	for _, loop := range s.loops {
		for _, conn := range loop.fdconns {
			loop.loopCloseConn(conn)
		}
	}
	if s.mainLoop != nil {
		s.mainLoop.poll.ClosePoll()
	}
}

func (s *server) closeLoops() {
	for _, loop := range s.loops {
		loop.poll.ClosePoll()
	}
}

func serve(eventHandler EventHandler, ln *listener, opts Options) error {
	s := new(server)
	s.ln = ln
	s.eventHandler = eventHandler
	s.opts = opts
	s.cond = sync.NewCond(&sync.Mutex{})
	s.tch = make(chan time.Duration)
	if opts.Codec != nil {
		s.codec = opts.Codec
	} else {
		s.codec = &defaultCodec{}
	}

	server := Server{NumLoops: opts.NumLoops, Addr: ln.lnaddr}
	action := s.eventHandler.Serving(server)
	switch action {
	case None:
	case Shutdown:
		return nil
	}
	if err := s.start(opts.NumLoops); err != nil {
		s.closeLoops()
		log.Printf("gnet server is stoping with error: %v\n", err)
		return err
	}
	defer s.stop()
	return nil
}
