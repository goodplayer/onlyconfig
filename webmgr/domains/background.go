package domains

import "log"

type Task struct {
	resultQueue chan any
	fn          func(result chan<- any)
}

func newFnTask(fn func(result chan<- any)) *Task {
	return &Task{
		resultQueue: make(chan any, 1),
		fn:          fn,
	}
}

type Worker struct {
	taskQueue chan *Task
}

func NewWorker() *Worker {
	w := &Worker{
		taskQueue: make(chan *Task, 1024),
	}
	w.start()
	return w
}

func (w *Worker) start() {
	go func() {
		for {
			task := <-w.taskQueue
			f := func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("recover from panic while processing task in worker:", r)
					}
				}()
				task.fn(task.resultQueue)
			}
			f()
		}
	}()
}

func (w *Worker) SubmitFnTask(fn func(result chan<- any)) <-chan any {
	t := newFnTask(fn)
	w.taskQueue <- t
	return t.resultQueue
}
