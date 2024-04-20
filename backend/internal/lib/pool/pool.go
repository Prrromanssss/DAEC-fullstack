package pool

import (
	"fmt"
	"sync"
)

type WorkerPool interface {
	Start()
	Stop()
	AddWork(PoolTask)
}

type PoolTask interface {
	Execute() error
	OnFailure(error)
}

type MyPool struct {
	tasks       chan PoolTask
	wg          sync.WaitGroup
	isExecuting bool
	onceStart   sync.Once
	onceStop    sync.Once
	numWorkers  int
}

func NewWorkerPool(numWorkers int, channelSize int) (*MyPool, error) {
	if numWorkers <= 0 {
		return nil, fmt.Errorf("incorect numWorkers")
	}
	if channelSize < 0 {
		return nil, fmt.Errorf("negative channelSize")
	}
	return &MyPool{
		tasks:       make(chan PoolTask, channelSize),
		isExecuting: false,
		numWorkers:  numWorkers,
	}, nil
}

func (mp *MyPool) Start() {
	mp.onceStart.Do(func() {
		mp.wg.Add(mp.numWorkers)
		for i := 0; i < mp.numWorkers; i++ {
			go func() {
				defer mp.wg.Done()
				for pt := range mp.tasks {
					err := pt.Execute()
					if err != nil {
						pt.OnFailure(err)
					}
				}
			}()
		}
		mp.isExecuting = true
	})
}

func (mp *MyPool) Stop() {
	mp.onceStop.Do(func() {
		mp.isExecuting = false
		close(mp.tasks)
		mp.wg.Wait()
	})

}

func (mp *MyPool) AddWork(pt PoolTask) {
	if mp.isExecuting {
		mp.tasks <- pt
	}
}
