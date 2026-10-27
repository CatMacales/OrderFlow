package domain

//go:generate go tool go-enum -f status.go

// Status represents the order lifecycle.
// ENUM(created, confirmed, completed, cancelled)
type Status string
