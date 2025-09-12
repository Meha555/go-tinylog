package tinylog

import (
	"sync"
	"testing"
)

func TestLogConcurrency(t *testing.T) {
	logger1, _ := NewStdLogger(LevelInfo, "concurrency", "[%t] [%c %l] [%f:%C:%L:%g] %m", false, Lcolored|Lstructured)
	logger2, _ := NewStdLogger(LevelInfo, "concurrency", "[%t] [%c %l] [%f:%C:%L:%g] %m", true, Lcolored|Lstructured)
	const goroutineNum = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutineNum; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i&1 == 1 {
				for j := 0; j < goroutineNum; j++ {
					logger1.Debugf("This is a debug message from goroutine[%d]", i)
					logger1.Infof("This is an info message from goroutine[%d]", i)
					logger1.Warnf("This is a warning message from goroutine[%d]", i)
					logger1.Errorf("This is an error message from goroutine[%d]", i)
				}
			} else {
				logger2.Debugf("This is a debug message from goroutine[%d]", i)
				logger2.Infof("This is an info message from goroutine[%d]", i)
				logger2.Warnf("This is a warning message from goroutine[%d]", i)
				logger2.Errorf("This is an error message from goroutine[%d]", i)
			}
		}(i)
	}
	wg.Wait()
	t.Log("done")
}

func TestStdLog(t *testing.T) {
	stdLogger, _ := NewStdLogger(LevelInfo, "STD_LOG_TEST", "[%t] [%c %l] [%f:%C:%L:%g] %m", false, Lcolored)
	t.Run("Print", func(t *testing.T) {
		stdLogger.Debug("This is a debug message")
		stdLogger.Info("This is an info message")
		stdLogger.Warn("This is a warning message")
		stdLogger.Error("This is an error message")

		stdLogger.Debugf("This is a debug message of %s", LevelStrs[LevelDebug])
		stdLogger.Infof("This is an info message of %s", LevelStrs[LevelInfo])
		stdLogger.Warnf("This is a warning message of %s", LevelStrs[LevelWarn])
		stdLogger.Errorf("This is an error message of %s", LevelStrs[LevelError])
	})
	t.Run("Level", func(t *testing.T) {
		// 测试不同日志级别的过滤
		stdLogger.SetLevel(LevelInfo)
		stdLogger.Debug("This debug message should be filtered")
		stdLogger.Info("This is an info message")
		stdLogger.Warn("This is a warning message")
		stdLogger.Error("This is an error message")

		stdLogger.SetLevel(LevelError)
		stdLogger.Debug("This debug message should be filtered")
		stdLogger.Info("This info message should be filtered")
		stdLogger.Warn("This warning message should be filtered")
		stdLogger.Error("This is an error message")
	})
	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Logf("Panic recovered: %v", err)
			} else {
				t.Errorf("Panic not recovered")
			}
		}()
		stdLogger.Panic("This is a panic message")
	})
}

func TestFileLog(t *testing.T) {
	fileLogger, _ := NewFileLogger(LevelDebug, "FILE_LOG_TEST", "[%t] [%c %l] [%f:%C:%L:%g] %m", "./log", "test.log", 1024*1024, false, Lstructured)

	t.Run("Print", func(t *testing.T) {
		fileLogger.Debug("This is a debug message for file log")
		fileLogger.Info("This is an info message for file log")
		fileLogger.Warn("This is a warning message for file log")
		fileLogger.Error("This is an error message for file log")

		fileLogger.Debugf("This is a debug message of %s for file log", LevelStrs[LevelDebug])
		fileLogger.Infof("This is an info message of %s for file log", LevelStrs[LevelInfo])
		fileLogger.Warnf("This is a warning message of %s for file log", LevelStrs[LevelWarn])
		fileLogger.Errorf("This is an error message of %s for file log", LevelStrs[LevelError])
	})
	t.Run("Level", func(t *testing.T) {
		fileLogger.SetLevel(LevelInfo)
		fileLogger.Debug("This debug message for file log should be filtered")
		fileLogger.Info("This is an info message for file log")
		fileLogger.Warn("This is a warning message for file log")
		fileLogger.Error("This is an error message for file log")

		fileLogger.SetLevel(LevelError)
		fileLogger.Debug("This debug message for file log should be filtered")
		fileLogger.Info("This info message for file log should be filtered")
		fileLogger.Warn("This warning message for file log should be filtered")
		fileLogger.Error("This is an error message for file log")
	})
	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Logf("Panic recovered: %v", err)
			} else {
				t.Errorf("Panic not recovered")
			}
		}()
		fileLogger.Panic("This is a panic message")
	})
}
