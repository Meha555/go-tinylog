# go-tinylog

A tiny log library for Go.

It encapsulates the log standard library and supports features like log levels, call stacks, goroutine IDs, terminal colors, and other convenient interfaces.

> Extracted from [pulse](https://github.com/Meha555/pulse.git).

## Installation

To install, use the go get command:

```shell
go get github.com/Meha555/go-tinylog
```

## Example

```go
stdLogger, err := tinylog.NewStdLogger(tinylog.LevelDebug, "STD_LOG_TEST", "[%t] [%c %l] [%f:%C:%L:%g] %m", false, tinylog.Lcolored|tinylog.Lstructured)
if err != nil {
    panic(err)
}
stdLogger.Debug("This is a debug message")

fileLogger, err := tinylog.NewFileLogger(tinylog.LevelDebug, "FILE_LOG_TEST", "[%t] [%c %l] [%f:%C:%L:%g] %m", "./log", "test.log", 1024*1024, true, tinylog.Lstructured)
if err != nil {
    panic(err)
}
fileLogger.Info("This is an info message for file log")
fileLogger.Error("This is an error message for file log")
```