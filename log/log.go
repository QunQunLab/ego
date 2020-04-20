package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

//// Logger log interface
//type Logger interface {
//	Trace(format string, arg ...interface{})
//	Info(format string, arg ...interface{})
//	Debug(format string, arg ...interface{})
//	Warn(format string, arg ...interface{})
//	Error(format string, arg ...interface{})
//	Fatal(format string, arg ...interface{})
//}

var glog = &log.Logger{}

func Init() {
	glog = log.New(os.Stderr, "", log.LstdFlags)
}

func Trace(format interface{}, arg ...interface{}) {
	glog.Printf(formatLog(format, arg...))
}

func Info(format interface{}, arg ...interface{}) {
	glog.Printf(formatLog(format, arg...))
}

func Debug(format interface{}, arg ...interface{}) {
	glog.Printf(formatLog(format, arg...))
}

func Warn(format interface{}, arg ...interface{}) {
	glog.Printf(formatLog(format, arg...))
}

func Error(format interface{}, arg ...interface{}) {
	glog.Printf(formatLog(format, arg...))
}

func Fatal(format interface{}, arg ...interface{}) {
	glog.Fatal(formatLog(format, arg...))
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
