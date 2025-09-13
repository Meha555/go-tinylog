package tinylog

const (
	Lplain   = 0
	Lcolored = 1 << iota
	Lstructured
)

type LogSinker interface {
	Sink(msg *logMsg) error
}
