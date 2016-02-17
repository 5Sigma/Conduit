package log

import (
	"fmt"
	"github.com/kardianos/osext"
	"github.com/ttacon/chalk"
	"os"
	"path/filepath"
	"time"
)

var ShowDebug = false
var LogFile = false

func logPath() string {
	now := time.Now().Format("2006-01-02")
	fName := fmt.Sprintf("%s.log", now)
	directory, _ := osext.ExecutableFolder()
	path := filepath.Join(directory, "logs", fName)
	return path
}

func init() {
	directory, _ := osext.ExecutableFolder()
	path := filepath.Join(directory, "logs")
	os.MkdirAll(path, 0755)
}

func Info(msg string) {
	write("INFO", msg, chalk.White.Color)
}

func Infof(msg string, a ...interface{}) {
	Info(fmt.Sprintf(msg, a...))
}

func Warn(msg string) {
	write("WARN", msg, chalk.Yellow.Color)
}

func Error(msg string) {
	Errorf("%s", msg)
}
func Errorf(msg string, a ...interface{}) {
	write("ERROR", fmt.Sprintf(msg, a...), chalk.Red.Color)
}

func Fatal(msg string) {
	write("FATAL", msg, chalk.Red.Color)
	os.Exit(-1)
}

func Debug(msg string) {
	if ShowDebug == true {
		write("DEBUG", msg, chalk.Blue.Color)
	}
}

func Alertf(msg string, a ...interface{}) {
	write("ALERT", fmt.Sprintf(msg, a...), chalk.Bold.TextStyle)
}

func Alert(msg string) {
	Alertf("%s", msg)
}

func noStyle(s string) string {
	return s
}

func writeFile(logType, text string) {
	file, err := os.OpenFile(logPath(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err == nil {
		now := time.Now().Format("2006-01-02 15:04:05")
		logText := fmt.Sprintf("[%s] %s - %s\n", logType, now, text)
		file.WriteString(logText)
	}
	if file != nil {
		file.Close()
	}
}
