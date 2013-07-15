package workout

import (
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

		t0 := time.Now()

		log.Trace("worker %d: got job %d", w.id, job.Id)

		atomic.AddUint64(&w.stat_attempt, 1)
		err = w.process(job)
		dur := time.Now().Sub(t0)

		if err != nil {
			log.Info("job %s failed in %v: %s, next attempt in %v", job.Describe(), dur, err, job.NextDelay())
			atomic.AddUint64(&w.stat_failure, 1)
			w.client.Release(job, err)
		} else {
			log.Info("job %s succeeded in %v", job.Describe(), dur)
			atomic.AddUint64(&w.stat_success, 1)
			w.client.Delete(job)
		}
	}
}

func (w *Worker) process(job *Job) (err error) {
	fn, ok := w.master.handlers[job.Tube]
	if !ok {
		return Error("no handler registered")
	}

	return fn(job)
}
