package test

import "os"

// No error should be reported here because there is no main package
func main() {
	os.Exit(1)
}
