package utils

// MapPtr returns a pointer to the value passed in. 
// It is useful for taking the address of a returned value or constant.
func MapPtr[T any](v T) *T {
	return &v
}
