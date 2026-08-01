package leagues

// AccessMethods es la proyección mínima de credenciales de una cuenta autenticada.
type AccessMethods struct {
	Email, Username        string
	HasPassword, HasGoogle bool
}
