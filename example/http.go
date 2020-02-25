package main

import (
	"flag"
	"fmt"
	"github.com/Victordong/zhanio"
	"time"
)

type handler struct {
	tick time.Duration
}

func (h *handler) Serving(s zhanio.Server) (action zhanio.Action) {
	fmt.Println("server", s.Addr.String())
	return zhanio.None
}

func (h *handler) Opened(c zhanio.Conn) (out []byte, action zhanio.Action) {
	return nil, zhanio.None
}

func (h *handler) Closed(c zhanio.Conn) (action zhanio.Action) {
	return zhanio.None
}
func (h *handler) Data(c zhanio.Conn, frame []byte) {
	fmt.Println(frame)
}

func (h *handler) Tick() (delay time.Duration, action zhanio.Action) {
	fmt.Println("this is a tick")
	return h.tick, zhanio.None
}

func main() {
	var port int
	var tick time.Duration
	flag.IntVar(&port, "port", 8080, "server port")
	flag.DurationVar(&tick, "tick", time.Second*5, "pushing tick")
	flag.Parse()
	h := &handler{tick: tick}
	opts := zhanio.Options{
		NumLoops:    2,
		LoadBalance: zhanio.RoundRobin,
	}
	err := zhanio.Serve(h, fmt.Sprintf("tcp://:%d", port), opts)
	if err != nil {
		fmt.Println(err)
	}
}
