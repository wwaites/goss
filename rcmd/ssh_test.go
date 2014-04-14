package rcmd

import (
	"testing"
	"os"
)

func TestSSH(t *testing.T) {
	user := os.ExpandEnv("${USER}")
	session, err := NewSSHSession(user, "localhost:22")
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	output, err := session.Exec("id")
	if err != nil {
		t.Fatal(err)
	}

	if len(output) == 0 {
		t.Fatal("should have got some output from ssh command")
	}
}
