package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// TestSubContext tests subcontext.
func TestSubContext(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()

	// Set the timeout to 1 second.
	// The objective function is executed every 100 milliseconds, which means it is executed ten times.
	t.Run("with timeout", func(t *testing.T) {
		sub1, _ := context.WithTimeout(root, 1*time.Second)
		subtest(sub1, "sub1", nil)
		time.Sleep(10 * time.Millisecond)
	})

	// Set the timeout to 1 second from now.
	// The objective function is executed every 100 milliseconds, which means it is executed ten times.
	t.Run("with deadline", func(t *testing.T) {
		sub2, _ := context.WithDeadline(root, time.Now().Add(1*time.Second))
		subtest(sub2, "sub2", nil)
		time.Sleep(10 * time.Millisecond)
	})

	// Call the cancel function handle after 1 second.
	// The objective function is executed every 100 milliseconds, which means it is executed ten times.
	t.Run("with cancel", func(t *testing.T) {
		sub3, cancelFunc3 := context.WithCancel(root)
		go subtest(sub3, "sub3", nil)
		time.Sleep(1 * time.Second)
		cancelFunc3()
		time.Sleep(10 * time.Millisecond)
	})

	// Call the cancel function handle after 1 second, and pass in the cancellation reason.
	// The objective function is executed every 100 milliseconds, which means it is executed ten times.
	t.Run("with cancel cause", func(t *testing.T) {
		sub4, cancelFunc4 := context.WithCancelCause(root)
		go subtest(sub4, "sub4", nil)
		time.Sleep(1 * time.Second)
		cancelFunc4(errors.New("stop"))
		time.Sleep(10 * time.Millisecond)
	})
}

// TestNestedContext tests nested contexts.
func TestNestedContext(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()

	t.Run("with timeout and its sub context with value", func(t *testing.T) {
		log.Println("When the parent times out, all children time out.")
		sub1, _ := context.WithTimeout(root, 1*time.Second)
		sub11 := context.WithValue(sub1, "parent", "sub1")
		go subtest(sub11, "sub11", nil)
		subtest(sub1, "sub1", nil)
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with deadline and its sub context with value", func(t *testing.T) {
		log.Println("When the parent reaches the deadline, the children also reach the same deadline.")
		sub2, _ := context.WithDeadline(root, time.Now().Add(1*time.Second))
		sub21 := context.WithValue(sub2, "parent", "sub2")
		go subtest(sub21, "sub21", nil)
		subtest(sub2, "sub2", nil)
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel and its sub context with value", func(t *testing.T) {
		log.Println("The parent context cancelled, the children stopped at the same time.")
		sub3, cancelFunc3 := context.WithCancel(root)
		sub31 := context.WithValue(sub3, "parent", "sub3")
		go subtest(sub31, "sub31", nil)
		go subtest(sub3, "sub3", nil)
		time.Sleep(1 * time.Second)
		cancelFunc3()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel cause and its sub context with value", func(t *testing.T) {
		log.Println("The parent cancelled with custom cause. The children also stops for the same cause.")
		sub4, cancelFunc4 := context.WithCancelCause(root)
		sub41 := context.WithValue(sub4, "parent", "sub4")
		go subtest(sub41, "sub41", nil)
		go subtest(sub4, "sub4", nil)
		time.Sleep(1 * time.Second)
		cancelFunc4(errors.New("stop")) // sub41 also output "stop"
		time.Sleep(10 * time.Millisecond)
	})
}

// TestNestedContextWithValue tests nested context with value.
func TestNestedContextWithValue(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()

	t.Run("with value and its sub context cancel", func(t *testing.T) {
		log.Println("If the child does not have the same key set, the parent's key will be propagated here.")
		sub1 := context.WithValue(root, "parent", "root")
		sub11, _ := context.WithCancel(sub1)
		log.Println(sub1.Value("parent"))  // root
		log.Println(sub11.Value("parent")) // root, if key not set, the corresponding value of the parent are propagated here.
	})

	t.Run("with value and its sub context value with same key", func(t *testing.T) {
		log.Println("If the child sets the key set by the parent node, the child's own value shall prevail.")
		sub2 := context.WithValue(root, "parent", "root")
		sub21 := context.WithValue(sub2, "parent", "sub2")
		log.Println(sub2.Value("parent"))  // root
		log.Println(sub21.Value("parent")) // sub2
	})

	t.Run("with value and its sub context value with different key", func(t *testing.T) {
		log.Println("The child sets the key set by the parent, and sets the key not set by the parent.")
		sub3 := context.WithValue(root, "parent", "root")
		sub31 := context.WithValue(sub3, "parent_sub3", "sub3")
		log.Println(sub3.Value("parent"))       // root
		log.Println(sub31.Value("parent"))      // root
		log.Println(sub31.Value("parent_sub3")) // sub3
	})
}

func TestWaitGroup(t *testing.T) {
	root := context.Background()

	t.Run("wait for all tasks done before continuing", func(t *testing.T) {
		const N = 5
		sub1, _ := context.WithCancel(root)
		var subContexts [N]context.Context
		var wg sync.WaitGroup
		wg.Add(N)
		for i := 0; i < N; i++ {
			i := i
			go func() {
				rands := rand.New(rand.NewSource(time.Now().UnixNano()))
				subContexts[i], _ = context.WithTimeout(sub1, time.Duration(rands.Uint32()%400+100)*time.Millisecond)
				subtest(subContexts[i], fmt.Sprintf("sub1%d", i), wg.Done)
			}()
		}
		wg.Wait()
		log.Println("all done")
	})

	t.Run("all tasks are waiting for notification to start", func(t *testing.T) {
		const N = 5
		value := [N]int{0, 0, 0, 0, 0}
		sub2, _ := context.WithCancel(root)
		var subContexts [N]context.Context
		var wgNotification, wgReady sync.WaitGroup
		wgNotification.Add(1)
		wgReady.Add(N)
		for i := 0; i < N; i++ {
			i := i
			go func() {
				wgNotification.Wait()
				rands := rand.New(rand.NewSource(time.Now().UnixNano()))
				subContexts[i], _ = context.WithTimeout(sub2, time.Duration(rands.Uint32()%400+100)*time.Millisecond)
				subtest(subContexts[i], fmt.Sprintf("sub2%d", i), func() {
					value[i] = 1
					wgReady.Done()
				})
			}()
		}
		time.Sleep(1 * time.Second)
		for i := 0; i < N; i++ {
			i := i
			assert.Equal(t, 0, value[i])
		}
		wgNotification.Done()
		log.Println("start")
		wgReady.Wait()
		for i := 0; i < N; i++ {
			i := i
			assert.Equal(t, 1, value[i])
		}
		log.Println("all done")
	})
}
