package util

import (
	"fmt"
	"os"
)

func Fatal(m interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", m)
	os.Exit(1)
}

func Fatalf(f string, args ...interface{}) {
	s := fmt.Sprintf(f, args...)
	fmt.Fprintf(os.Stderr, "%s\n", s)
	os.Exit(1)
}