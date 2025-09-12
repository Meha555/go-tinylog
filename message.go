package tinylog

import (
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/petermattis/goid"
)

// LogMsg 单条日志的内容
// 考虑到性能问题，LogMsg是按需赋值各个字段。
// 结构化日志目前固定写死为JSON格式。可选的字段用json的omitempty标记，必需的字段则不使用omitempty标记。
type LogMsg struct {
	Level       int      `json:"level"`
	Category    string   `json:"category"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	FuncName    string   `json:"func_name,omitempty"`
	GoroutineID int64    `json:"goroutine_id,omitempty"`
	Timestamp   string   `json:"timestamp,omitempty"`
	Stack       []string `json:"stack,omitempty"`
	Content     string   `json:"content"`

	callDepth int `json:"-"` // 调用深度，用于记录日志时，获取调用栈的深度（file\line\func，不含stack）。`json:"-"`表示在进行序列化时忽略该字段
}

// func New(level int, category, content string) *LogMsg {
// 	funcName, fileName, lineNo, err := getRuntimeInfo(5)
// 	if err != nil {
// 		// 处理错误，这里简单使用默认值
// 		funcName = "nil"
// 		fileName = "nil"
// 		lineNo = 0
// 	}
// 	return &LogMsg{
// 		Level:       level,
// 		Category:    category,
// 		File:        fileName,
// 		Line:        lineNo,
// 		FuncName:    funcName,
// 		GoroutineID: goid.Get(),
// 		Timestamp:   time.Now(),
// 		Content:     content,
// 	}
// }

func NewMsg(level int, category, content string) *LogMsg {
	return &LogMsg{
		Level:     level,
		Category:  category,
		Content:   content,
		File:      "unknown",
		FuncName:  "unknown",
		Line:      -1,
		callDepth: -1,
	}
}

func (l *LogMsg) WithFile(skip int) *LogMsg {
	if l.File == "unknown" {
		if pc, file, line, ok := runtime.Caller(skip); ok {
			l.File = path.Base(file)
			l.FuncName = runtime.FuncForPC(pc).Name()
			l.Line = line
		}
	}
	return l
}

func (l *LogMsg) WithLine(skip int) *LogMsg {
	if l.Line == -1 {
		if pc, file, line, ok := runtime.Caller(skip); ok {
			l.File = runtime.FuncForPC(pc).Name()
			l.FuncName = path.Base(file)
			l.Line = line
		}
	}
	return l
}

func (l *LogMsg) WithFuncName(skip int) *LogMsg {
	if l.FuncName == "unknown" {
		if pc, file, line, ok := runtime.Caller(skip); ok {
			l.File = runtime.FuncForPC(pc).Name()
			l.FuncName = path.Base(file)
			l.Line = line
		}
	}
	return l
}

func (l *LogMsg) WithGoroutineID() *LogMsg {
	l.GoroutineID = goid.Get()
	return l
}

func (l *LogMsg) WithTimestamp() *LogMsg {
	l.Timestamp = time.Now().Format(time.DateTime)
	return l
}

func (l *LogMsg) WithStack(skip int) *LogMsg {
	for i := skip; ; i++ { // 跳过log包内的调用栈
		if pc, fileName, line, ok := runtime.Caller(i); ok {
			l.Stack = append(l.Stack, fmt.Sprintf("0x%x %s:%d", pc, fileName, line))
			if i == skip {
				l.File = fileName
				l.FuncName = runtime.FuncForPC(pc).Name()
				l.Line = line
			}
		} else {
			break
		}
	}
	return l
}

func (l *LogMsg) WithCallDepth(skip int) *LogMsg {
	l.callDepth = skip
	return l
}

// func (l *LogMsg) WithStack() *LogMsg {
// 	if l.Level != LevelDebug {
// 		return l
// 	}
// 	// 每次都创建新的字节切片
// 	stack := make([]byte, 1024)
// 	n := runtime.Stack(stack, true) // NOTE 会触发STW，所以不要频繁调用
// 	// 使用新的字节切片
// 	l.stack = stack[:n]
// 	return l
// }
