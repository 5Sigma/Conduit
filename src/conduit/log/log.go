package log

import (
	"fmt"
	"github.com/ttacon/chalk"
)

func Info(msg string) {
	fmt.Println(chalk.White.Color(msg))
}

func Warn(msg string) {
	fmt.Println(chalk.Yellow.Color(msg))
}

func Error(msg string) {
	fmt.Println(chalk.Red.Color(msg))
}
