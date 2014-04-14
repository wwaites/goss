package log

import (
	"log"
	"os"
)

var Emerg *log.Logger
var Alert *log.Logger
var Crit *log.Logger
var Err *log.Logger
var Warn *log.Logger
var Notice *log.Logger
var Info *log.Logger
var Debug *log.Logger

func init() {
	flags := log.Ldate | log.Ltime | log.Lshortfile

	Emerg = log.New(os.Stderr, "[emerg] ", flags) 
	Alert = log.New(os.Stderr, "[alert] ", flags) 
	Crit = log.New(os.Stderr, "[crit] ", flags) 
	Err = log.New(os.Stderr, "[err] ", flags) 
	Warn = log.New(os.Stderr, "[warn] ", flags) 
	Notice = log.New(os.Stderr, "[notice] ", flags) 
	Info = log.New(os.Stderr, "[info] ", flags) 
	Debug = log.New(os.Stderr, "[debug] ", flags)

	log.SetFlags(flags)
	log.SetPrefix("[misc] ")
}
