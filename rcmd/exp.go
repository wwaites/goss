package rcmd

import (
	"fmt"
	"hubs.net.uk/foreign/gexpect"
	"regexp"
)

type ExpSession struct {
	*gexpect.ExpectSubprocess
	promptRe *regexp.Regexp
}

func NewExpSession(user, host, prompt string) (session Session, err error) {
	promptRe, err := regexp.Compile(prompt)
	if err != nil {
		return
	}

	cmd := fmt.Sprintf("ssh -t -l %s %s", user, host)

	es, err := gexpect.Spawn(cmd)
	if err != nil {
		return
	}

	_, err = es.ExpectRegex(promptRe)
	if err != nil {
		es.Close()
		return
	}

	session = &ExpSession{es, promptRe}
	return
}

func (es *ExpSession) Exec(s string) (buf []byte, err error) {
	err = es.SendLine(s)
	if err != nil {
		return
	}
	_, err = es.ReadLine()
	if err != nil {
		return
	}

	buf, err = es.ExpectRegex(es.promptRe)
	return
} 
