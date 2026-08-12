package accounts

import (
	"context"
	"errors"
	"testing"
)

type purgeRepositoryStub struct {
	deleted int64
	limit   int
	err     error
}

func (s *purgeRepositoryStub) PurgeExpired(_ context.Context, limit int) (int64, error) {
	s.limit = limit
	return s.deleted, s.err
}

func TestPurgeExpiredUsesBoundedBatch(t *testing.T) {
	t.Parallel()
	repository := &purgeRepositoryStub{deleted: 2}
	deleted, err := NewService(repository).PurgeExpired(context.Background())
	if err != nil {
		t.Fatalf("PurgeExpired() error = %v", err)
	}
	if deleted != 2 {
		t.Fatalf("PurgeExpired() = %d, want 2", deleted)
	}
	if repository.limit != purgeBatchSize {
		t.Fatalf("batch size = %d, want %d", repository.limit, purgeBatchSize)
	}
}

func TestPurgeExpiredPropagatesRepositoryError(t *testing.T) {
	t.Parallel()
	want := errors.New("database unavailable")
	_, err := NewService(&purgeRepositoryStub{err: want}).PurgeExpired(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("PurgeExpired() error = %v, want %v", err, want)
	}
}
