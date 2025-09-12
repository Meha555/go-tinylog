package tinylog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
)

type LogFormater struct {
	items []formatItem
}

// NOTE 为了能返回 out 中的字符串，我们需要使用带缓冲的 io.Writer，这里选择使用 bytes.Buffer。
// bytes.Buffer 实现了 io.Writer 接口，并且可以通过 String() 方法获取存储的字符串。
func (f *LogFormater) Format(msg *LogMsg) string {
	out := bytes.NewBuffer([]byte{})
	for _, item := range f.items {
		if err := item.format(out, msg); err != nil {
			// 处理格式化错误，这里简单忽略，实际使用中可以根据需求处理
			continue
		}
	}
	return out.String()
}

func NewLogFormatter(format string) (*LogFormater, error) {
	f := &LogFormater{}
	if err := f.parseFormat(format); err != nil {
		return nil, err
	}
	return f, nil
}

var formatItemMap = map[byte]formatItem{
	'l': &levelFormatItem{},
	'c': &categoryFormatItem{},
	'f': &fileFormatItem{},
	'L': &lineFormatItem{},
	'C': &funcNameFormatItem{},
	't': &timestampFormatItem{},
	'p': &plainTextFormatItem{},
	'g': &goroutineIDFormatItem{},
	// 's': &stackTraceFormatItem{},
	'm': &contentFormatItem{},
	'n': &newLineFormatItem{},
	'%': &precentSignFormatItem{},
}

func (f *LogFormater) parseFormat(format string) error {
	formatLen := len(format)
	if formatLen <= 0 {
		return errors.New("log format string can not be nil")
	}
	const (
		doScan = iota
		doCreate
	)
	var stat = doScan
	for i := 0; i < formatLen; i++ {
		switch stat {
		case doScan: // 一直取出普通字符
			var j = i
			for ; j < formatLen; j++ {
				if format[j] == '%' {
					stat = doCreate
					break
				}
			}
			content := format[i:j]
			i = j
			// NOTE 注意这里items是接口值，但是传入的是实现类的指针，因此go的多态也是父类指针指向子类对象
			f.items = append(f.items, &plainTextFormatItem{content})
		case doCreate: // 提取出特殊格式字符
			item, ok := formatItemMap[format[i]]
			if !ok {
				return fmt.Errorf("no such item for %%%c", format[i])
			}
			f.items = append(f.items, item)
			stat = doScan
		}
	}
	return nil
}

type formatItem interface {
	format(out io.Writer, msg *LogMsg) error
}

type levelFormatItem struct{}

func (item *levelFormatItem) format(out io.Writer, msg *LogMsg) error {
	levelStr, ok := LevelStrs[msg.Level]
	if !ok {
		levelStr = "UNKNOWN"
	}
	_, err := io.WriteString(out, levelStr)
	return err
}

type categoryFormatItem struct{}

func (item *categoryFormatItem) format(out io.Writer, msg *LogMsg) error {
	_, err := io.WriteString(out, msg.Category)
	return err
}

type fileFormatItem struct{}

func (item *fileFormatItem) format(out io.Writer, msg *LogMsg) error {
	// msg.WithFile(msg.callDepth)
	_, err := io.WriteString(out, filepath.Base(msg.File))
	return err
}

type lineFormatItem struct{}

func (item *lineFormatItem) format(out io.Writer, msg *LogMsg) error {
	// msg.WithLine(7)
	lineStr := strconv.Itoa(msg.Line)
	_, err := io.WriteString(out, lineStr)
	return err
}

type funcNameFormatItem struct{}

func (item *funcNameFormatItem) format(out io.Writer, msg *LogMsg) error {
	// msg.WithFuncName(7)
	_, err := io.WriteString(out, msg.FuncName)
	return err
}

type goroutineIDFormatItem struct{}

func (item *goroutineIDFormatItem) format(out io.Writer, msg *LogMsg) error {
	msg.WithGoroutineID()
	goroutineIDStr := strconv.Itoa(int(msg.GoroutineID))
	_, err := io.WriteString(out, goroutineIDStr)
	return err
}

type timestampFormatItem struct{}

func (item *timestampFormatItem) format(out io.Writer, msg *LogMsg) error {
	msg.WithTimestamp()
	_, err := io.WriteString(out, msg.Timestamp)
	return err
}

// type stackTraceFormatItem struct{}

// func (item *stackTraceFormatItem) format(out io.Writer, msg *LogMsg) error {
// 	msg.WithStack(msg.callDepth)
// 	var builder strings.Builder
// 	for _, trace := range msg.Stack {
// 		builder.WriteString(trace)
// 		builder.WriteByte('\n')
// 	}
// 	_, err := io.WriteString(out, builder.String())
// 	return err
// }

type contentFormatItem struct{}

func (item *contentFormatItem) format(out io.Writer, msg *LogMsg) error {
	_, err := io.WriteString(out, msg.Content)
	return err
}

type newLineFormatItem struct{}

func (item *newLineFormatItem) format(out io.Writer, msg *LogMsg) error {
	_, err := io.WriteString(out, string('\n'))
	return err
}

type precentSignFormatItem struct{}

func (item *precentSignFormatItem) format(out io.Writer, msg *LogMsg) error {
	_, err := io.WriteString(out, string('%'))
	return err
}

type plainTextFormatItem struct {
	plainText string
}

func (item *plainTextFormatItem) format(out io.Writer, msg *LogMsg) error {
	_, err := io.WriteString(out, item.plainText)
	return err
}
