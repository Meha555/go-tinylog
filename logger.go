package tinylog

import (
	"fmt"
	"os"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal // 记录日志并额外执行 os.Exit(1)
	LevelPanic // 记录日志和堆栈并额外执行一条 panic() 语句
)

var LevelStrs = map[int]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
	LevelPanic: "PANIC",
}

// 预先定义一个全局默认的Logger实例，方便使用
var Log *Logger

func init() {
	var err error
	Log, err = NewStdLogger(LevelInfo, "DEFAULT", "[%t] [%c %l] [%f:%C:%L:%g] %m", false, Lcolored)
	if err != nil {
		panic(err)
	}
}

// Logger 日志器类
// NOTE logCh表示写日志是异步的，为了简单考虑没有保证由于主协程先于sink协程退出而导致的日志丢失。qi情况应该使用同步写
type Logger struct {
	baseLevel int
	category  string
	sinker    LogSinker
	logCh     chan *logMsg
}

// NewStdLogger 创建使用标准输出/错误为后端的日志器
//
//	level: 日志级别
//	category: 所属模块
//	format: 格式化字符串
//	async: 是否启用异步日志记录
//	flags: 日志标志位，支持 Lcolored 和 Lstructured
func NewStdLogger(level int, category, format string, async bool, flags int) (*Logger, error) {
	s, err := NewStdSinker(format, flags)
	if err != nil {
		return nil, err
	}
	l := &Logger{
		baseLevel: level,
		category:  category,
		sinker:    s,
		logCh:     nil,
	}

	if async {
		l.logCh = make(chan *logMsg, 10)
		go func() {
			for msg := range l.logCh {
				l.sinker.Sink(msg)
			}
		}()
	}

	return l, nil
}

// NewStdLogger 创建使用标准输出/错误为后端的日志器
//
//	level: 日志级别
//	category: 所属模块
//	format: 格式化字符串
//	filePath: 日志文件路径
//	fileName: 日志文件名
//	maxLogSize: 日志文件最大大小，超出大小自动分割
//	async: 是否启用异步日志记录
//	flags: 日志标志位，支持 Lcolored 和 Lstructured
func NewFileLogger(level int, category, format, filePath, fileName string, maxLogSize int64, async bool, flags int) (*Logger, error) {
	s, err := newFileSinker(format, filePath, fileName, maxLogSize, flags)
	if err != nil {
		return nil, err
	}
	l := &Logger{
		baseLevel: level,
		category:  category,
		sinker:    s,
		logCh:     nil,
	}

	if async {
		l.logCh = make(chan *logMsg, 10)
		go func() {
			for msg := range l.logCh {
				err := l.sinker.Sink(msg)
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
	}

	return l, nil
}

func (l *Logger) enable(level int) bool {
	return level >= l.baseLevel
}

func (l *Logger) doLog(level int, content interface{}, callDepth int, traceSkip int) {
	if l.enable(level) {
		var contentStr string

		// 根据不同类型进行处理，避免不必要的格式化
		if str, ok := content.(string); ok {
			contentStr = str
		} else if stringer, ok := content.(fmt.Stringer); ok {
			contentStr = stringer.String()
		} else {
			contentStr = fmt.Sprintf("%v", content)
		}

		msg := newMsg(level, l.category, contentStr).withCallDepth(callDepth)
		if traceSkip >= 0 {
			msg.withStack(traceSkip)
		}
		msg.withFile(3)
		if l.logCh == nil { // if sync
			l.sinker.Sink(msg)
		} else { // if async
			// msg.WithFile(3)
			l.logCh <- msg
		}
	}
}

func (l *Logger) Log(level int, content interface{}) {
	l.doLog(level, content, 6, -1)
}

func (l *Logger) Debug(content interface{}) {
	l.doLog(LevelDebug, content, 6, -1)
}

func (l *Logger) Info(content interface{}) {
	l.doLog(LevelInfo, content, 6, -1)
}

func (l *Logger) Warn(content interface{}) {
	l.doLog(LevelWarn, content, 6, -1)
}

func (l *Logger) Error(content interface{}) {
	l.doLog(LevelError, content, 6, -1)
}

func (l *Logger) Fatal(content interface{}) {
	l.doLog(LevelFatal, content, 6, -1)
	os.Exit(1)
}

// Panic 为了排除log包的调用栈，不得不自行构造LogMsg，然后传入sinker.Sink()
func (l *Logger) Panic(content interface{}) {
	l.doLog(LevelPanic, content, 6, 3)
	panic(content)
}

func (l *Logger) Logf(level int, format string, v ...interface{}) {
	if !l.enable(level) {
		return
	}
	l.doLog(level, fmt.Sprintf(format, v...), 6, -1)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if !l.enable(LevelDebug) {
		return
	}
	l.doLog(LevelDebug, fmt.Sprintf(format, v...), 6, -1)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if !l.enable(LevelInfo) {
		return
	}
	l.doLog(LevelInfo, fmt.Sprintf(format, v...), 6, -1)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if !l.enable(LevelWarn) {
		return
	}
	l.doLog(LevelWarn, fmt.Sprintf(format, v...), 6, -1)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if !l.enable(LevelError) {
		return
	}
	l.doLog(LevelError, fmt.Sprintf(format, v...), 6, -1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if !l.enable(LevelFatal) {
		return
	}
	l.doLog(LevelFatal, fmt.Sprintf(format, v...), 6, -1)
	os.Exit(1)
}

// Panicf 为了排除log包的调用栈，不得不自行构造LogMsg，然后传入sinker.Sink()
func (l *Logger) Panicf(format string, v ...interface{}) {
	if !l.enable(LevelPanic) {
		return
	}
	content := fmt.Sprintf(format, v...)
	l.doLog(LevelPanic, content, 6, 3)
	panic(content)
}

func (l *Logger) Level() int {
	return l.baseLevel
}

func (l *Logger) SetLevel(level int) {
	// 不对baseLevel的修改进行加锁，这里对一致性要求不高
	l.baseLevel = level
}
