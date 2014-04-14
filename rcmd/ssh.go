package rcmd

import (
	ssh "code.google.com/p/gosshnew/ssh"
	agent "code.google.com/p/gosshnew/ssh/agent"
	"errors"
	"net"
	"os"
)

type SSHSession struct {
	*ssh.Session
}
	
func GetSSHAgent() (ac agent.Agent, cleanup func(),  err error) {
	pipe := os.ExpandEnv("${SSH_AUTH_SOCK}")
	if pipe == "" {
		err = errors.New("ssh agent not running")
		return
	}

	conn, err := net.Dial("unix", pipe)
	if err != nil {
		err = errors.New("net.Dial(" + pipe + "): " + err.Error())
		return
	}

	ac = agent.NewClient(conn)
	cleanup = func() { conn.Close() }
	return

}

func NewSSHSession(user, host string) (session Session, err error) {
	agent, cleanup, err := GetSSHAgent()
	if err != nil {
		return
	}
	defer cleanup()

	config := &ssh.ClientConfig{User: user}
        config.Auth = append(config.Auth, ssh.PublicKeysCallback(agent.Signers))

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return
	}

	ss, err := client.NewSession()
	if err != nil {
		return
	}

	session = &SSHSession{ss}

	return
}

func (ss *SSHSession) Exec(cmd string) ([]byte, error) {
	return ss.CombinedOutput(cmd)
}