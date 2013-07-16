package workout

import (
	"errors"
	"sync/atomic"
	"time"
)

type Worker struct {
	client       *Client
	master       *Master
	id           int
	stat_attempt uint64
	stat_success uint64
	stat_failure uint64
}

var (
	ErrJobTimeout = errors.New("job timed out")
)

func NewWorker(m *Master, wid int) (w *Worker) {
	var err error

	w = new(Worker)
	w.master = m
	w.id = wid
	w.client, err = NewClient(m.url, m.tubes)

	if m.ReserveTimeout > time.Duration(0) {
		w.client.ReserveTimeout = m.ReserveTimeout
	}

	if err != nil {
		log.Warn("worker %d: client error: %s", wid, err)
	}

	return
}

func (w *Worker) run() {
	w.master.wg.Add(1)
	defer w.master.wg.Done()

	var job *Job
	var ok bool
	var err error

	log.Debug("worker %d: starting", w.id)
	defer log.Debug("worker %d: stopped", w.id)

	for {
		select {
		case <-w.master.quit:
			log.Debug("worker %d: quitting...", w.id)
			return
		default:
		}

		if job, ok, err = w.client.Reserve(); !ok {
			continue
		}

		log.Trace("worker %d: got job %d", w.id, job.Id)

		atomic.AddUint64(&w.stat_attempt, 1)
		err = w.process(job)

		if err != nil {
			atomic.AddUint64(&w.stat_failure, 1)
			w.client.Release(job, err)
		} else {
			atomic.AddUint64(&w.stat_success, 1)
			w.client.Delete(job)
		}
	}
}

func (w *Worker) process(job *Job) (err error) {
	t0 := time.Now()

	hfn, ok := w.master.handlers[job.Tube]
	if !ok {
		return Error("no handler registered")
	}

	to, ok := w.master.timeouts[job.Tube]
	if !ok {
		to = time.Duration(12) * time.Hour
	}

	ch := make(chan error)

	go func(fn JobHandler, j *Job) {
		ch <- fn(j)
	}(hfn, job)

	select {
	case err = <-ch:
	case <-time.After(to):
		err = ErrJobTimeout
	}

	dur := time.Now().Sub(t0)

	if cfn, ok := w.master.callbacks[job.Tube]; ok {
		cfn(job, err, dur)
	}

	return
}
