package errorUtils

import (
	"fmt"
	"strings"

	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
)

func FormatError(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf("%s %s", colorUtils.Error("Error:"), msg)
}

func WrapError(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, a...)
	if strings.Contains(err.Error(), "Error:") {
		return fmt.Errorf("%s: %v", msg, err)
	}
	return fmt.Errorf("%s %s: %v", colorUtils.Error("Error:"), msg, err)
}
