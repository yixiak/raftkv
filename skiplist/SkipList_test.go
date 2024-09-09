package skiplist_test

import (
	"raftkv/skiplist"
	"sync"
	"testing"
	"time"

	"math/rand"
)

var (
	TEST_COUNT    = 500000
	ROUTINE_COUNT = 1
)

func TestSkipListIn1(t *testing.T) {

	list := skiplist.NewSkipList("temp")
	st := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(ROUTINE_COUNT)
	for i := 0; i < ROUTINE_COUNT; i++ {
		go func() {
			for i := 0; i < TEST_COUNT/ROUTINE_COUNT; i++ {
				list.Insert(int(rand.Int31n(int32(TEST_COUNT))), "a")
			}
			wg.Done()
		}()

	}
	wg.Wait()
	end := time.Since(st)
	t.Logf("insert %d items cost %d ms with %d goroutine", TEST_COUNT, end.Milliseconds(), ROUTINE_COUNT)

	// test QPS

	st = time.Now()
	wg.Add(ROUTINE_COUNT)
	for i := 0; i < ROUTINE_COUNT; i++ {
		go func() {
			for i := 0; i < TEST_COUNT/ROUTINE_COUNT; i++ {
				list.Seach(int(rand.Int31n(int32(TEST_COUNT))))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	end = time.Since(st)
	list.Store()
	t.Logf("search %d items cost %d ms with %d goroutine", TEST_COUNT, end.Milliseconds(), ROUTINE_COUNT)
}
