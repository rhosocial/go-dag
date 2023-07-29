package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
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

func TestNestedContextWithSiblings(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	root := context.Background()

	// The current node is cancelled, and sibling nodes are not affected.
	t.Run("sibling canceled", func(t *testing.T) {
		sub1 := context.WithValue(root, "parent", "root")
		sub11, cancel11 := context.WithCancel(sub1)
		defer cancel11() // The cancel handler must be called here, otherwise it will cause a leak and affect the accuracy of other tests.
		sub12, cancel12 := context.WithCancel(sub1)
		go subtest(sub11, "sub11", nil)
		go subtest(sub12, "sub12", nil)
		cancel12()
		time.Sleep(time.Second)
	})

	// The current node is canceled and all sibling nodes are notified to cancel.
	t.Run("the current canceled, and notify the siblings to cancel", func(t *testing.T) {
		sub1 := context.WithValue(root, "parent", "root")
		sub11, cancel11 := context.WithCancel(sub1)
		sub12, cancel12 := context.WithCancel(sub1)
		go subtest(sub11, "sub11", nil)
		go subtest(sub12, "sub12", func() { cancel11() })
		time.Sleep(time.Millisecond * 100)
		cancel12()
		time.Sleep(time.Second)
	})

	t.Run("the current canceled, and notify the siblings and their children to cancel", func(t *testing.T) {
		sub1 := context.WithValue(root, "parent", "root")
		sub11, cancel11 := context.WithCancel(sub1)
		sub12, cancel12 := context.WithCancel(sub11)
		sub21, cancel21 := context.WithCancel(sub1)
		sub22, cancel22 := context.WithCancel(sub21)
		cancel := func() {
			cancel11()
			cancel21()
		}
		go subtest(sub11, "sub11", cancel)
		go subtest(sub12, "sub12", func() { cancel12() })
		go subtest(sub21, "sub21", cancel)
		go subtest(sub22, "sub22", func() { cancel21() })
		time.Sleep(time.Millisecond * 100)
		cancel22()
		time.Sleep(time.Second)

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

func TestErrGroup(t *testing.T) {
	root := context.Background()

	t.Run("errgroup only records the first error received.", func(t *testing.T) {
		const N = 5
		g, _ := errgroup.WithContext(root)
		var sub [5]bool
		for i := 0; i < N; i++ {
			i := i
			g.Go(func() error {
				sub[i] = true
				log.Printf("sub1%d", i)
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				return fmt.Errorf("err sub1%d", i)
			})
		}

		// Although only the first error is recorded, it will not continue until all goroutines are executed.
		// That is, if one of them reports an error, it will not prevent other goroutines from continuing to execute.
		if err := g.Wait(); err != nil {
			assert.Equal(t, "err sub10", err.Error())
		}
		for _, s := range sub {
			assert.True(t, s)
		}
	})

	// The parent context of errgroup is the context with the handle of the cancellation function.
	t.Run("The parent context of errgroup is canceled directly.", func(t *testing.T) {
		const N = 5
		sub1, cancelFunc := context.WithCancel(root)
		g, _ := errgroup.WithContext(sub1)
		var stats [5]bool
		for i := 0; i < N; i++ {
			i := i
			g.Go(func() error {
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				stats[i] = true
				return fmt.Errorf("err sub1%d", i)
			})
		}
		// After errgroup starts all coroutines, the parent context is directly canceled without waiting for them to end.
		cancelFunc()
		// But this does not cause all coroutines to be undone.
		// errgroup still records the first returned error.
		if err := g.Wait(); err != nil {
			assert.Equal(t, "err sub10", err.Error())
		}
		// and all goroutines have been executed.
		for _, s := range stats {
			assert.True(t, s)
		}
	})

	// Each goroutine started by errgroup monitors the completion status of the parent context.
	// If the context ends, the specified function is executed.
	t.Run("errgroup with goroutine checks context done status and execute done function.", func(t *testing.T) {
		const N = 5
		sub1, cancelFunc := context.WithCancel(root)
		g, _ := errgroup.WithContext(sub1)
		var stats [5]bool
		for i := 0; i < N; i++ {
			i := i
			g.Go(func() error {
				time.Sleep(time.Duration((i+1)*100) * time.Millisecond)
				return subtesterr(sub1, "sub1", func() {
					stats[i] = true
				})
			})
		}
		cancelFunc()
		if err := g.Wait(); err != nil {
			assert.Equal(t, "context canceled", err.Error())
		}
		// and all done functions have been executed.
		for _, s := range stats {
			assert.True(t, s)
		}
	})

	// Each goroutine started by errgroup monitors the completion status of the parent context.
	// No function is executed after the context ends.
	t.Run("errgroup with goroutine checks context done status without executing any done function.", func(t *testing.T) {
		const N = 5
		sub1, cancelFunc := context.WithCancel(root)
		g, _ := errgroup.WithContext(sub1)
		var stats [5]bool
		for i := 0; i < N; i++ {
			i := i
			g.Go(func() error {
				time.Sleep(time.Duration((i+1)*100) * time.Millisecond)
				return subtesterr(sub1, "sub1", nil)
			})
		}
		cancelFunc()
		if err := g.Wait(); err != nil {
			assert.Equal(t, "context canceled", err.Error())
		}
		// and no done functions have been executed.
		for _, s := range stats {
			assert.False(t, s)
		}
	})

	// One of the goroutines in the errgroup reports an error and notifies the other goroutines to abort.
	t.Run("One of the goroutines in the errgroup reports an error and notifies the other goroutines to abort.", func(t *testing.T) {
		const N = 5
	})
}

func TestExitAfterBeingNotified(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	c := make(chan int)
	go func() {
		log.Println("channel input")
		time.Sleep(time.Second)
		c <- 10
		// close(c)
	}()
	b := false
	for {
		if b {
			log.Println("exiting...")
			break
		}
		select {
		case x, ok := <-c:
			b = true
			assert.True(t, ok)
			assert.Equal(t, 10, x)
		}
	}
}

func TestProduceAndConsume(t *testing.T) {
	// Since there is no consumer listening to the channel during production, a buffer must be set for the channel,
	// otherwise it will block for a long time because there is no consumer and eventually lead to deadlock.
	t.Run("Produce first, consume later.", func(t *testing.T) {
		clientChan = make(chan string, 1)
		defer close(clientChan)
		producer("channel")
		content := consumer()
		assert.NotNil(t, content)
		assert.Equal(t, "channel", *content)
	})

	// If it is produced twice and consumed once, the consumer gets the content in the order of production.
	t.Run("Produce twice first, consume later.", func(t *testing.T) {
		clientChan = make(chan string, 2)
		defer close(clientChan)
		producer("channel1")
		producer("channel2")
		content := consumer()
		assert.NotNil(t, content)
		assert.Equal(t, "channel1", *content)
	})

	// The consumer first listens to the channel, and the producer then sends data to the channel.
	// The consumer will block until it receives data.
	t.Run("Consumer wait for producer to produce.", func(t *testing.T) {
		clientChan = make(chan string)
		defer close(clientChan)
		go func() {
			content := consumer()
			assert.NotNil(t, content)
			assert.Equal(t, "channel1", *content)
		}()
		producer("channel1")
	})
}

func TestDAGSequential(t *testing.T) {
	// The flowchart is:
	//         chanInOut
	// input -------------> output
	t.Run("Input and output only", func(t *testing.T) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["chanInOut"] = make(chan string)
		defer close(dagChanMap["chanInOut"])
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			content := dagOutput("chanInOut")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			t.Log(*content)
			wg.Done()
		}()
		dagInput("channelInput", "chanInOut")
		wg.Wait()
		t.Log("finished.")
	})

	// The flowchart is:
	// (chan name)  chanInTran              chanTranOut
	//      input --------------> transit --------------> output
	t.Run("A single input and a single output, with a transit in between.", func(t *testing.T) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["chanInTran"] = make(chan string)
		defer close(dagChanMap["chanInTran"])
		dagChanMap["chanTranOut"] = make(chan string)
		defer close(dagChanMap["chanTranOut"])

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			content := dagOutput("chanTranOut")
			t.Log("output")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("chanInTran")
			dagInput(*content, "chanTranOut")
			t.Log("transit")
			wg.Done()
		}()
		t.Log("input")
		dagInput("channelInput", "chanInTran")
		wg.Wait()
		t.Log("finished.")
	})

	// The flowchart is:
	//            input                   transit1
	// input --------------> transit ---------------> output1
	//                            |       transit2
	//                            + ----------------> output2
	//                            |       transit3
	//                            + ----------------> output3
	t.Run("A single input and three outputs, with a transit in between.", func(t *testing.T) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["input"] = make(chan string)
		defer close(dagChanMap["input"])
		dagChanMap["transit1"] = make(chan string)
		defer close(dagChanMap["transit1"])
		dagChanMap["transit2"] = make(chan string)
		defer close(dagChanMap["transit2"])
		dagChanMap["transit3"] = make(chan string)
		defer close(dagChanMap["transit3"])

		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			content := dagOutput("transit3")
			t.Log("output1")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit2")
			t.Log("output2")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit1")
			t.Log("output3")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("input")
			dagInput(*content, "transit1", "transit2", "transit3")
			t.Log("transit")
			wg.Done()
		}()
		t.Log("input")
		dagInput("channelInput", "input")
		wg.Wait()
		t.Log("finished.")
	})

	// The flowchart is:
	//         input               transit1               transit2               transit3
	// input ---------> transit1 ------------> transit2 ------------> transit3 ------------> output
	t.Run("A single input and a single output, with three sequential transits in between.", func(t *testing.T) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["input"] = make(chan string)
		defer close(dagChanMap["input"])
		dagChanMap["transit1"] = make(chan string)
		defer close(dagChanMap["transit1"])
		dagChanMap["transit2"] = make(chan string)
		defer close(dagChanMap["transit2"])
		dagChanMap["transit3"] = make(chan string)
		defer close(dagChanMap["transit3"])

		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			content := dagOutput("transit3")
			t.Log("output")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit2")
			t.Log("transit3")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			dagInput(*content, "transit3")
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit1")
			t.Log("transit2")
			assert.NotNil(t, content)
			assert.Equal(t, "channelInput", *content)
			dagInput(*content, "transit2")
			wg.Done()
		}()
		go func() {
			content := dagOutput("input")
			dagInput(*content, "transit1")
			t.Log("transit1")
			wg.Done()
		}()
		t.Log("input")
		dagInput("channelInput", "input")
		wg.Wait()
		t.Log("finished.")
	})
}

func TestDAGComplex(t *testing.T) {
	// The two inputs are propagated to three transits respectively,
	// and then each is propagated to two output nodes.
	// The flowchart is:
	//         +-------> transit1 --------+------> output1
	//         |            |             |
	// input1 -+  +---------+             |
	//         |  |                       |
	//         +-------> transit2 --------+
	//         |  |         |             |
	// input2 ----+---------+             |
	//         |  |                       |
	//         +-------> transit3 --------+------> output2
	//            |         |
	//            +---------+
	t.Run("Multiple inputs and a single output with several transits in between.", func(t *testing.T) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["input11"] = make(chan string)
		defer close(dagChanMap["input11"])
		dagChanMap["input12"] = make(chan string)
		defer close(dagChanMap["input12"])
		dagChanMap["input13"] = make(chan string)
		defer close(dagChanMap["input13"])
		dagChanMap["input21"] = make(chan string)
		defer close(dagChanMap["input21"])
		dagChanMap["input22"] = make(chan string)
		defer close(dagChanMap["input22"])
		dagChanMap["input23"] = make(chan string)
		defer close(dagChanMap["input23"])
		dagChanMap["transit1"] = make(chan string)
		defer close(dagChanMap["transit1"])
		dagChanMap["transit2"] = make(chan string)
		defer close(dagChanMap["transit2"])
		dagChanMap["transit3"] = make(chan string)
		defer close(dagChanMap["transit3"])

		var wg sync.WaitGroup
		wg.Add(4)

		go func() {
			content := dagOutput("transit1", "transit2", "transit3")
			t.Log("output")
			assert.NotNil(t, content)
			t.Log(*content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("input11", "input21")
			assert.NotNil(t, content)
			dagInput(*content, "transit1")
			t.Log("transit1")
			assert.NotNil(t, content)
			t.Log(*content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("input12", "input22")
			assert.NotNil(t, content)
			dagInput(*content, "transit2")
			t.Log("transit2")
			assert.NotNil(t, content)
			t.Log(*content)
			wg.Done()
		}()
		go func() {
			content := dagOutput("input13", "input23")
			assert.NotNil(t, content)
			dagInput(*content, "transit3")
			t.Log("transit3")
			assert.NotNil(t, content)
			t.Log(*content)
			wg.Done()
		}()
		t.Log("input")
		start := time.Now().UnixMicro()
		dagInput("channelInput1", "input11", "input12", "input13")
		dagInput("channelInput2", "input21", "input22", "input23")
		wg.Wait()
		end := time.Now().UnixMicro()
		t.Log("finished.")
		t.Log(fmt.Sprintf("Time Elapsed: %d (micro sec)", end-start))
	})
}

func BenchmarkDAGSequential(b *testing.B) {
	b.Run("A single input and a single output, with three sequential transits in between.", func(t *testing.B) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["input"] = make(chan string)
		defer close(dagChanMap["input"])
		dagChanMap["transit1"] = make(chan string)
		defer close(dagChanMap["transit1"])
		dagChanMap["transit2"] = make(chan string)
		defer close(dagChanMap["transit2"])
		dagChanMap["transit3"] = make(chan string)
		defer close(dagChanMap["transit3"])

		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			dagOutput("transit3")
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit2")
			dagInput(*content, "transit3")
			wg.Done()
		}()
		go func() {
			content := dagOutput("transit1")
			dagInput(*content, "transit2")
			wg.Done()
		}()
		go func() {
			content := dagOutput("input")
			dagInput(*content, "transit1")
			wg.Done()
		}()
		dagInput("channelInput", "input")
		wg.Wait()
	})
}

func BenchmarkDAGComplex(b *testing.B) {
	b.Run("Multiple inputs and a single output with several transits in between.", func(t *testing.B) {
		dagChanMap = make(map[string]chan string)
		dagChanMap["input11"] = make(chan string)
		defer close(dagChanMap["input11"])
		dagChanMap["input12"] = make(chan string)
		defer close(dagChanMap["input12"])
		dagChanMap["input13"] = make(chan string)
		defer close(dagChanMap["input13"])
		dagChanMap["input21"] = make(chan string)
		defer close(dagChanMap["input21"])
		dagChanMap["input22"] = make(chan string)
		defer close(dagChanMap["input22"])
		dagChanMap["input23"] = make(chan string)
		defer close(dagChanMap["input23"])
		dagChanMap["transit1"] = make(chan string)
		defer close(dagChanMap["transit1"])
		dagChanMap["transit2"] = make(chan string)
		defer close(dagChanMap["transit2"])
		dagChanMap["transit3"] = make(chan string)
		defer close(dagChanMap["transit3"])

		var wg sync.WaitGroup
		wg.Add(4)

		go func() {
			dagOutput("transit1", "transit2", "transit3")
			wg.Done()
		}()
		go func() {
			content := dagOutput("input11", "input21")
			dagInput(*content, "transit1")
			wg.Done()
		}()
		go func() {
			content := dagOutput("input12", "input22")
			dagInput(*content, "transit2")
			wg.Done()
		}()
		go func() {
			content := dagOutput("input13", "input23")
			dagInput(*content, "transit3")
			wg.Done()
		}()
		dagInput("channelInput1", "input11", "input12", "input13")
		dagInput("channelInput2", "input21", "input22", "input23")
		wg.Wait()
	})
}

type DAGStraightPipe struct {
	DAG[string, string]
}

func NewDAGStraightPipe() *DAGStraightPipe {
	f := DAGStraightPipe{}
	f.InitChannels("straight_pipe")
	f.InitWorkflow("straight_pipe", "straight_pipe")
	return &f
}

func TestDAGStraightPipe(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	t.Run("run", func(t *testing.T) {
		f := NewDAGStraightPipe()
		var input = "test"
		var results = f.Run(&input)
		assert.Equal(t, "test", *results)
	})
}

type DAGOneTransit struct {
	DAG[string, string]
}

// NewDAGOneTransit defines a workflow.
func NewDAGOneTransit() *DAGOneTransit {
	f := DAGOneTransit{}
	f.InitChannels("input", "output")
	//         input              output
	// input ---------> transit ----------> output
	f.InitWorkflow("input", "output", &DAGWorkflowTransit{
		Name:           "transit",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) any {
			log.Println("transit: ", a)
			return a[0]
		},
	})
	return &f
}

func TestDAGOneTransit(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	t.Run("run", func(t *testing.T) {
		f := NewDAGOneTransit()
		var input = "test"
		var results = f.Run(&input)
		t.Log(results)
	})
}

type DAGTwoParallelTransits struct {
	DAG[string, string]
}

// NewDAGTwoParallelTransits defines a workflow.
func NewDAGTwoParallelTransits() *DAGTwoParallelTransits {
	f := DAGTwoParallelTransits{}
	f.InitChannels("input", "t11", "t12", "t21", "t22", "output")
	//         input               t11              output
	// input ----+----> transit1 -------> transit ----------> output
	//           |                 t12       ^
	//           +----> transit2 ------------+
	f.InitWorkflow("input", "output", &DAGWorkflowTransit{
		Name:           "input",
		channelInputs:  []string{"input"},
		channelOutputs: []string{"t11", "t12"},
		worker: func(a ...any) any {
			return a[0]
		},
	}, &DAGWorkflowTransit{
		Name:           "transit1",
		channelInputs:  []string{"t11"},
		channelOutputs: []string{"t21"},
		worker: func(a ...any) any {
			return a[0]
		},
	}, &DAGWorkflowTransit{
		Name:           "transit2",
		channelInputs:  []string{"t12"},
		channelOutputs: []string{"t22"},
		worker: func(a ...any) any {
			return a[0]
		},
	}, &DAGWorkflowTransit{
		Name:           "transit",
		channelInputs:  []string{"t21", "t22"},
		channelOutputs: []string{"output"},
		worker: func(a ...any) any {
			var r string
			for i, c := range a {
				r = r + strconv.Itoa(i) + c.(string)
			}
			return r
		},
	})
	return &f
}

func TestDAGTwoParallelTransits(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	t.Run("run", func(t *testing.T) {
		f := NewDAGTwoParallelTransits()
		var input = "test"
		var results = f.Run(&input)
		assert.Equal(t, "0test1test", *results)
	})
}

func TestSelectCaseReturning(t *testing.T) {
	t.Run("for-loop case", func(t *testing.T) {
		signal := make(chan struct{})
		signalText := make(chan string)
		go func() {
			defer func() {
				signalText <- "exited"
			}()
			for {
				select {
				case <-signal:
					return
				default:
				}
			}
		}()
		time.Sleep(time.Second)
		signal <- struct{}{}
		assert.Equal(t, "exited", <-signalText)
	})
	t.Run("wait-group normal case", func(t *testing.T) {
		signal := make(chan struct{})
		signalText := make(chan string)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer func() {
				signalText <- "exited"
			}()
			defer wg.Done()
			for {
				select {
				case <-signal:
					return
				default:
				}
			}
		}()
		signal <- struct{}{}
		wg.Wait()
		assert.Equal(t, "exited", <-signalText)
	})
}

func TestOneDoneMultiNotified(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	t.Run("one done and multi are notified", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		notified := func(index int) {
			defer wg.Done()
			flag := true
			for flag {
				select {
				case <-ctx.Done():
					log.Println("worker[%i] notified to stop.", index)
					flag = false
				}
			}
		}
		wg.Add(2)
		go notified(1)
		go notified(2)
		cancel()
		wg.Wait()
		log.Println("finished.")
		time.Sleep(time.Millisecond * 10)
	})
	// Note that the following logic will only loop once.
	t.Run("The main notifies the subsequent two worker to start and end in sequence without default branch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan struct{})
		ch2 := make(chan struct{})
		notified1 := func() {
			<-ch1
			log.Println("worker 1 is notified to start.")
			defer func() {
				ch2 <- struct{}{}
				log.Println("worker 1 has notified the worker 2 to start.")
			}()
			flag := true
			index := 0
			for flag {
				time.Sleep(time.Millisecond * 100)
				index++
				select {
				case <-ctx.Done():
					log.Println("worker 1 is notified to stop.")
					flag = false
				}
			}
			log.Printf("worker 1 has looped %d times.", index)
		}
		notified2 := func() {
			<-ch2
			log.Println("worker 2 is notified to start.")
			flag := true
			index := 0
			for flag {
				time.Sleep(time.Millisecond * 100)
				index++
				select {
				case <-ctx.Done():
					log.Println("worker 2 is notified to stop.")
					flag = false
				}
			}
			log.Printf("worker 2 has looped %d times.", index)
			time.Sleep(time.Millisecond * 10)
		}
		go notified1()
		go notified2()
		log.Println("the main has notified the worker 1 to start.")
		ch1 <- struct{}{}
		time.Sleep(time.Second)
		log.Println("the main canceled.")
		cancel()
		time.Sleep(time.Millisecond * 100)
	})
	// If default is included, it will loop multiple times, otherwise the for-select will be scanned only once.
	t.Run("The main notifies the subsequent two worker to start and end in sequence with default branch.", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch1 := make(chan struct{})
		ch2 := make(chan struct{})
		notified1 := func() {
			<-ch1
			log.Println("worker 1 is notified to start.")
			defer func() {
				ch2 <- struct{}{}
				log.Println("worker 1 has notified the worker 2 to start.")
			}()
			flag := true
			index := 0
			for flag {
				index++
				select {
				case <-ctx.Done():
					log.Println("worker 1 is notified to stop.")
					flag = false
				default:
					time.Sleep(time.Millisecond * 100)
				}
			}
			log.Printf("worker 1 has looped %d times.", index)
		}
		notified2 := func() {
			<-ch2
			log.Println("worker 2 is notified to start.")
			flag := true
			index := 0
			for flag {
				index++
				select {
				case <-ctx.Done():
					log.Println("worker 2 is notified to stop.")
					flag = false
				default:
					time.Sleep(time.Millisecond)
				}
			}
			log.Printf("worker 2 has looped %d times.", index)
			time.Sleep(time.Millisecond * 10)
		}
		go notified1()
		go notified2()
		log.Println("the main has notified the worker 1 to start.")
		ch1 <- struct{}{}
		time.Sleep(time.Second)
		log.Println("the main canceled.")
		cancel()
		time.Sleep(time.Millisecond * 100)
	})

	t.Run("The", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		ch := make([]chan struct{}, 2) // Only the array of channels is declared, no channels are declared.
		var wg sync.WaitGroup
		notified := func(index int, cancel context.CancelFunc) {
			defer wg.Done()
			<-ch[index]
			flag := true
			times := 0
			for flag {
				times++
				select {
				case <-ctx.Done():
					log.Printf("worker %d is notified to stop.", index)
					flag = false
				default:
					time.Sleep(time.Millisecond * 100)
					if index == 0 && times > 2 {
						log.Println("canceled.")
						cancel()
					}
				}
			}
			log.Printf("worker %d has looped %d times.", index, times)
			time.Sleep(time.Millisecond * 10)
		}
		wg.Add(2)
		go notified(0, cancel)
		go notified(1, cancel)
		ch[0] = make(chan struct{})
		ch[1] = make(chan struct{})
		ch[0] <- struct{}{}
		ch[1] <- struct{}{}
		wg.Wait()
		log.Println("finished.")
	})
}
