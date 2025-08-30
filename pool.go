package godoo

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type idleConn struct {
	conn net.Conn
	t    time.Time
}

type Pool struct {
	conns map[string][]idleConn
	mu    sync.Mutex

	// Maximum number of idle connections per host.
	maximalIdle int

	// Maximum number of hosts.
	maxHosts int

	// Duration after which an idle connection is closed.
	idleTimeout time.Duration

	// stop channel for closing the pool
	stop chan struct{}
}

func NewPool(maximalIdle, maxHosts int, idleTimeout time.Duration) *Pool {
	p := &Pool{
		conns:       make(map[string][]idleConn),
		maximalIdle: maximalIdle,
		maxHosts:    maxHosts,
		idleTimeout: idleTimeout,
		stop:        make(chan struct{}),
	}
	go p.run()
	return p
}

func (p *Pool) Get(ctx context.Context, host string) (net.Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns[host]) > 0 {
		// Return an idle connection if available.
		conn := p.conns[host][0].conn
		p.conns[host] = p.conns[host][1:]
		return conn, nil
	}

	if _, ok := p.conns[host]; !ok && len(p.conns) >= p.maxHosts {
		return nil, fmt.Errorf("maximum number of hosts reached")
	}

	// Create a new connection.
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *Pool) Put(host string, conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.conns[host]) >= p.maximalIdle {
		// Maximum number of idle connections reached for this host.
		conn.Close()
		return
	}

	p.conns[host] = append(p.conns[host], idleConn{conn, time.Now()})
}

func (p *Pool) CloseExpired() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for host, conns := range p.conns {
		var unexpired []idleConn
		for _, c := range conns {
			if time.Since(c.t) > p.idleTimeout {
				c.conn.Close()
			} else {
				unexpired = append(unexpired, c)
			}
		}
		p.conns[host] = unexpired
		p.conns[host] = conns
	}
}

func (p *Pool) run() {
	ticker := time.NewTicker(p.idleTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.CloseExpired()
		case <-p.stop:
			return
		}
	}
}

func (p *Pool) Close() {
	close(p.stop)
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conns := range p.conns {
		for _, c := range conns {
			c.conn.Close()
		}
	}
}
