package workout

import (
	"sync"
	"time"
)

type JobHandler func(*Job) error

type Master struct {
	ReserveTimeout time.Duration
	concurrency    int
	url            string
	tubes          []string
	workers        []*Worker
	handlers       map[string]JobHandler
	job            chan *Job
	quit           chan bool
	mg             sync.WaitGroup
	wg             sync.WaitGroup
}

func NewMaster(url string, tubes []string, concurrency int) *Master {
	return &Master{
		url:         url,
		tubes:       tubes,
		concurrency: concurrency,
		handlers:    make(map[string]JobHandler),
	}
}

func (m *Master) RegisterHandler(name string, fn JobHandler) {
	m.handlers[name] = fn
	return
}

func (m *Master) Start() (err error) {
	m.mg.Add(1)

	m.job = make(chan *Job, 2)
	m.quit = make(chan bool, m.concurrency)

	m.workers = make([]*Worker, m.concurrency)

	log.Info("master: starting %d workers...", m.concurrency)

	for i := 0; i < m.concurrency; i++ {
		m.workers[i] = NewWorker(m, i)
		go m.workers[i].run()
	}

	log.Info("master: ready")

	return
}

func (m *Master) Stop() (err error) {
	log.Info("master: stopping %d workers...", m.concurrency)

	for i := 0; i < m.concurrency; i++ {
		m.quit <- true
	}

	m.wg.Wait()
	log.Info("master: %d workers stopped", m.concurrency)

	m.mg.Done()
	log.Info("master: stopped")

	return
}

func (m *Master) Wait() {
	m.mg.Wait()
}
