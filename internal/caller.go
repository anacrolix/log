package internal

// Runs a function with a frame from inside this package as a caller.
func Run(f func()) {
	f()
}
