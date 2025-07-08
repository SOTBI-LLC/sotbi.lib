package utils

import (
	"sync"
	"testing"
)

func TestSonyflake_UniquenessAndConcurrency(t *testing.T) {
	sf := NewSonyflake(SonyflakeConfig{MachineID: 1})

	const n = 10000
	ids := make([]uint64, n)

	var wg sync.WaitGroup

	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			id, err := sf.NextID()
			if err != nil {
				t.Errorf("NextID error: %v", err)
				return
			}

			ids[idx] = id

			wg.Done()
		}(i)
	}

	wg.Wait()

	idMap := make(map[uint64]struct{}, n)
	for _, id := range ids {
		if _, exists := idMap[id]; exists {
			t.Errorf("Duplicate ID found: %d", id)
		}

		idMap[id] = struct{}{}
	}
}
