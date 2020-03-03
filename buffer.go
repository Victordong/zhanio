package zhanio

type RingBuffer struct {
	buf     []byte
	size    int
	rPos    int
	wPos    int
	isEmpty bool
}

const InitSize = 1024

func NewBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:     make([]byte, size),
		size:    size,
		rPos:    0,
		wPos:    0,
		isEmpty: true,
	}
}

func (r *RingBuffer) ReadRaw() ([]byte, []byte) {
	if r.isEmpty {
		return nil, nil
	}
	if r.rPos < r.wPos {
		return r.buf[r.rPos:r.wPos], nil
	} else {
		return r.buf[r.rPos:], r.buf[:r.wPos]
	}
}

func (r *RingBuffer) ClearN(n int) {
	if n <= 0 {
		return
	}
	if n < r.size {
		r.rPos = (r.rPos + n) % r.size
		if r.rPos == r.wPos {
			r.isEmpty = true
		}
	} else {
		r.Reset()
	}
}

func (r *RingBuffer) Write(p []byte) {
	n := len(p)
	if n > r.Free() {
		r.malloc(n - r.Free())
	}
	if r.wPos+n > r.size {
		head, tail := p[:r.size-r.wPos], p[r.size-r.wPos:]
		copy(r.buf[r.wPos:], head)
		copy(r.buf[r.wPos+len(head):], tail)
	} else {
		copy(r.buf[r.wPos:], p)
	}
	r.wPos = (r.wPos + n) % r.size
	r.isEmpty = false
}

func (r *RingBuffer) IsFull() bool {
	return !r.isEmpty && r.rPos == r.wPos
}

func (r *RingBuffer) Length() int {
	if r.wPos-r.rPos != 0 {
		return (r.wPos - r.rPos + r.size) % r.size
	} else {
		if r.isEmpty {
			return 0
		} else {
			return r.size
		}
	}
}

func (r *RingBuffer) Cap() int {
	return r.size
}

func (r *RingBuffer) Free() int {
	return r.size - r.Length()
}

func (r *RingBuffer) IsEmpty() bool {
	return r.isEmpty
}

func (r *RingBuffer) Reset() {
	r.wPos = 0
	r.rPos = 0
	r.isEmpty = true
}

func (r *RingBuffer) malloc(cap int) {
	var newSize int
	if cap == 0 {
		if r.size < InitSize*InitSize {
			newSize = r.size * 2
		} else {
			newSize = r.size + InitSize*InitSize
		}
	} else {
		newSize = r.size + cap
	}
	newBuf := make([]byte, newSize)
	head, tail := r.ReadRaw()
	copy(newBuf, head)
	copy(newBuf[len(head):], tail)
	r.wPos, r.rPos, r.size, r.buf = r.size, 0, newSize, newBuf
}
