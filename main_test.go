package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSubContext(t *testing.T) {
	root := context.Background()

	t.Run("with timeout", func(t *testing.T) {
		sub1, _ := context.WithTimeout(root, 1*time.Second)
		subtest(sub1, "sub1")
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with deadline", func(t *testing.T) {
		sub2, _ := context.WithDeadline(root, time.Now().Add(1*time.Second))
		subtest(sub2, "sub2")
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel", func(t *testing.T) {
		sub3, cancelFunc3 := context.WithCancel(root)
		go subtest(sub3, "sub3")
		time.Sleep(1 * time.Second)
		cancelFunc3()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel cause", func(t *testing.T) {
		sub4, cancelFunc4 := context.WithCancelCause(root)
		go subtest(sub4, "sub4")
		time.Sleep(1 * time.Second)
		cancelFunc4(errors.New("stop"))
		time.Sleep(10 * time.Millisecond)
	})
}

func TestNestedContext(t *testing.T) {
	root := context.Background()

	t.Run("with timeout and its sub context with value", func(t *testing.T) {
		sub1, _ := context.WithTimeout(root, 1*time.Second)
		sub11 := context.WithValue(sub1, "parent", "sub1")
		go subtest(sub11, "sub11")
		subtest(sub1, "sub1")
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with deadline and its sub context with value", func(t *testing.T) {
		sub2, _ := context.WithDeadline(root, time.Now().Add(1*time.Second))
		sub21 := context.WithValue(sub2, "parent", "sub2")
		go subtest(sub21, "sub21")
		subtest(sub2, "sub2")
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel and its sub context with value", func(t *testing.T) {
		sub3, cancelFunc3 := context.WithCancel(root)
		sub31 := context.WithValue(sub3, "parent", "sub3")
		go subtest(sub31, "sub31")
		go subtest(sub3, "sub3")
		time.Sleep(1 * time.Second)
		cancelFunc3()
		time.Sleep(10 * time.Millisecond)
	})

	t.Run("with cancel cause and its sub context with value", func(t *testing.T) {
		sub4, cancelFunc4 := context.WithCancelCause(root)
		sub41 := context.WithValue(sub4, "parent", "sub4")
		go subtest(sub41, "sub41")
		go subtest(sub4, "sub4")
		time.Sleep(1 * time.Second)
		cancelFunc4(errors.New("stop"))
		time.Sleep(10 * time.Millisecond)
	})
}
