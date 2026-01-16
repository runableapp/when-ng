package zh

import (
	"regexp"
	"strconv"
	"time"

	"github.com/runableapp/when-ng/rules"
)

/*
	规则名称：精确到月份的日期
*/

func ExactMonthDate(s rules.Strategy) rules.Rule {
	overwrite := s == rules.Override

	return &rules.F{
		RegExp: regexp.MustCompile("" +
			"(?:\\b|^)" + // can't use \W here due to Chinese characters
			"(?:" +
			"(1[0-2]|[1-9])" + "\\s*" + "(?:-|/|\\.)" + "\\s*" + "(1[0-9]|2[0-9]|3[0-1]|[1-9])" +
			"|" +
			"(?:" +
			"(1[0-2]|[1-9]|" + MON_WORDS_PATTERN + ")" + "\\s*" +
			"(月)" + "\\s*" +
			")?" +
			"(?:" +
			"(1[0-9]|2[0-9]|3[0-1]|[1-9]|" + DAY_WORDS_PATTERN + ")" + "\\s*" +
			"(日|号)" +
			")?" +
			")",
		),

		Applier: func(m *rules.Match, c *rules.Context, o *rules.Options, ref time.Time) (bool, error) {
			_ = overwrite

			// the default value of month is the current month, and the default
			// value of day is the first day of the month, so that we can handle
			// cases like "4月" (Apr 1st) and "12号" (12th this month)
			var monInt = int(ref.Month())
			var dayInt = 1
			var exist bool

			if m.Captures[0] == "" && m.Captures[2] == "" && m.Captures[4] == "" {
				return false, nil
			}

			if m.Captures[2] != "" {
				monInt, exist = MON_WORDS[compressStr(m.Captures[2])]
				if !exist {
					mon, err := strconv.Atoi(m.Captures[2])
					if err != nil {
						return false, nil
					}
					monInt = mon
				}
			}

			if m.Captures[4] != "" {
				dayInt, exist = DAY_WORDS[compressStr(m.Captures[4])]
				if !exist {
					day, err := strconv.Atoi(m.Captures[4])
					if err != nil {
						return false, nil
					}
					dayInt = day
				}
			}

			if m.Captures[0] != "" && m.Captures[1] != "" {
				mon, err := strconv.Atoi(m.Captures[0])
				if err != nil {
					return false, nil
				}
				day, err := strconv.Atoi(m.Captures[1])
				if err != nil {
					return false, nil
				}
				monInt = mon
				dayInt = day
			}

			c.Month = &monInt
			c.Day = &dayInt

			return true, nil
		},
	}
}
