package logging

const (
	Lplain   = 0
	Lcolored = 1 << iota
	Lstructured
)

type ILogSinker interface {
	Sink(msg *LogMsg) error
	Flags() int
}
