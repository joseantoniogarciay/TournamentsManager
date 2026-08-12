// Package accounts contiene los casos de uso del ciclo de vida de una cuenta.
package accounts

import "context"

const purgeBatchSize = 100

// PurgeRepository elimina cuentas cuya ventana de baja ya ha vencido.
type PurgeRepository interface {
	PurgeExpired(context.Context, int) (int64, error)
}

// Service coordina operaciones del ciclo de vida de cuentas.
type Service struct{ repository PurgeRepository }

// NewService construye el servicio de cuentas.
func NewService(repository PurgeRepository) Service { return Service{repository: repository} }

// PurgeExpired elimina un lote acotado de cuentas con baja vencida.
func (s Service) PurgeExpired(ctx context.Context) (int64, error) {
	return s.repository.PurgeExpired(ctx, purgeBatchSize)
}
