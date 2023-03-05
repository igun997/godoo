package godoo

import (
	"context"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	// Create a new connection pool.
	pool := NewPool(2, 2, 5*time.Second)

	// Test creating a new connection.
	conn1, err := pool.Get(context.Background(), "localhost:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn1.Close()
	pool.Put("localhost:8090", conn1)

	// Test getting an idle connection.
	conn2, err := pool.Get(context.Background(), "localhost:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn2.Close()

	if conn1 != conn2 {
		t.Errorf("Expected conn1 and conn2 to be the same, but got different connections")
	}

	// Test maximum number of idle connections per host.
	conn3, err := pool.Get(context.Background(), "localhost:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn3.Close()

	conn4, err := pool.Get(context.Background(), "localhost:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn4.Close()

	conn5, err := pool.Get(context.Background(), "localhost:8090")
	if err == nil {
		conn5.Close()
		t.Fatalf("Expected error when getting connection, but got nil")
	}

	// Test maximum number of hosts.
	conn6, err := pool.Get(context.Background(), "127.0.0.1:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn6.Close()

	conn7, err := pool.Get(context.Background(), "localhost:8081")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn7.Close()

	conn8, err := pool.Get(context.Background(), "example.com:80")
	if err == nil {
		conn8.Close()
		t.Fatalf("Expected error when getting connection, but got nil")
	}

	// Test idle timeout.
	time.Sleep(6 * time.Second)
	pool.CloseExpired()

	conn9, err := pool.Get(context.Background(), "localhost:8090")
	if err != nil {
		t.Fatalf("Failed to get connection: %s", err)
	}
	defer conn9.Close()

	if conn9 == conn1 {
		t.Errorf("Expected conn9 to be a new connection, but got an idle connection")
	}
}
