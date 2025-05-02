package internal

import "fmt"

// KeyPrefix adds prefix to key
type KeyPrefix interface {
	Prefix(string) string
}

// WithPrefix add name as prefix to key
type WithPrefix struct {
	Name string
}

// Prefix returns key with prefix
func (wp *WithPrefix) Prefix(key string) string {
	return fmt.Sprintf("%s.%s", wp.Name, key)
}

// NoPrefix adds no prefix
type NoPrefix struct {
}

// Prefix returns original key
func (np *NoPrefix) Prefix(key string) string {
	return key
}
