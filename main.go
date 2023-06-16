package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

func main() {
}

func subtest(ctx context.Context, name string, doneCallback func()) {
	index := 0
	for {
		time.Sleep(100 * time.Millisecond)
		index++
		select {
		case <-ctx.Done():
			log.Printf("%s: %s\n", name, context.Cause(ctx))
			if value, ok := ctx.Value("parent").(string); ok {
				log.Printf("%s: parent:%s\n", name, value)
			}
			if doneCallback != nil {
				doneCallback()
			}
			return
		default:
			log.Printf("%s: %d working...\n", name, index)
		}
	}
}

// clientChan must be initialized before use
var clientChan chan string

func producer(content string) {
	clientChan <- content
}

func consumer() *string {
	if content, ok := <-clientChan; ok {
		return &content
	}
	return nil
}

// dagChanMap must be initialized before use
var dagChanMap map[string]chan string

func dagInput(content string, chanNames ...string) {
	for _, next := range chanNames {
		next := next
		go func(next string) {
			dagChanMap[next] <- content
		}(next)
	}
}

func dagOutput(chanNames ...string) *string {
	var count = len(chanNames)
	var contents = make([]string, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i, name := range chanNames {
		i := i
		name := name
		go func(i int, name string) {
			contents[i] = <-dagChanMap[name]
			wg.Done()
		}(i, name)
	}
	wg.Wait()
	var content = strings.Join(contents, ",")
	return &content
}

type DAG[T1 interface{}, T2 interface{}] interface {
}
