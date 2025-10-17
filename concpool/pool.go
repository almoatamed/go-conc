// Package concpool provides a small concurrent worker pool for running
// functions of the form func() error. It is designed to be simple to use
// and publish as a standalone package.
package concpool

import (
	"sync"
)

// TaskResult represents the outcome of a single task executed by the pool.
type TaskResult struct {
	Success bool
	Err     error
}

// Pool runs up to maxCount tasks concurrently. Use New to create a pool,
// Run to submit tasks, and Wait to block until all submitted work is done.
type Pool struct {
	maxCount int
	queue    []func() error
	running  int
	results  chan TaskResult

	mu sync.Mutex

	terminated         bool
	terminationChannel chan bool
	runCheckChannel    chan bool
}

// New creates a new Pool that will run up to maxCount tasks concurrently.
func New(maxCount int) *Pool {
	if maxCount <= 0 {
		maxCount = 1
	}

	return &Pool{
		maxCount:           maxCount,
		queue:              make([]func() error, 0, maxCount*2),
		results:            make(chan TaskResult, 1),
		terminationChannel: make(chan bool, 1),
		runCheckChannel:    make(chan bool, 1),
	}
}

func (p *Pool) pushToQueue(task func() error) {
	p.mu.Lock()
	p.queue = append(p.queue, task)
	p.mu.Unlock()
}

func (p *Pool) attemptTermination() {
	p.mu.Lock()
	if p.terminated {
		p.mu.Unlock()
		return
	}
	p.terminated = true
	// non-blocking send so callers don't deadlock if Wait isn't yet listening
	select {
	case p.terminationChannel <- true:
	default:
	}
	p.mu.Unlock()
}

func (p *Pool) attemptCheck() {
	// non-blocking signal to ask the event loop to check the queue
	select {
	case p.runCheckChannel <- true:
	default:
	}
}

func (p *Pool) checkQueue() {
	p.mu.Lock()
	if p.terminated {
		p.mu.Unlock()
		return
	}

	available := p.maxCount - p.running
	if available <= 0 {
		p.mu.Unlock()
		return
	}

	for i := 0; i < available; i++ {
		if len(p.queue) == 0 {
			if p.running == 0 {
				p.mu.Unlock()
				p.attemptTermination()
				return
			}
			p.mu.Unlock()
			return
		}

		task := p.queue[0]
		p.queue = p.queue[1:]

		p.running++

		// run task in its own goroutine
		go func(t func() error) {
			err := t()
			if err != nil {
				p.results <- TaskResult{Success: false, Err: err}
			} else {
				p.results <- TaskResult{Success: true, Err: nil}
			}

			p.mu.Lock()
			p.running--
			p.mu.Unlock()

			// ask the loop to check if more work can be started
			p.attemptCheck()
		}(task)
	}

	p.mu.Unlock()
}

// Run submits a task to the pool. The task must be func() error.
// Tasks are executed in FIFO order as workers become available.
func (p *Pool) Run(task func() error) {
	p.pushToQueue(task)
	p.attemptCheck()
}

// Wait blocks until all submitted tasks have finished and returns the
// slice of TaskResult values in the order they completed.
func (p *Pool) Wait() []TaskResult {
	// start any available work
	p.checkQueue()

	results := make([]TaskResult, 0)

	for {
		select {
		case <-p.runCheckChannel:
			p.checkQueue()
		case r := <-p.results:
			results = append(results, r)
		case <-p.terminationChannel:
			return results
		}
	}
}
