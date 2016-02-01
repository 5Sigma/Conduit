package log

import (
	"fmt"
	"github.com/ttacon/chalk"
	"os"
	"strings"
)

var ShowDebug = false
var LogStdOut = true

func Info(msg string) {
	if LogStdOut {
		fmt.Println(chalk.White.Color(msg))
	}
}

func Infof(msg string, a ...interface{}) {
	if LogStdOut {
		Info(fmt.Sprintf(msg, a...))
	}
}

func Warn(msg string) {
	if LogStdOut {
		fmt.Println(chalk.Yellow.Color(msg))
	}
}

func Error(msg string) {
	if LogStdOut {
		fmt.Println(chalk.Red.Color(msg))
	}
}

func Fatal(msg string) {
	if LogStdOut {
		fmt.Println(chalk.Red.Color(msg))
		os.Exit(-1)
	}
}

func Debug(msg string) {
	if ShowDebug == true && LogStdOut {
		fmt.Println(chalk.Blue.Color(msg))
	}
}

func Stats(name string, value interface{}) {
	nameStr := fmt.Sprintf("%s:", name)
	padding := strings.Repeat(" ", 20-len(nameStr))
	str := fmt.Sprintln(nameStr, padding, chalk.Blue, value,
		chalk.ResetColor)
	fmt.Print(str)
}
