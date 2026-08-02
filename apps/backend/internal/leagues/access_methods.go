package leagues

// AccessMethods es la proyección mínima de credenciales de una cuenta autenticada.
type AccessMethods struct {
	Email, Username        string
	HasPassword, HasGoogle bool
}

// CurrentSession es la identidad y vigencia de una sesión autenticada.
type CurrentSession struct {
	AccountID, Username, IdleExpiresAt, AbsoluteExpiresAt string
}
