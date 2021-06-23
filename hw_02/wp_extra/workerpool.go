/*
Worker Pool in which you can dynamically, while the program is running, add and remove workers.
*/

package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Something what can execute.
type Task interface {
	Execute()
}

// workerPool is object a pool of workers.
type Pool struct {
	minPoolSize  int
	maxPoolSize  int
	tasksChan    chan Task
	quitChan     chan struct{}
	workersCount int32
	waitGroup    *sync.WaitGroup
	sensorData   chan float64
}

// NewWorkerPool is function constructor to create new pool.
func NewPool(maxSize int, queueLen int) *Pool {
	pool := &Pool{
		minPoolSize: runtime.NumCPU() * 2, // cause can't be zero.
		maxPoolSize: maxSize,
		tasksChan:   make(chan Task, queueLen),
		quitChan:    make(chan struct{}),
		sensorData:  make(chan float64, 1),
		waitGroup:   &sync.WaitGroup{},
	}

	if pool.minPoolSize > pool.maxPoolSize {
		pool.minPoolSize = pool.maxPoolSize
	}

	// Start monitoring sensor metric
	pool.waitGroup.Add(1)
	go func() {
		defer pool.waitGroup.Done()
		pool.sensorCall()
	}()

	pool.waitGroup.Add(pool.maxPoolSize)
	for i := 0; i < pool.maxPoolSize; i++ {
		go pool.worker()
	}

	return pool
}

func (p *Pool) sensorCall() {
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

CPUload:
	for {
		select {
		case <-ticker.C:
			p.sensorData <- rand.Float64()
		case <-p.quitChan:
			break CPUload
		}
	}
	close(p.sensorData)
}

// worker method select task and execut it or exit.
func (p *Pool) worker() {
	defer p.waitGroup.Done()
	for {
		select {
		case task, ok := <-p.tasksChan:
			if !ok {
				return
			}
			task.Execute()
		case <-p.quitChan:
			return
		}
	}
}

// addWorker add ome more worker in pool.
func (p *Pool) addWorker() {
	atomic.AddInt32(&p.workersCount, 1)
	go p.worker()
}

// delWorker delete one worker form pool.
func (p *Pool) delWorker() {
	atomic.AddInt32(&p.workersCount, -1)
	p.quitChan <- struct{}{}
}

func (p *Pool) dispatchPool() {
	go func() {
		for {
			select {
			case metric, ok := <-p.sensorData:
				if !ok {
					return
				}

				currentWorkers := atomic.LoadInt32(&p.workersCount)
				if currentWorkers < int32(p.maxPoolSize) && metric < 0.45 {
					p.addWorker()
					fmt.Printf("adding, total workers %d\n", currentWorkers)
				}

				if currentWorkers > int32(p.minPoolSize) && metric > 0.65 {
					p.delWorker()
					fmt.Printf("delete, total workers %d\n", atomic.LoadInt32(&p.workersCount))
				}

			case <-p.quitChan:
				return
			}
		}
	}()
}

// Close chanel and wait while all gorutines are done.
func (p *Pool) CloseAndWait() {
	close(p.tasksChan)
	p.waitGroup.Wait()
}

// Submit task to worker , pass by chanel.
func (p *Pool) Exec(task Task) {
	p.tasksChan <- task
}

type SimpleTask int

func (st SimpleTask) Execute() {
	fmt.Println("executing task:", int(st)+1)
	time.Sleep(3 * time.Second)
}

func main() {
	pool := NewPool(20, 100)
	// add very simple tasks.
	for i := 0; i < 1000; i++ {
		pool.Exec(SimpleTask(i))
	}

	pool.dispatchPool()
	pool.CloseAndWait()
}
