package cblog

import (
	"bytes"
	"errors"
	"github.com/codingbeard/go-logger"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

type MultipleWriter struct {
	writers []io.Writer
}

func (m MultipleWriter) Write(p []byte) (n int, err error) {
	for _, writer := range m.writers {
		n, e := writer.Write(p)
		if e != nil {
			return n, e
		}
	}

	return n, nil
}

func NewMultipleWriter(writers ...io.Writer) MultipleWriter {
	return MultipleWriter{writers:writers}
}

type LogLevel int

const (
	CriticalLevel LogLevel = iota + 1
	ErrorLevel
	WarningLevel
	NoticeLevel
	InfoLevel
	DebugLevel
)

const (
	Black = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type LoggerConfig struct {
	LogLevel                LogLevel
	Format                  string // placeholders: %{id}, %{time[:fmt]}, %{module}, %{filename}, %{line}, %{level}, %{message}, %{category}
	LogToFile               bool
	FilePath                string
	FilePerm                os.FileMode
	LogToStdOut             bool
	StdOutColor             int
	LogToUnixSocket         bool
	UnixSocketPath          string
	AdditionalWriters       []io.Writer
	AdditionalWriterClosers []io.WriteCloser
	SetAsDefaultLogger      bool
	/*
		todo:
			ErrorReporter           func (e error)
			Rotate                  bool
			RotateFileSize          uint64
			RotateKeepCount         int
			Upload                  bool
			UploadInterval          time.Duration
			Uploader                func(fileName string, content []byte) error
	*/
}

type Logger struct {
	defaultFile *os.File
	config LoggerConfig
	logger *logger.Logger
	closers []io.Closer
}

func DefaultLoggerConfig() LoggerConfig {
	return LoggerConfig{
		LogLevel:    InfoLevel,
		Format:      "%{time:2006-01-02 15:04:05.000 -0700} : %{category} : %{level} : %{file}:%{line} : %{message}",
		LogToStdOut: true,
		FilePerm: os.ModePerm,
	}
}

func NewLogger(config LoggerConfig) (*Logger, error) {
	var l *logger.Logger
	var writers []io.Writer

	cblogger := &Logger{
		config: config,
	}

	if config.LogToFile {
		wr, e := os.OpenFile(config.FilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, config.FilePerm)
		if e != nil {
			return nil, e
		}
		cblogger.closers = append(cblogger.closers, wr)
		writers = append(writers, wr)
		cblogger.defaultFile = wr
	}

	if config.LogToStdOut {
		writers = append(writers, os.Stdout)
	}

	if config.LogToUnixSocket {
		writers = append(writers, NewUnixSockerLogger(config.UnixSocketPath))
	}

	if len(config.AdditionalWriters) != 0 {
		writers = append(writers, config.AdditionalWriters...)
	}

	if len(config.AdditionalWriterClosers) != 0 {
		for _, writeCloser := range config.AdditionalWriterClosers {
			cblogger.closers = append(cblogger.closers, writeCloser)
			writers = append(writers, writeCloser)
		}
	}

	l, e := logger.New(NewMultipleWriter(writers...), config.StdOutColor, config.LogLevel)
	if e != nil {
		return nil, e
	}

	l.SetFormat(config.Format)

	cblogger.logger = l

	if config.SetAsDefaultLogger {
		log.SetOutput(cblogger)
	}

	cblogger.InfoF("CBLOG", "Logger initialised")

	return cblogger, nil
}

func (l *Logger) GetUnderlyingLogger() *logger.Logger {
	return l.logger
}

func (l *Logger) Close() error {
	var es []string
	for _, closer := range l.closers {
		e := closer.Close()
		if e != nil {
			es = append(es, e.Error())
		}
	}

	if len(es) != 0 {
		return errors.New(strings.Join(es, "\n"))
	}

	return nil
}

func (l *Logger) FatalF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.FatalF(category, format, a...)
	} else {
		l.logger.Fatal(category, format)
	}
}

func (l *Logger) PanicF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.PanicF(category, format, a...)
	} else {
		l.logger.Panic(category, format)
	}
}

func (l *Logger) CriticalF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.CriticalF(category, format, a...)
	} else {
		l.logger.Critical(category, format)
	}
}

func (l *Logger) ErrorF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.ErrorF(category, format, a...)
	} else {
		l.logger.Error(category, format)
	}
}

func (l *Logger) WarningF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.WarningF(category, format, a...)
	} else {
		l.logger.Warning(category, format)
	}
}

func (l *Logger) NoticeF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.NoticeF(category, format, a...)
	} else {
		l.logger.Notice(category, format)
	}
}

func (l *Logger) InfoF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.InfoF(category, format, a...)
	} else {
		l.logger.Info(category, format)
	}
}

func (l *Logger) DebugF(category, format string, a ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if len(a) > 0 {
		l.logger.DebugF(category, format, a...)
	} else {
		l.logger.Debug(category, format)
	}
}

func (l *Logger) StackAsError(category, message string) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	stack := Stack()
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	l.ErrorF(category, message+stack)
}

func (l *Logger) StackAsCritical(category, message string) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(3)
	defer l.logger.SetPosOverride(pos)
	if message == "" {
		message = "Stack info"
	}
	message += "\n"
	stack := Stack()
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	l.CriticalF(category, message+stack)
}

func Stack() string {
	buf := make([]byte, 1000000)
	runtime.Stack(buf, false)
	buf = bytes.Trim(buf, "\x00")
	stack := string(buf)
	stackParts := strings.Split(stack, "\n")
	newStackParts := []string{stackParts[0]}
	newStackParts = append(newStackParts, stackParts[3:]...)
	stack = strings.Join(newStackParts, "\n")
	return stack
}

func (l *Logger) Write(bytes []byte) (int, error) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(5)
	defer l.logger.SetPosOverride(pos)
	return l.logger.Write(bytes)
}

func (l *Logger) Print(v ...interface{}) {
	pos := l.logger.GetPosOverride()
	l.logger.SetPosOverride(5)
	defer l.logger.SetPosOverride(pos)
	l.logger.Print(v...)
}

type UnixSocketLogger struct {
	socket net.Conn
	path   string
}

func NewUnixSockerLogger(path string) *UnixSocketLogger {
	return &UnixSocketLogger{path: path}
}

func (u *UnixSocketLogger) Write(message []byte) (n int, err error) {
	if u.socket == nil {
		u.initSocket()
	}

	if u.socket == nil {
		return
	}

	n, err = u.socket.Write(append(message, []byte("\n")...))
	if err != nil {
		u.initSocket()
		return u.socket.Write(append(message, []byte("\n")...))
	}

	return n, err
}

func (u *UnixSocketLogger) initSocket() {
	conn, e := net.Dial("unix", u.path)
	if e != nil {
		return
	}

	u.socket = conn
}
