package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type stdSinker struct {
	formater *LogFormater
	flags    int
	// mtx      sync.Mutex // 不需要，(os.File).Write底层是加了锁的
}

func NewStdSinker(format string, flags int) (*stdSinker, error) {
	f, err := NewLogFormatter(format)
	if err != nil {
		return nil, err
	}
	s := &stdSinker{
		formater: f,
		flags:    flags,
	}
	return s, nil
}

const (
	// 重置所有颜色设置，恢复默认颜色
	ColorReset = "\033[0m"
	// 前景色为红色
	ColorRed = "\033[31m"
	// 前景色为绿色
	ColorGreen = "\033[32m"
	// 前景色为黄色
	ColorYellow = "\033[33m"
	// 前景色为蓝色
	ColorBlue = "\033[34m"
	// 前景色为灰色
	ColorGray = "\033[90m"
	// 前景色为高亮白色，背景色为红色
	ColorHiRed = "\033[97;41m"
	// 前景色为高亮黄色，背景色为红色
	ColorHiYellowOnRed = "\033[93;41m"
)

func (s *stdSinker) Flags() int {
	return s.flags
}

func (s *stdSinker) Sink(msg *LogMsg) (err error) {
	var builder strings.Builder
	var logStr, finalLogStr string

	if s.flags&Lstructured != 0 {
		// 填充编码为JSON的逻辑，不要indent
		jsonBytes, err := json.Marshal(*msg)
		if err != nil {
			// 处理JSON序列化错误，可以返回错误或使用默认字符串
			logStr = fmt.Sprintf("failed to marshal log message: %v", err)
		} else {
			logStr = string(jsonBytes)
		}
	} else {
		logStr = s.formater.Format(msg)
	}

	if s.flags&Lcolored != 0 {
		switch msg.Level {
		case LevelDebug:
			builder.WriteString(ColorGray)
		case LevelWarn:
			builder.WriteString(ColorYellow)
		case LevelError:
			builder.WriteString(ColorRed)
		case LevelFatal:
			builder.WriteString(ColorHiRed)
		case LevelPanic:
			builder.WriteString(ColorHiYellowOnRed)
		}
	}

	builder.WriteString(logStr)
	if msg.Level == LevelPanic && s.flags&Lstructured == 0 {
		builder.WriteByte('\n')
		for _, trace := range msg.Stack {
			builder.WriteString(trace)
			builder.WriteByte('\n')
		}
	}

	if s.flags&Lcolored != 0 {
		builder.WriteString(ColorReset)
	}

	builder.WriteByte('\n')
	finalLogStr = builder.String()

	// s.mtx.Lock()
	// defer s.mtx.Unlock()
	if msg.Level < LevelError {
		_, err = os.Stdout.WriteString(finalLogStr)
	} else {
		_, err = os.Stderr.WriteString(finalLogStr)
	}
	return
}
