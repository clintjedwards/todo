package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/clintjedwards/todo/proto"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// UnixMilli returns a humanized version of time given in unix millisecond. The zeroMsg is the string returned when
// the time is 0 and assumed to be not set.
func UnixMilli(unix int64, zeroMsg string, detail bool) string {
	if unix == 0 {
		return zeroMsg
	}

	if !detail {
		return humanize.Time(time.UnixMilli(unix))
	}

	relativeTime := humanize.Time(time.UnixMilli(unix))
	realTime := time.UnixMilli(unix).Format(time.RFC850)

	return fmt.Sprintf("%s (%s)", realTime, relativeTime)
}

// Takes a string enum and turns them into title case. If the value is unknown we turn it into
// a string of your choosing.
func NormalizeEnumValue[s ~string](value s, unknownString string) string {
	toTitle := cases.Title(language.AmericanEnglish)
	toLower := cases.Lower(language.AmericanEnglish)
	state := toTitle.String(toLower.String(string(value)))

	if strings.Contains(strings.ToLower(state), "unknown") {
		return unknownString
	}

	return state
}

func ColorizeTaskState(state string) string {
	switch strings.ToUpper(state) {
	case proto.Task_UNRESOLVED.String():
		return color.YellowString(state)
	case proto.Task_COMPLETED.String():
		return color.GreenString(state)
	default:
		return state
	}
}
