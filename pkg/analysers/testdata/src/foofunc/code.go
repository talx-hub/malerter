package main

import "os"

// No error should be reported here because there is no main func
func foo() {
	os.Exit(1)
}
