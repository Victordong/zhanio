package zhanio

import (
	"fmt"
	"testing"
)

func TestRingBuffer_ReadRaw(t *testing.T) {
	buffer := NewBuffer(10)
	for i := 0; i < 10; i++ {
		buffer.Write([]byte{byte(i)})
	}
	fmt.Println(buffer.rPos, buffer.wPos)
	head, tail := buffer.ReadRaw()
	fmt.Println(head)
	fmt.Println(tail)
	buffer.ClearN(5)
	fmt.Println(buffer.rPos, buffer.wPos)
	head, tail = buffer.ReadRaw()
	fmt.Println(head)
	fmt.Println(tail)
	for i := 0; i < 3; i++ {
		buffer.Write([]byte{byte(i)})
	}
	fmt.Println(buffer.rPos, buffer.wPos)
	head, tail = buffer.ReadRaw()
	fmt.Println(head)
	fmt.Println(tail)
	for i := 0; i < 4; i++ {
		buffer.Write([]byte{byte(i)})
	}
	fmt.Println(buffer.rPos, buffer.wPos)
	head, tail = buffer.ReadRaw()
	fmt.Println(head)
	fmt.Println(tail)
}
