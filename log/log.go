package log

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/QunQunLab/ego/conf"
	"github.com/zerak/log"
	"github.com/zerak/log/provider"
)

type LogOption struct {
	Dir         string `json:"dir,omitempty"`          // log directory(default: .)
	Filename    string `json:"filename,omitempty"`     // log filename(default: <appName>.log)
	NoSymlink   bool   `json:"nosymlink,omitempty"`    // doesn't create symlink to latest log file(default: false)
	MaxSize     int    `json:"maxsize,omitempty"`      // max bytes number of every log file(default: 64M)
	DailyAppend bool   `json:"daily_append,omitempty"` // append to existed file instead of creating a new file(default: true)
	Suffix      string `json:"suffix,omitempty"`       // filename suffix
	DateFormat  string `json:"date_format,omitempty"`  // date format string(default: %04d%02d%02d)
	Level       string `json:"level,omitempty"`        //level, 0:fatal 1:error 2:warn 3:info 4:debug 5:trace
}

func Init(opts ...LogOption) error {

	if len(opts) > 0 {
		mfOpts, _ := json.Marshal(opts[0])
		consoleOpts := fmt.Sprintf(`{"tostderrlevel":%d}`, opts[0].Level)
		p := provider.NewMixProvider(provider.NewFile(string(mfOpts)), provider.NewColoredConsole(consoleOpts))
		log.InitWithProvider(p)
		log.SetLevelFromString(opts[0].Level)
	} else {
		var (
			rootDir  = "./log"
			filename = "app"
			level    = "debug"
			maxSize  = int64(1 << 26) // 1*2^26 = 64M
		)
		l := conf.Get("log")
		if l != nil {
			rootDir, _ = l.String("root", "./log")
			filename, _ = l.String("name", "app")
			level, _ = l.String("level")
			maxSize, _ = l.Int("maxsize", 1<<26)
		}
		mfOpts, _ := json.Marshal(&LogOption{
			Dir:      rootDir,
			Filename: filename,
			MaxSize:  int(maxSize),
			Level:    level,
		})
		consoleOpts := fmt.Sprintf(`{"tostderrlevel":%s}`, level)
		p := provider.NewMixProvider(provider.NewFile(string(mfOpts)), provider.NewColoredConsole(consoleOpts))
		log.InitWithProvider(p)
		log.SetLevelFromString(level)
	}
	return nil
}

func Uninit(err error) {
	log.Uninit(err)
}

func Trace(format interface{}, arg ...interface{}) {
	log.Trace(formatLog(format, arg...))
}

func Info(format interface{}, arg ...interface{}) {
	log.Info(formatLog(format, arg...))
}

func Debug(format interface{}, arg ...interface{}) {
	log.Debug(formatLog(format, arg...))
}

func Warn(format interface{}, arg ...interface{}) {
	log.Warn(formatLog(format, arg...))
}

func Error(format interface{}, arg ...interface{}) {
	log.Error(formatLog(format, arg...))
}

func Fatal(format interface{}, arg ...interface{}) {
	log.Fatal(formatLog(format, arg...))
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

// If returns an `IfLogger`
func If(ok bool) log.IfLogger {
	if ok {
		return log.IfLogger(0xFF)
	} else {
		return log.IfLogger(0x00)
	}
}
