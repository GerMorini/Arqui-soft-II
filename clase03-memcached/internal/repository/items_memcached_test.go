package repository

import (
	"clase03-memcached/internal/domain"
	"context"
	"net"
	"os"
	"testing"
	"time"
)

func isMemcachedUp(host string, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func TestMemcachedItemsRepository_CRUD(t *testing.T) {
	host := os.Getenv("MEMCACHED_HOST")
	if host == "" {
		host = "memcached"
	}
	port := os.Getenv("MEMCACHED_PORT")
	if port == "" {
		port = "11211"
	}

	if !isMemcachedUp(host, port) {
		t.Fatalf("memcached not reachable at %s:%s", host, port)
	}

	repo := NewMemcachedItemsRepository(host, port, 2*time.Second)
	ctx := context.Background()

	item := domain.Item{ID: "mc-1", Name: "Test MC", Price: 10}

	// Create
	if _, err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// GetByID
	got, err := repo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got.Name != item.Name || got.Price != item.Price {
		t.Fatalf("unexpected item: %+v", got)
	}

	// Update
	updated := domain.Item{Name: "Updated MC", Price: 20}
	got, err = repo.Update(ctx, item.ID, updated)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if got.ID != item.ID || got.Name != "Updated MC" || got.Price != 20 {
		t.Fatalf("unexpected updated item: %+v", got)
	}

	// Read after update
	got, err = repo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID after update error: %v", err)
	}
	if got.Name != "Updated MC" || got.Price != 20 {
		t.Fatalf("unexpected item after update: %+v", got)
	}

	// Delete
	if err := repo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	// Ensure deleted
	if _, err := repo.GetByID(ctx, item.ID); err == nil {
		t.Fatalf("expected error getting deleted key, got nil")
	}
}
