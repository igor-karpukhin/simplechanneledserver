package PersistentCounter

import "testing"

func TestNewRingBuffer(t *testing.T) {
	r := NewRingBuffer(10)
	if len(r.items) != 10 {
		t.Fail()
	}
}

func TestRingBuffer_AddItem(t *testing.T) {
	r := NewRingBuffer(10)

	for i := 0; i < 10; i++ {
        r.AddItem(uint64(i))
	}

    if len(r.items) != 10 {
		t.Fail()
	}
}

func TestRingBuffer_Map(t *testing.T) {
	r := NewRingBuffer(10)

	for i := 0; i < 10; i++ {
        r.AddItem(uint64(i))
	}

	result := []int{}

	r.Map(func (item uint64){
		result = append(result)
	})

	for i := 0; i < len(result); i++ {
		if result[i] != i {
			t.Fail()
		}
	}
}
