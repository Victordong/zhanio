package zhanio

type RingBuffer struct {
	buf     []byte
	size    int
	rPos    int
	wPos    int
	isEmpty bool
}

func NewBuffer(size int) *RingBuffer {
	return nil
}

func (r *RingBuffer) Read(p []byte) (int, error) {
	return 0, nil
}

func (r *RingBuffer) Write(p []byte) (int, error) {
	return 0, nil
}

func (r *RingBuffer) IsFull() bool {
	return !r.isEmpty
}

func (r *RingBuffer) Length() int {
	return 0
}

func (r *RingBuffer) IsEmpty() bool {
	return r.isEmpty
}

func (r *RingBuffer) reset() {

}

func (r *RingBuffer) malloc(size int) {

}
