package suggestions

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	item Item
	body string
	err  error
}

func (r *repositoryStub) CreateSuggestion(_ context.Context, _ string, body string) (Item, error) {
	r.body = body
	return r.item, r.err
}

type notifierStub struct {
	item Item
	err  error
}

func (n *notifierStub) NotifySuggestion(_ context.Context, item Item) error {
	n.item = item
	return n.err
}

func TestSubmitTrimsPersistsAndNotifies(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{item: Item{ID: "suggestion", Username: "person", Body: "Una mejora", CreatedAt: "2026-09-13T10:00:00Z"}}
	notifier := &notifierStub{}
	result, err := NewService(repository, notifier).Submit(context.Background(), "account", "  Una mejora  ")
	if err != nil || result.NotificationFailed {
		t.Fatalf("Submit() = %#v, %v", result, err)
	}
	if repository.body != "Una mejora" || notifier.item.ID != repository.item.ID {
		t.Fatalf("persisted = %q, notified = %#v", repository.body, notifier.item)
	}
}

func TestSubmitRejectsInvalidUnicodeLength(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{}
	if _, err := NewService(repository, nil).Submit(context.Background(), "account", " siete "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Submit() error = %v, want %v", err, ErrInvalidInput)
	}
	if repository.body != "" {
		t.Fatal("invalid suggestion reached repository")
	}
}

func TestSubmitKeepsDurableSuccessWhenNotificationFails(t *testing.T) {
	t.Parallel()
	repository := &repositoryStub{item: Item{Body: "Una mejora"}}
	result, err := NewService(repository, &notifierStub{err: errors.New("smtp unavailable")}).Submit(context.Background(), "account", "Una mejora")
	if err != nil || !result.NotificationFailed {
		t.Fatalf("Submit() = %#v, %v; want durable success with notification failure", result, err)
	}
}
