package model

type Person struct {
	ID        int64
	UUID      string
	Email     string
	Username  string
	CreatedAt int64
	UpdatedAt *int64
	IsActive  bool
	IsDefault bool
}

