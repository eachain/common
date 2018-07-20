/*
同下面定义相同:
	https://github.com/eachain/logger
*/

package logger

import (
	"io"
	"log"
	"os"
)

type Logger interface {
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

/*
WithPrefix 用于添加日志公共前缀，例如：

	func HandleLogicA(...) {
		log := logger.WithPrefix(logger.Get(), "HandleLogicA: ")
		log.Infof("hello world")
		// Output:
		// HandleLogicA: hello world

		logicB(log)
	}

	func logicB(log logger.Logger) {
		log = logger.WithPrefix(log, "logicB: ")
		log.Infof("hello world")
		// Output:
		// HandleLogicA: logicB: hello world
	}
*/
func WithPrefix(logger Logger, prefix string) Logger {
	return fmtLogger{l: logger, f: prefixFormatter{prefix: prefix}}
}

// WithSuffix 用于添加日志公共后缀，参考WithPrefix.
func WithSuffix(logger Logger, suffix string) Logger {
	return fmtLogger{l: logger, f: suffixFormatter{suffix: suffix}}
}

var (
	_logger Logger
)

func init() {
	SetOutput(os.Stderr)
}

// New 返回一个Logger，可在单独逻辑内使用。
func New(w io.Writer) Logger {
	l := log.New(w, "", log.LstdFlags|log.Lmicroseconds)
	return newLogger(l)
}

// SetOutput 应该被主线程调用，
// 确保在进程的整个生命周期内只被调用一次，
// 而不是经常更换。
func SetOutput(w io.Writer) {
	_logger = New(w)
}

// Get 获取全局Logger。
func Get() Logger {
	return _logger
}

func Infof(format string, a ...interface{}) {
	_logger.Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	_logger.Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	_logger.Errorf(format, a...)
}

// - - - - - - - - - - format logger - - - - - - - - - -

type formatter interface {
	format(string) string
}

type prefixFormatter struct {
	prefix string
}

func (pf prefixFormatter) format(s string) string {
	return pf.prefix + s
}

type suffixFormatter struct {
	suffix string
}

func (sf suffixFormatter) format(s string) string {
	return s + sf.suffix
}

type fmtLogger struct {
	l Logger
	f formatter
}

func (fl fmtLogger) Infof(format string, a ...interface{}) {
	fl.l.Infof(fl.f.format(format), a...)
}

func (fl fmtLogger) Warnf(format string, a ...interface{}) {
	fl.l.Warnf(fl.f.format(format), a...)
}

func (fl fmtLogger) Errorf(format string, a ...interface{}) {
	fl.l.Errorf(fl.f.format(format), a...)
}

// - - - - - - - - - - go logger - - - - - - - - - -

type printer interface {
	Printf(format string, a ...interface{})
}

type printLogger struct {
	l printer
}

func (p printLogger) Infof(format string, a ...interface{}) {
	p.l.Printf(format, a...)
}

func (p printLogger) Warnf(format string, a ...interface{}) {
	p.l.Printf(format, a...)
}

func (p printLogger) Errorf(format string, a ...interface{}) {
	p.l.Printf(format, a...)
}

// - - - - - - - - - - common logger - - - - - - - - - -

type logger struct {
	info Logger
	warn Logger
	err  Logger
}

func newLogger(p printer) logger {
	gl := printLogger{l: p}
	return logger{
		info: WithPrefix(gl, "[INFO] "),
		warn: WithPrefix(gl, "[WARN] "),
		err:  WithPrefix(gl, "[ERR] "),
	}
}

func (l logger) Infof(format string, a ...interface{}) {
	l.info.Infof(format, a...)
}

func (l logger) Warnf(format string, a ...interface{}) {
	l.warn.Warnf(format, a...)
}

func (l logger) Errorf(format string, a ...interface{}) {
	l.err.Errorf(format, a...)
}
