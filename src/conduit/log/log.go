package log

import (
	"fmt"
	"github.com/ttacon/chalk"
	"os"
)

var showDebug = false

func Info(msg string) {
	fmt.Println(chalk.White.Color(msg))
}

func Infof(msg string, a ...interface{}) {
	Info(fmt.Sprintf(msg, a...))
}

func Warn(msg string) {
	fmt.Println(chalk.Yellow.Color(msg))
}

func Error(msg string) {
	fmt.Println(chalk.Red.Color(msg))
}

func Fatal(msg string) {
	fmt.Println(chalk.Red.Color(msg))
	os.Exit(-1)
}

func Debug(msg string) {
	fmt.Println(chalk.Blue.Color(msg))
}
