package log

import (
	"fmt"
	"github.com/ttacon/chalk"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var ShowDebug = false
var LogStdOut = true
var LogPath = ""
var LogFile = false

func logPath() string {
	if LogPath == "" {
		currentUser, _ := user.Current()
		return filepath.Join(currentUser.HomeDir, ".conduit", "log.txt")
	} else {
		return LogPath
	}

}

func Info(msg string) {
	if LogStdOut {
		write(chalk.White.Color(msg))
	}
}

func Infof(msg string, a ...interface{}) {
	if LogStdOut {
		Info(fmt.Sprintf(msg, a...))
	}
}

func Warn(msg string) {
	if LogStdOut {
		write(chalk.Yellow.Color(msg))
	}
}

func Error(msg string) {
	if LogStdOut {
		write(chalk.Red.Color(msg))
	}
}

func Fatal(msg string) {
	if LogStdOut {
		write(chalk.Red.Color(msg))
		os.Exit(-1)
	}
}

func Debug(msg string) {
	if ShowDebug == true {
		if LogStdOut {
			write(chalk.Blue.Color(msg))
		}
	}
}

func Stats(name string, value interface{}) {
	nameStr := fmt.Sprintf("%s:", name)
	padding := strings.Repeat(" ", 20-len(nameStr))
	var valueStr = ""
	switch value.(type) {
	case int64:
		valueStr = fmt.Sprintf("%d", value)
	case string:
		valueStr = value.(string)
	}
	str := fmt.Sprintf("%s %s %s %s %s", nameStr, padding, chalk.Blue, valueStr,
		chalk.ResetColor)
	write(str)
}

func Status(label, value string, success bool) {
	var (
		nameStr string
		padding string
	)
	nameStr = fmt.Sprintf("%s:", label)
	padding = strings.Repeat(" ", 20-len(nameStr))
	if success {
		value = chalk.Green.Color(value)
	} else {
		value = chalk.Red.Color(value)
	}
	str := nameStr + padding + value
	write(str)
}

func write(text string) {
	fmt.Println(text)
}

func writeFile(logType, text string) {
	file, err := os.OpenFile(logPath(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err == nil {
		now := time.Now().Format("2006-01-02 15:04:05")
		logText := fmt.Sprintf("[%s] %s - %s", logType, now, text)
		file.WriteString(logText)
	}
	if file != nil {
		file.Close()
	}
}
