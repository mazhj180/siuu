package mux

import (
	"context"
	"errors"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type SessionState int32

const (
	Active SessionState = iota
	Busy
	Dead
)

type poolSession struct {
	Session
	state     int32
	createdAt time.Time
	lastUsed  int64 // unix timestamp
}

func (ps *poolSession) getState() SessionState {
	return SessionState(atomic.LoadInt32(&ps.state))
}

func (ps *poolSession) setState(state SessionState) {
	atomic.StoreInt32(&ps.state, int32(state))
}

func (ps *poolSession) updateLastUsed() {
	atomic.StoreInt64(&ps.lastUsed, time.Now().Unix())
}

func (ps *poolSession) getLastUsed() time.Time {
	return time.Unix(atomic.LoadInt64(&ps.lastUsed), 0)
}

func (ps *poolSession) IsHealthy() bool {
	return ps.Session != nil && !ps.IsClosed() && ps.getState() == Active
}

func (ps *poolSession) IsDead() bool {
	return !ps.IsHealthy() && ps.getState() == Dead
}

type Pool struct {
	sessions []*poolSession
	maxSize  int
	mu       sync.RWMutex
	ins      Interface
	dialer   func(ctx context.Context) (net.Conn, error)

	closed int32
	stats  PoolStats
}

type PoolStats struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	SessionCreated  int64
	SessionCleaned  int64
}

func NewPool(ins Interface, dialer func(ctx context.Context) (net.Conn, error), maxSize int) *Pool {
	p := &Pool{
		sessions: make([]*poolSession, 0, maxSize),
		maxSize:  maxSize,
		ins:      ins,
		dialer:   dialer,
	}

	return p
}

func (p *Pool) GetStream(ctx context.Context) (Stream, error) {
	atomic.AddInt64(&p.stats.TotalRequests, 1)

	// Try to get a stream from healthy session
	if stream, err := p.tryGetStreamFromHealthySessions(); err == nil {
		atomic.AddInt64(&p.stats.SuccessRequests, 1)
		return stream, nil
	}

	// try to create a new session and get a stream
	if stream, err := p.tryCreateNewSessionIfNeeded(ctx); err == nil {
		atomic.AddInt64(&p.stats.SuccessRequests, 1)
		return stream, nil
	}

	atomic.AddInt64(&p.stats.FailedRequests, 1)
	return nil, errors.New("no available session and max session count reached")
}

func (p *Pool) tryGetStreamFromHealthySessions() (Stream, error) {
	p.mu.RLock()
	sessions := make([]*poolSession, len(p.sessions))
	copy(sessions, p.sessions)
	p.mu.RUnlock()

	if len(sessions) == 0 {
		return nil, errors.New("no sessions available")
	}

	// only consider healthy sessions, sort by load
	var healthySessions []*poolSession

	for _, ps := range sessions {
		if ps != nil && ps.IsHealthy() {
			healthySessions = append(healthySessions, ps)
		}
	}

	if len(healthySessions) == 0 {
		p.cleanupUnhealthySession()
		return nil, errors.New("no healthy sessions available")
	}

	sort.Slice(healthySessions, func(i, j int) bool {
		return healthySessions[i].NumStreams() < healthySessions[j].NumStreams()
	})

	for _, ps := range healthySessions {
		stream, err := ps.OpenStream()
		if err == nil {
			ps.updateLastUsed()
			return stream, nil
		}

		if errors.Is(err, FataError{}) {
			ps.setState(Dead)
		} else {
			ps.setState(Busy)
		}
	}

	return nil, errors.New("failed to get stream from healthy sessions")
}

func (p *Pool) tryCreateNewSessionIfNeeded(ctx context.Context) (Stream, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	healthyCount := 0
	for _, ps := range p.sessions {
		if ps != nil && ps.IsHealthy() {
			healthyCount++
		}
	}

	// If there is a healthy session, it indicates that the issue may be temporary, and no new connection should be created.
	if healthyCount > 0 {
		return nil, errors.New("healthy sessions exist but failed to get stream")
	}

	// If the maximum number has been reached, no new ones can be created
	if len(p.sessions) >= p.maxSize {
		return nil, errors.New("max session count reached and no healthy sessions")
	}

	ps, err := p.createSession(ctx)
	if err != nil {
		return nil, err
	}

	p.sessions = append(p.sessions, ps)
	atomic.AddInt64(&p.stats.SessionCreated, 1)

	stream, err := ps.OpenStream()
	if err != nil {
		// create failed, remove this session
		p.sessions = p.sessions[:len(p.sessions)-1]
		ps.setState(Dead)
		if ps.Session != nil {
			_ = ps.Session.Close()
		}
		return nil, err
	}

	ps.updateLastUsed()
	return stream, nil
}

func (p *Pool) cleanupUnhealthySession() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var livingSessions []*poolSession
	var cleanedCount int

	for _, ps := range p.sessions {
		if ps != nil && !ps.IsDead() {
			livingSessions = append(livingSessions, ps)
		} else {
			if ps != nil && ps.Session != nil {
				_ = ps.Close()
			}
			cleanedCount++
		}
	}

	if cleanedCount == 0 {
		return
	}

	p.sessions = livingSessions
	atomic.AddInt64(&p.stats.SessionCleaned, int64(cleanedCount))
}

func (p *Pool) Close() error {
	atomic.StoreInt32(&p.closed, 1)

	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for _, ps := range p.sessions {
		if ps != nil && ps.Session != nil {
			if err := ps.Close(); err != nil {
				lastErr = err
			}
		}
	}

	p.sessions = nil
	return lastErr
}

func (p *Pool) createSession(ctx context.Context) (*poolSession, error) {
	conn, err := p.dialer(ctx)
	if err != nil {
		return nil, err
	}

	session, err := p.ins.Client(conn)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &poolSession{
		Session:   session,
		state:     int32(Active),
		createdAt: now,
		lastUsed:  now.Unix(),
	}, nil
}

// GetStats Getting Connection Pool Statistics
func (p *Pool) GetStats() PoolStats {
	return PoolStats{
		TotalRequests:   atomic.LoadInt64(&p.stats.TotalRequests),
		SuccessRequests: atomic.LoadInt64(&p.stats.SuccessRequests),
		FailedRequests:  atomic.LoadInt64(&p.stats.FailedRequests),
		SessionCreated:  atomic.LoadInt64(&p.stats.SessionCreated),
		SessionCleaned:  atomic.LoadInt64(&p.stats.SessionCleaned),
	}
}

// HealthCheck Check the health status of the connection pool
func (p *Pool) HealthCheck() (healthy, total int) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total = len(p.sessions)
	for _, ps := range p.sessions {
		if ps != nil && ps.IsHealthy() {
			healthy++
		}
	}
	return
}
