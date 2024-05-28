package utils

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func demoFunc(id string) {
	fmt.Printf("%s-----start: %s\n", id, time.Now().Format(StrTimeFormatMill))
	<-time.After(2 * time.Second)
	fmt.Printf("%s-----end: %s\n", id, time.Now().Format(StrTimeFormatMill))
}

func panicFunc(id string) {
	fmt.Printf("%s-----start panic: %s\n", id, time.Now().Format(StrTimeFormatMill))
	<-time.After(2 * time.Second)
	panic(fmt.Errorf("id: %s is panic", id))
}

func demoTimeoutFunc(id string) {
	fmt.Printf("%s-----start Timeout: %s\n", id, time.Now().Format(StrTimeFormatMill))
	<-time.After(5 * time.Second)
	fmt.Printf("%s-----end Timeout: %s\n", id, time.Now().Format(StrTimeFormatMill))
}

func TestConcurrent(t *testing.T) {
	num := 3
	pool, err := NewPool(num)
	defer pool.Release()
	assert.NoError(t, err)
	var i int
	for ; i <= 20; i++ {
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		fmt.Println(pool.p.Waiting(), pool.p.Running())
		err = pool.Submit(func() {
			demoFunc(id)
		})
		assert.NoError(t, err, id)
	}
	pool.Wait()
}

func TestConcurrentByRoutine(t *testing.T) {
	num := 3
	pool, err := NewPool(num)
	defer pool.Release()
	assert.NoError(t, err)
	var i int
	for ; i <= 40; i++ {
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		go func(idc string) {
			err = pool.Submit(func() {
				fmt.Println(pool.p.Waiting(), pool.p.Running())
				demoFunc(idc)
			})
			assert.NoError(t, err, id)
		}(id)

	}
	pool.Wait()
}

func TestConcurrentByChan(t *testing.T) {
	pool := make(chan struct{}, 3)
	var wg sync.WaitGroup
	var i int
	for ; i <= 20; i++ {
		wg.Add(1)
		pool <- struct{}{}
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		go func(idc string) {
			defer func() {
				<-pool
				wg.Done()
			}()
			demoFunc(idc)
		}(id)
	}
	wg.Wait()
}

func TestPanicConcurrent(t *testing.T) {
	num := 10
	pool, err := NewPool(num)
	assert.NoError(t, err)
	defer pool.Release()
	var i int = 1
	for ; i <= 2; i++ {
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		err = pool.Submit(func() {
			demoFunc(id)
		})
		assert.NoError(t, err)
	}
	id := fmt.Sprint(i)
	fmt.Printf("%s-----insert panic: %s\n", id, time.Now().Format(StrTimeFormatMill))
	err = pool.Submit(func() {
		panicFunc(id)
	})
	assert.NoError(t, err)
	i++
	<-time.After(6 * time.Second)
	for ; i <= 5; i++ {
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		err = pool.Submit(func() {
			demoFunc(id)
		})
		assert.NoError(t, err)
	}
	id = fmt.Sprint(i)
	fmt.Printf("%s-----insert panic: %s\n", id, time.Now().Format(StrTimeFormatMill))
	err = pool.Submit(func() {
		panicFunc(id)
	})
	pool.Wait()
	// <-time.After(6 * time.Second)
}

func demoTimeoutFuncWithContext(ctx context.Context, id string) {
	fmt.Printf("%s-----start Timeout: %s\n", id, time.Now().Format(StrTimeFormatMill))
	defer func() {
		fmt.Printf("%s-----end Timeout: %s\n", id, time.Now().Format(StrTimeFormatMill))
	}()
	tic := time.NewTicker(time.Second)
	for {
		select {
		case <-tic.C:
			fmt.Println(time.Now())
		case <-ctx.Done():
			fmt.Printf("%s is Done\n", id)
			return
		}
	}
}

func TestTimeoutConcurrent(t *testing.T) {
	num := 3
	pool, err := NewPool(num)
	assert.NoError(t, err)
	defer pool.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for i := 1; i <= 10; i++ {
		id := fmt.Sprint(i)
		fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
		err = pool.Submit(func() {
			demoTimeoutFuncWithContext(ctx, id)
		})
		assert.NoError(t, err)
	}
	pool.Wait()
	// <-time.After(6 * time.Second)
	pool.Release()
}

func TestForConcurrent(t *testing.T) {
	num := 3
	pool, err := NewPool(num, WithExpiryDuration(2*time.Second))
	assert.NoError(t, err)
	defer pool.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	for {
		for i := 1; i <= 10; i++ {
			id := fmt.Sprint(i)
			fmt.Printf("%s-----insert: %s\n", id, time.Now().Format(StrTimeFormatMill))
			err = pool.Submit(func() {
				demoTimeoutFunc(id)
			})
			assert.NoError(t, err)
		}
		select {
		case <-ctx.Done():
			fmt.Println("is done")
			return
		}
	}
}
