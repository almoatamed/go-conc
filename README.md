concpool
=======

A tiny, dependency-free concurrent worker pool for Go. Submit tasks of type `func() error` and wait for all to complete. Each task's result is returned as a `TaskResult` describing success or the error returned by the task.

This project was refactored from a single-file example into the `concpool` package to make it easy to publish and import.

Installation
------------

Install with `go get` (module-aware Go):

    go get github.com/almoatamed/go-conc

Usage
-----

Import the package and create a new pool:
```go
import "github.com/almoatamed/go-conc/concpool"

p := concpool.New(10)
```

Submit tasks using `Run` and block until all tasks finish with `Wait`:

```go
p.Run(func() error {
    // do work
    return nil
})

results := p.Wait()
```

API
---

- func New(maxCount int) *Pool
  - Creates a new pool that runs up to `maxCount` tasks concurrently. If `maxCount <= 0` the function will use `1`.

- func (p *Pool) Run(task func() error)
  - Submit a task to the pool. Tasks are executed in FIFO order as workers free up.

- func (p *Pool) Wait() []TaskResult
  - Blocks until all submitted tasks have completed and returns a slice of `TaskResult` in the order tasks completed.

TaskResult
----------

- Success bool
- Err error

Notes
-----

- The pool is intentionally small and simple. It does not support context cancellation, task timeouts, or priorities. Those can be added in follow-up changes.

- Results are returned in completion order. If you need ordering by submission, attach sequence metadata to tasks or collect results differently.

Example
-------

See `main.go` in this repository for a simple example program that uses the package.
