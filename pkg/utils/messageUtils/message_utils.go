package messageUtils

import (
	"fmt"

	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

type MessageType string

const (
	Success MessageType = "success"
	Warning MessageType = "warning"
	Info    MessageType = "info"
	Error   MessageType = "error"
)

func Format(msgType MessageType, message string) string {
	switch msgType {
	case Success:
		return colorUtils.Success("Success:") + " " + message
	case Warning:
		return colorUtils.Warning("Warning:") + " " + message
	case Info:
		return colorUtils.Info("Info:") + " " + message
	case Error:
		return colorUtils.Error("Error:") + " " + message
	default:
		return message
	}
}

func Formatf(msgType MessageType, format string, a ...any) string {
	message := fmt.Sprintf(format, a...)
	return Format(msgType, message)
}
func SuccessMsg(message string) string {
	return Format(Success, message)
}

func SuccessMsgf(format string, a ...any) string {
	return Formatf(Success, format, a...)
}

func WarningMsg(message string) string {
	return Format(Warning, message)
}

func WarningMsgf(format string, a ...any) string {
	return Formatf(Warning, format, a...)
}

func InfoMsg(message string) string {
	return Format(Info, message)
}

func InfoMsgf(format string, a ...any) string {
	return Formatf(Info, format, a...)
}

func ErrorMsg(message string) string {
	return Format(Error, message)
}

func ErrorMsgf(format string, a ...any) string {
	return Formatf(Error, format, a...)
}

func Bold(message string) string {
	return colorUtils.Bold(message)
}

func Emphasize(message string) string {
	return colorUtils.Emphasize(message)
}

func Highlight(message string) string {
	return colorUtils.Highlight(message)
}

func Seperator(message string) string {
	return colorUtils.Seperator(message)
}
