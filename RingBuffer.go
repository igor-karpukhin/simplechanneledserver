package SimpleChanneledServer

import "errors"

type RingBuffer struct {
	items        []uint64
	currentIndex int
}

func NewRingBuffer(capacity uint64) (b *RingBuffer) {
	b = new(RingBuffer)
	b.items = make([]uint64, capacity)
	b.currentIndex = 0
	return b
}

func (b *RingBuffer) TailReached() (reached bool) {
	return b.currentIndex > len(b.items)-1
}

func (b *RingBuffer) CopyDataFrom(data []uint64) (err error) {
	if len(data) != len(b.items) {
		return errors.New("Size mismatch")
	}
	for i := 0; i < len(b.items); i++ {
		b.items[i] = data[i]
	}
	return nil
}

func (b *RingBuffer) AddItem(item uint64) {
	if b.TailReached() {
		b.currentIndex = 0
	}
	b.items[b.currentIndex] = item

	b.currentIndex++
}

func (b *RingBuffer) InsertAt(item uint64, pos int) (err error) {
	if pos > len(b.items)-1 {
		return errors.New("Index out of range")
	}
	b.items[pos] = item
	return nil
}

func (b *RingBuffer) SummElements() (result uint64) {
	result = 0
	for i := 0; i < len(b.items); i++ {
		result += b.items[i]
	}
	return result
}

func (b *RingBuffer) Map(fn func(item uint64)) {
	for i := 0; i < len(b.items); i++ {
		fn(b.items[i])
	}
}
