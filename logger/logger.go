package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Trace = iota
	Debug
	Info
	Warn
	Error
	Fatal

	TRACE = "TRACE"
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

var (
	TimeFromat = "02.01.2006 15:04:05.000"
	m          map[string]int
)

type Logger struct {
	outputs    []io.Writer
	logLevel   int
	timeFormat string
	once       sync.Once
}

func init() {
	m = map[string]int{
		TRACE: Trace,
		DEBUG: Debug,
		INFO:  Info,
		WARN:  Warn,
		ERROR: Error,
		FATAL: Fatal,
	}
}

func NewLogger() *Logger {
	log := &Logger{}
	log.SetLevel(Trace)
	log.SetTimeFormat(TimeFromat)

	return log
}

func (l *Logger) SetTimeFormat(format string) {
	l.timeFormat = format
}

func (l *Logger) SetLevel(level int) {
	l.logLevel = level
}

func (l *Logger) SetOutput(outputs ...io.Writer) {
	l.once.Do(func() {
		l.outputs = outputs
	})
}

func (l *Logger) Trace(text string) {
	l.write(TRACE, text, getFrame(1))
}

func (l *Logger) Debug(text string) {
	l.write(DEBUG, text, getFrame(1))
}

func (l *Logger) Info(text string) {
	l.write(INFO, text, getFrame(1))
}

func (l *Logger) Warn(text string) {
	l.write(WARN, text, getFrame(1))
}

func (l *Logger) Error(text string) {
	l.write(ERROR, text, getFrame(1))
}

func (l *Logger) Fatal(text string) {
	l.write(FATAL, text, getFrame(1))
	os.Exit(1)
}

func (l *Logger) write(logLevel string, text string, frame runtime.Frame) {
	if l.logLevel > m[logLevel] {
		return
	}

	for _, output := range l.outputs {
		log := l.getLog(logLevel, text, frame)

		_, err := output.Write([]byte(log))
		if err != nil {
			panic(fmt.Errorf("log write error: %w", err))
		}
	}
}

func (l *Logger) getLog(logLevel string, text string, frame runtime.Frame) string {
	return fmt.Sprintf("%s %s [%s.%d goid-%d] %s\n", l.getTimeNow(), logLevel, frame.Function, frame.Line, goid(), text)
}

func (l *Logger) getTimeNow() string {
	return time.Now().Format(l.timeFormat)
}

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
