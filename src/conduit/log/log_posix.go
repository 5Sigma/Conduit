// +build !windows

package log

import (
	"fmt"
	"github.com/ttacon/chalk"
	"strings"
)

func Status(label, value string, success bool) {
	if success {
		value = chalk.Green.Color(value)
	} else {
		value = chalk.Red.Color(value)
	}
	valueStr := fmt.Sprintf("[ %s ] ", value)
	str := valueStr + label
	write("", str, noStyle)
}

func write(tag, text string, style func(string) string) {
	fmt.Println(style(text))
	if LogFile == true {
		writeFile(tag, text)
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
	write("", str, noStyle)
}
