package pool

import "fmt"

// 优化goroutine，可以理解成goroutine池

type Worker struct {
	Id        int
	taskQueue chan Task
}

func NewWorker(maxTaskLen int) *Worker {
	return &Worker{
		taskQueue: make(chan Task, maxTaskLen),
	}
}

func (w *Worker) StartWork(workerId int) {
	fmt.Println("Worker.StartWork | workerId: ", workerId)
	w.Id = workerId

	for {
		select {
		case task := <-w.taskQueue:
			task()
		}
	}
}

func (w *Worker) Enqueue(task Task) {
	fmt.Println("Worker.Enqueue | workerId: ", w.Id)
	w.taskQueue <- task
}
