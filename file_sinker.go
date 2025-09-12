package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type fileCounter struct {
	fileName string
	count    int
	isErr    bool
}

type fileSinker struct {
	formater *LogFormater

	filePath   string
	fileName   string
	fileObj    *os.File
	errFileObj *os.File

	rotateCnt    uint
	errRotateCnt uint
	maxLogSize   int64

	flags int
	// mtx sync.Mutex
}

// 提取文件名中的序号，并判断是否含有.err后缀
func extractIndex(prefix string) (int, bool) {
	re := regexp.MustCompile(`\.(\d+)(?:\.err)?$`)
	matches := re.FindStringSubmatch(prefix)
	hasErrSuffix := strings.HasSuffix(prefix, ".err")
	if len(matches) > 1 {
		index, err := strconv.Atoi(matches[1])
		if err == nil {
			return index, hasErrSuffix
		}
	}
	return -1, hasErrSuffix
}

func NewFileSinker(format, filePath, fileName string, maxLogSize int64, flags int) (*fileSinker, error) {
	// 将相对路径转换为绝对路径
	if !path.IsAbs(filePath) {
		dir, _ := os.Getwd()
		filePath = path.Join(dir, filePath)
	}
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return nil, err
	}
	f, err := NewLogFormatter(format)
	if err != nil {
		return nil, err
	}
	s := &fileSinker{
		formater:     f,
		filePath:     filePath,
		fileName:     fileName,
		rotateCnt:    0,
		errRotateCnt: 0,
		maxLogSize:   maxLogSize,
		flags:        flags,
	}
	if err := s.initFile(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *fileSinker) Flags() int {
	return s.flags
}

func (s *fileSinker) Sink(msg *LogMsg) (err error) {
	// 文件分割，中途由于文件句柄被切换、关闭，可能造成并发问题，所以不要以协程方式执行
	s.splitFile(msg.Level)
	var logStr string

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

	var builder strings.Builder
	builder.WriteString(logStr)
	if msg.Level == LevelPanic && s.flags&Lstructured == 0 {
		builder.WriteByte('\n')
		for _, trace := range msg.Stack {
			builder.WriteString(trace)
			builder.WriteByte('\n')
		}
	}
	builder.WriteByte('\n')
	logStr = builder.String()

	// s.mtx.Lock()
	// defer s.mtx.Unlock()
	if msg.Level < LevelError {
		_, err = s.fileObj.WriteString(logStr)
	} else {
		_, err = s.errFileObj.WriteString(logStr)
	}
	return
}

// 初始化保存的日志文件和错误日志文件
func (f *fileSinker) initFile() (err error) {
	logFileName := path.Join(f.filePath, f.fileName)

	// 检查filePath下面最后一个日志文件的rotateCnt，更新当前的rotateCnt
	if files, err := os.ReadDir(f.filePath); err == nil {
		// fmt.Println(files)
		fileCntors := make([]*fileCounter, 0)
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			index, isErr := extractIndex(file.Name())
			if index != -1 {
				fileCntors = append(fileCntors, &fileCounter{
					fileName: file.Name(),
					count:    index,
					isErr:    isErr,
				})
			}
		}
		sort.Slice(fileCntors, func(i, j int) bool {
			return fileCntors[i].count > fileCntors[j].count
		})
		setRotateCnt := false
		setErrRotateCnt := false
		for _, fileCntor := range fileCntors {
			if fileCntor.isErr {
				f.errRotateCnt = uint(fileCntor.count)
				setErrRotateCnt = true
			} else {
				f.rotateCnt = uint(fileCntor.count)
				setRotateCnt = true
			}
			if setRotateCnt && setErrRotateCnt {
				break
			}
		}
	}

	f.fileObj, err = os.OpenFile(fmt.Sprintf("%s.%d", logFileName, f.rotateCnt), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		err = fmt.Errorf("open log file failed!: %v", err)
		return
	}
	f.errFileObj, err = os.OpenFile(fmt.Sprintf("%s.%d.err", logFileName, f.errRotateCnt), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("open err log file failed!: %v", err)
		return
	}
	return
}

func (f *fileSinker) splitFile(level int) error {
	var oldLogFile *os.File
	if level < LevelError {
		oldLogFile = f.fileObj
	} else {
		oldLogFile = f.errFileObj
	}

	fileInfo, err := oldLogFile.Stat()
	if err != nil {
		return err
	}
	// 文件超出大小时创建新文件
	if fileInfo.Size() <= f.maxLogSize {
		return nil
	}

	logFileName := path.Join(f.filePath, f.fileName)
	f.rotateCnt++
	newFullFileName := fmt.Sprintf("%s.%d", logFileName, f.rotateCnt)

	// 打开新的日志文件
	var newLogFile *os.File
	if level < LevelError {
		newLogFile, err = os.OpenFile(newFullFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open new log file failed: %v", err)
		}
		f.fileObj = newLogFile
	} else {
		newLogFile, err = os.OpenFile(newFullFileName+".err", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("open new error log file failed: %v", err)
		}
		f.errFileObj = newLogFile
	}

	// 关闭旧文件
	if err = oldLogFile.Close(); err != nil {
		return fmt.Errorf("close old log file failed: %v", err)
	}

	return nil
}
