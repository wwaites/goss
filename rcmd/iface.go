package rcmd

type Session interface {
	Exec(string) ([]byte, error)
	Close() error
}