package common

import (
	"regexp"
	"strconv"
	"time"

	"github.com/runableapp/when-ng/rules"
)

/*
ISO date formats:
- YYYY-MM-DD
- YYYY-MM-DD HH:MM:SS
- YYYY-MM-DD HH:MM:SS AM/PM
- RFC3339: YYYY-MM-DDTHH:MM:SS±TZ
- 2026-01-16
- 2026-01-16 04:00:00
- 2026-01-16 04:00:00 AM
- 2020-05-22T15:55-04:00
- 2020-05-22T15:55:00Z
*/

func ISODate(s rules.Strategy) rules.Rule {
	return &rules.F{
		// Match ISO date pattern: YYYY-MM-DD with optional time
		// Match full string by including time in the capture group
		RegExp: regexp.MustCompile("(?i)(?:\\W|^)" +
			"((?:1|2)[0-9]{3}\\-[0-1][0-9]\\-[0-3][0-9]" + // Date part
			"(?:[Tt\\s]+[0-9]{1,2}\\:[0-5][0-9](?:\\:[0-5][0-9])?(?:\\s*(?:A\\.|P\\.|A\\.M\\.|P\\.M\\.|AM?|PM?))?(?:[\\+\\-][0-9]{1,2}\\:[0-9]{2}|Z)?)?" + // Optional time
			")(?:\\W|$)"),
		Applier: func(m *rules.Match, c *rules.Context, o *rules.Options, ref time.Time) (bool, error) {
			if (c.Year != nil || c.Month != nil || c.Day != nil || c.Hour != nil || c.Minute != nil) && s != rules.Override {
				return false, nil
			}

			// Use m.Text which should contain the full matched string
			matchedText := m.Text
			
			// Always try to extract full ISO pattern from context text (clustered match)
			// This ensures we get the time part even if m.Text only has the date
			fullPattern := regexp.MustCompile("((?:1|2)[0-9]{3}\\-[0-1][0-9]\\-[0-3][0-9](?:[Tt\\s]+[0-9]{1,2}\\:[0-5][0-9](?:\\:[0-5][0-9])?(?:\\s*(?:A\\.|P\\.|A\\.M\\.|P\\.M\\.|AM?|PM?))?(?:[\\+\\-][0-9]{1,2}\\:[0-9]{2}|Z)?)?)")
			if fullMatches := fullPattern.FindStringSubmatch(c.Text); len(fullMatches) > 1 {
				matchedText = fullMatches[1]
			}
			
			// Trim whitespace
			matchedText = regexp.MustCompile("^\\s+|\\s+$").ReplaceAllString(matchedText, "")
			
			// Try manual parsing first (more reliable for our specific formats)
			var parsedTime time.Time
			parsed := false
			
			// Manual parsing - extract components directly from the string
			// This is more reliable than Go's parser for our specific formats
			dateTimePattern := regexp.MustCompile("^((?:1|2)[0-9]{3})\\-([0-1][0-9])\\-([0-3][0-9])(?:[Tt\\s]+([0-9]{1,2})\\:([0-5][0-9])(?:\\:([0-5][0-9]))?(?:\\s*(AM|PM|A\\.M\\.|P\\.M\\.|A\\.|P\\.))?(?:([\\+\\-])([0-9]{2})\\:([0-9]{2})|Z)?)?$")
			manualMatches := dateTimePattern.FindStringSubmatch(matchedText)
			
			// If manual parsing doesn't match, try without anchors (in case of extra chars)
			if len(manualMatches) <= 3 {
				dateTimePattern2 := regexp.MustCompile("((?:1|2)[0-9]{3})\\-([0-1][0-9])\\-([0-3][0-9])(?:[Tt\\s]+([0-9]{1,2})\\:([0-5][0-9])(?:\\:([0-5][0-9]))?(?:\\s*(AM|PM|A\\.M\\.|P\\.M\\.|A\\.|P\\.))?(?:([\\+\\-])([0-9]{2})\\:([0-9]{2})|Z)?)?")
				manualMatches = dateTimePattern2.FindStringSubmatch(matchedText)
			}
			
			if len(manualMatches) > 3 {
				year, _ := strconv.Atoi(manualMatches[1])
				month, _ := strconv.Atoi(manualMatches[2])
				day, _ := strconv.Atoi(manualMatches[3])
				
				hour, minute, second := 0, 0, 0
				if len(manualMatches) > 4 && manualMatches[4] != "" {
					hour, _ = strconv.Atoi(manualMatches[4])
				}
				if len(manualMatches) > 5 && manualMatches[5] != "" {
					minute, _ = strconv.Atoi(manualMatches[5])
				}
				if len(manualMatches) > 6 && manualMatches[6] != "" {
					second, _ = strconv.Atoi(manualMatches[6])
				}
				
				// Handle AM/PM
				if len(manualMatches) > 7 && manualMatches[7] != "" {
					ampm := manualMatches[7]
					if len(ampm) > 0 && (ampm[0] == 'P' || ampm[0] == 'p') {
						if hour < 12 {
							hour += 12
						}
					} else if hour == 12 {
						hour = 0
					}
				}
				
				// Set values directly from parsed components
				c.Year = &year
				c.Month = &month
				c.Day = &day
				c.Hour = &hour
				c.Minute = &minute
				c.Second = &second
				
				return true, nil
			} else {
				// Fallback to Go's time parser
				layouts := []string{
					time.RFC3339,
					time.RFC3339Nano,
					"2006-01-02T15:04:05-07:00",
					"2006-01-02T15:04:05Z",
					"2006-01-02T15:04:05",
					"2006-01-02 03:04:05 PM",
					"2006-01-02 03:04:05 AM",
					"2006-01-02 15:04:05",
					"2006-01-02 3:04:05 PM",
					"2006-01-02 3:04:05 AM",
					"2006-01-02 03:04 PM",
					"2006-01-02 03:04 AM",
					"2006-01-02 15:04",
					"2006-01-02",
				}
				
				var err error
				for _, layout := range layouts {
					parsedTime, err = time.Parse(layout, matchedText)
					if err == nil {
						parsed = true
						break
					}
				}
				
				if !parsed {
					return false, nil
				}
			}

			// Extract components from parsed time (fallback to Go's parser)
			year := parsedTime.Year()
			month := int(parsedTime.Month())
			day := parsedTime.Day()
			hour := parsedTime.Hour()
			minute := parsedTime.Minute()
			second := parsedTime.Second()

			c.Year = &year
			c.Month = &month
			c.Day = &day
			c.Hour = &hour
			c.Minute = &minute
			c.Second = &second

			return true, nil
		},
	}
}
