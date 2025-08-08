package concurrent

type Semaphore struct {
	sem chan struct{}
}

func NewSemaphore(limit int) *Semaphore {
	return &Semaphore{
		sem: make(chan struct{}, limit),
	}
}

func (s *Semaphore) Acquire() {
	if s.sem == nil {
		s.sem = make(chan struct{}, 1)
	}
	s.sem <- struct{}{}
}

func (s *Semaphore) Release() {
	if s.sem == nil {
		return
	}
	<-s.sem
}
