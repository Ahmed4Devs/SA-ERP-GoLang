package ports

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (bool, error)
}

type TokenGenerator interface {
	GenerateSessionToken() (string, string, error)
}
