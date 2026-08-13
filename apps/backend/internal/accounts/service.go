// Package accounts contains account lifecycle use cases.
package accounts

import "context"

const purgeBatchSize = 100

// PurgeRepository removes accounts whose deletion window has expired.
type PurgeRepository interface {
	PurgeExpired(context.Context, int) (int64, error)
}

// Service coordinates account lifecycle operations.
type Service struct{ repository PurgeRepository }

// NewService builds the account service.
func NewService(repository PurgeRepository) Service { return Service{repository: repository} }

// PurgeExpired removes a bounded batch of accounts with expired deletion windows.
func (s Service) PurgeExpired(ctx context.Context) (int64, error) {
	return s.repository.PurgeExpired(ctx, purgeBatchSize)
}
