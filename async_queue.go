package zhanio

import "sync"

type AsyncQueue struct {
	locker sync.Locker
	jobs   []func() error
}

func (q *AsyncQueue) Push(job func() error) {
	q.locker.Lock()
	q.jobs = append(q.jobs, job)
	q.locker.Unlock()
}

func (q *AsyncQueue) ForEach(iter func(job func() error) error) error {
	q.locker.Lock()
	jobs := q.jobs
	q.jobs = nil
	q.locker.Unlock()
	for _, job := range jobs {
		if err := iter(job); err != nil {
			return err
		}
	}
	return nil
}
