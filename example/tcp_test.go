package main

import (
	"fmt"
	"testing"
)

func TestHandler_Closed(t *testing.T) {
	head := make([]byte, 0)
	head = append(head, []byte{'a', 'b', 'c'}...)
	var tail []byte = nil
	result := make([]byte, 10)
	copy(result, head)
	copy(result[len(head):], tail)
	fmt.Println(result)
}
