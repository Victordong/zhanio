package main

import (
	"fmt"
	"github.com/Victordong/zhanio"
	"syscall"
	"time"
)

func handle(fd int, events uint32) error {
	if events&syscall.EPOLLIN != 0 {
		result := make([]byte, 12)
		n, err := syscall.Read(fd, result)
		fmt.Println(result)
		if err != nil || n == 0 {
			return err
		}
	}
	return nil
}

func main() {
	poll, err := zhanio.OpenPoll()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("poll open")
	poll.Add(syscall.Stdin, syscall.EPOLLET)
	go func() {
		time.Sleep(time.Second * 10)
		poll.Trigger(func() error {
			fmt.Println("trigger beautiful")
			return nil
		})
	}()

	err = poll.Wait(handle)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("poll end")
}
