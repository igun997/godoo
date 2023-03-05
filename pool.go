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
}

func NewPool(maximalIdle, maxHosts int, idleTimeout time.Duration) *Pool {
	return &Pool{
		conns:       make(map[string][]idleConn),
		maximalIdle: maximalIdle,
		maxHosts:    maxHosts,
		idleTimeout: idleTimeout,
	}
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

	if len(p.conns) >= p.maxHosts {
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
		for i := 0; i < len(conns); {
			if time.Since(conns[i].t) > p.idleTimeout {
				conns[i].conn.Close()
				copy(conns[i:], conns[i+1:])
				conns = conns[:len(conns)-1]
			} else {
				i++
			}
		}
		p.conns[host] = conns
	}
}
