package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustismo/heavyfishdesign/dynmap"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info  LogLevel = iota
	Error LogLevel = iota
)

func (ll LogLevel) String() string {
	switch ll {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Error:
		return "error"
	}
	return ""
}

type LogMessage struct {
	*dynmap.DynMap
	Timestamp time.Time
	Level     LogLevel
	Message   string
}

func (lm LogMessage) ToDynMap() *dynmap.DynMap {
	dm := lm.DynMap.Clone()
	dm.Put("message", lm.Message)
	dm.Put("timestamp", lm.Timestamp)
	dm.Put("level", lm.Level.String())
	return dm
}

// This logger is intended to serve a single Render operation
// it tracks all log messages in a queue, that can then be displayed to the
// end user or dropped.
// if Debug is on, then all messages are written to STD out.
type HfdLog struct {
	// True if there are is least one error message
	hasErrors bool
	// should we also write messages to std out?
	LogToStdOut bool
	Messages    []LogMessage
	// these fields will be added to all log messages
	StaticFields *dynmap.DynMap
	parent       *HfdLog
}

func NewLog() *HfdLog {
	return &HfdLog{
		hasErrors:    false,
		LogToStdOut:  true,
		Messages:     []LogMessage{},
		StaticFields: dynmap.New(),
	}
}

func (logger *HfdLog) HasErrors() bool {
	return logger.hasErrors
}

// Creates a new logger from this one, that shares many of the same
// configurations, and anything logged to it, will also be logged to self
func (logger *HfdLog) NewChild() *HfdLog {
	return &HfdLog{
		hasErrors:    false,
		LogToStdOut:  logger.LogToStdOut,
		Messages:     []LogMessage{},
		StaticFields: dynmap.New(),
		parent:       logger,
	}
}

func (logger *HfdLog) logFromChild(msg LogMessage) {
	f := msg.DynMap
	if !logger.StaticFields.IsEmpty() {
		f = logger.StaticFields.Clone().Merge(f)
	}
	message := LogMessage{
		DynMap:    f,
		Message:   msg.Message,
		Timestamp: msg.Timestamp,
		Level:     msg.Level,
	}
	logger.addMsg(message)
}

func (logger *HfdLog) addMsg(message LogMessage) {
	// There is a possibility of a race condition here, but
	// I don't expect this to be a real problem.
	// maybe someday make it atomic..
	if logger.Messages == nil {
		logger.Messages = []LogMessage{}
	}
	logger.Messages = append(logger.Messages, message)
	if message.Level == Error {
		logger.hasErrors = true
	}
}

func (logger *HfdLog) Debugfd(fields *dynmap.DynMap, format string, v ...interface{}) {
	logger.Logf(Debug, fields, format, v...)
}
func (logger *HfdLog) Debugf(format string, v ...interface{}) {
	logger.Logf(Debug, dynmap.New(), format, v...)
}

func (logger *HfdLog) Infofd(fields *dynmap.DynMap, format string, v ...interface{}) {
	logger.Logf(Info, fields, format, v...)
}
func (logger *HfdLog) Infof(format string, v ...interface{}) {
	logger.Logf(Info, dynmap.New(), format, v...)
}

func (logger *HfdLog) Errorfd(fields *dynmap.DynMap, format string, v ...interface{}) {
	logger.Logf(Error, fields, format, v...)
}
func (logger *HfdLog) Errorf(format string, v ...interface{}) {
	logger.Logf(Error, dynmap.New(), format, v...)
}

func (logger *HfdLog) Logf(level LogLevel, fields *dynmap.DynMap, format string, v ...interface{}) {
	f := fields
	if !logger.StaticFields.IsEmpty() {
		f = logger.StaticFields.Clone().Merge(fields)
	}
	message := LogMessage{
		DynMap:    f,
		Message:   fmt.Sprintf(format, v...),
		Timestamp: time.Now(),
		Level:     level,
	}
	logger.addMsg(message)
	// log it to the parent logger as well
	if logger.parent != nil {
		logger.parent.logFromChild(message)
	}
	if logger.LogToStdOut {
		fmt.Printf("%s:: %s\n", strings.ToUpper(level.String()), message.Message)
	}
}
