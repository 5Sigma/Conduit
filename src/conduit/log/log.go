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
	str := fmt.Sprintln(nameStr, padding, chalk.Blue, value,
		chalk.ResetColor)
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
