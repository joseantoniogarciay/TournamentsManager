package leagues

// AccessMethods is the minimal credential projection of an authenticated account.
type AccessMethods struct {
	Email, Username        string
	HasPassword, HasGoogle bool
}

// CurrentSession is the identity and validity of an authenticated session.
type CurrentSession struct {
	AccountID, Username, IdleExpiresAt, AbsoluteExpiresAt string
}
