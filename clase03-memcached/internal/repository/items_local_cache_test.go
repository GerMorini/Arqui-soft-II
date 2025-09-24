package repository

import (
	"clase03-memcached/internal/domain"
	"context"
	"testing"
	"time"
)

func TestItemsLocalCacheRepository_CRUD(t *testing.T) {
	repo := NewItemsLocalCacheRepository(2 * time.Second)
	ctx := context.Background()

	item := domain.Item{ID: "local-1", Name: "Test Local", Price: 10}

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
	updated := domain.Item{Name: "Updated Local", Price: 20}
	got, err = repo.Update(ctx, item.ID, updated)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if got.ID != item.ID || got.Name != "Updated Local" || got.Price != 20 {
		t.Fatalf("unexpected updated item: %+v", got)
	}

	// Read after update
	got, err = repo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("GetByID after update error: %v", err)
	}
	if got.Name != "Updated Local" || got.Price != 20 {
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
