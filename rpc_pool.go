package godoo

import (
	"fmt"
	"sync"
	"time"

	"github.com/wongpinter/godoo/logr" // Assuming a local logr package
)

// ConnectionPool manages a collection of godoo clients.
type ConnectionPool struct {
	clients       map[string]*Client
	mutex         sync.RWMutex
	maxRetries    int
	retryInterval time.Duration
}

// NewConnectionPool creates and initializes a new connection pool.
// It attempts to connect to all servers defined in serverConfigs and retries failed connections in the background.
func NewConnectionPool(retryInterval time.Duration, maxRetries int, serverConfigs map[string]*ClientConfig) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		clients:       make(map[string]*Client),
		maxRetries:    maxRetries,
		retryInterval: retryInterval,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to protect access to failedClients map

	failedClients := make(map[string]*ClientConfig)

	for key, cfg := range serverConfigs {
		client, err := NewClient(cfg)
		if err != nil {
			logr.Warnf("Connecting to XML-RPC %s... Failed with error: %v", cfg.URL, err)
			failedClients[key] = cfg
			continue
		}
		logr.Infof("Connecting to XML-RPC %s... Success", cfg.URL)
		pool.clients[key] = client
	}

	// Retry failed connections in the background.
	if len(failedClients) > 0 {
		wg.Add(len(failedClients))
		for key, cfg := range failedClients {
			go func(k string, c *ClientConfig) {
				defer wg.Done()
				client := pool.retryConnection(c)
				if client != nil {
					mu.Lock()
					pool.clients[k] = client
					mu.Unlock()
				}
			}(key, cfg)
		}
	}

	return pool, nil
}

// retryConnection attempts to establish a connection with retries.
func (p *ConnectionPool) retryConnection(cfg *ClientConfig) *Client {
	for i := 1; i <= p.maxRetries; i++ {
		client, err := NewClient(cfg)
		if err == nil {
			logr.Infof("Successfully reconnected to %s", cfg.URL)
			return client
		}
		logr.Warnf("Retry %d/%d for %s failed: %v", i, p.maxRetries, cfg.URL, err)
		time.Sleep(p.retryInterval)
	}
	logr.Warnf("Failed to create client for URL %s after %d retries", cfg.URL, p.maxRetries)
	return nil
}

// GetClient retrieves a client from the pool by its key.
func (p *ConnectionPool) GetClient(key string) (*Client, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	client, ok := p.clients[key]
	if !ok {
		return nil, fmt.Errorf("client with key %s does not exist", key)
	}

	return client, nil
}

// NumConnections returns the number of active connections in the pool.
func (p *ConnectionPool) NumConnections() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.clients)
}

// Close gracefully closes all client connections in the pool.
func (p *ConnectionPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, client := range p.clients {
		client.Close()
	}
	p.clients = make(map[string]*Client) // Clear the map
}

// RemoveClient closes and removes a client from the pool.
func (p *ConnectionPool) RemoveClient(key string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if client, ok := p.clients[key]; ok {
		client.Close()
		delete(p.clients, key)
	}
}

// AddClient creates a new client and adds it to the pool.
func (p *ConnectionPool) AddClient(key string, cfg ClientConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, ok := p.clients[key]; ok {
		return fmt.Errorf("client with key %s already exists", key)
	}
	client, err := NewClient(&cfg)
	if err != nil {
		return err
	}
	p.clients[key] = client
	return nil
}

// CheckConnections verifies all active connections and attempts to reconnect any that have failed.
func (p *ConnectionPool) CheckConnections() {
	p.mutex.RLock()
	disconnected := make(map[string]*ClientConfig)
	for key, client := range p.clients {
		if err := client.Ping(); err != nil {
			logr.Warnf("Connection for %s is disconnected: %v", key, err)
			disconnected[key] = client.Config()
		}
	}
	p.mutex.RUnlock()

	if len(disconnected) == 0 {
		return
	}

	// Remove disconnected clients first
	p.mutex.Lock()
	for key := range disconnected {
		if client, ok := p.clients[key]; ok {
			client.Close()
			delete(p.clients, key)
		}
	}
	p.mutex.Unlock()

	// Retry connections in the background
	var wg sync.WaitGroup
	wg.Add(len(disconnected))
	for key, cfg := range disconnected {
		go func(k string, c *ClientConfig) {
			defer wg.Done()
			if client := p.retryConnection(c); client != nil {
				p.mutex.Lock()
				p.clients[k] = client
				p.mutex.Unlock()
			}
		}(key, cfg)
	}
}