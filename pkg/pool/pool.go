package pool

const defaultPoolSize = 10
const maxTaskLen = 10 // TODO 可配置

type WorkerPool struct {
	PoolSize int
	Workers  []*Worker
}

func NewWorkerPool(poolSize int) *WorkerPool {
	pool := &WorkerPool{
		PoolSize: poolSize,
		Workers:  make([]*Worker, poolSize),
	}
	return pool
}

func NewDefaultWorkerPool() *WorkerPool {
	return NewWorkerPool(defaultPoolSize)
}

func (pool *WorkerPool) Start() {
	for i := 0; i < pool.PoolSize; i++ {
		pool.Workers[i] = NewWorker(maxTaskLen)

		go pool.Workers[i].StartWork(i)
	}
}
