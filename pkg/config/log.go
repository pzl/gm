package config

type ctxKey int

const (
	LogKey ctxKey = iota
	DbusKey
	FileKey
)
