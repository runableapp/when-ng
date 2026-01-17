package when

import (
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/runableapp/when-ng/rules"
	"github.com/runableapp/when-ng/rules/br"
	"github.com/runableapp/when-ng/rules/common"
	"github.com/runableapp/when-ng/rules/en"
	"github.com/runableapp/when-ng/rules/nl"
	"github.com/runableapp/when-ng/rules/ru"
)

// Parser is a struct which contains options
// rules, and middlewares to call
type Parser struct {
	options    *rules.Options
	rules      []rules.Rule
	middleware []func(string) (string, error)
}

// Result is a struct which contains parsing meta-info
type Result struct {
	// Index is a start index
	Index int
	// Text is a text found and processed
	Text string
	// Source is input string
	Source string
	// Time is an output time
	Time time.Time
}

// Parse returns Result and error if any. If have not matches it returns nil, nil.
func (p *Parser) Parse(text string, base time.Time) (*Result, error) {
	res := Result{
		Source: text,
		Time:   base,
		Index:  -1,
	}

	if p.options == nil {
		p.options = defaultOptions
	}

	var err error
	// apply middlewares
	for _, b := range p.middleware {
		text, err = b(text)
		if err != nil {
			return nil, err
		}
	}

	// First, try Go's built-in date/time parsing for standard formats
	// This handles ISO 8601, RFC3339, and other standard formats
	if parsedTime, matchedText := tryStandardTimeFormats(text, base); parsedTime != nil {
		res.Time = *parsedTime
		res.Text = matchedText
		res.Index = 0
		return &res, nil
	}

	// find all matches
	matches := make([]*rules.Match, 0)
	c := float64(0)
	for _, rule := range p.rules {
		r := rule.Find(text)
		if r != nil {
			r.Order = c
			c++
			matches = append(matches, r)
		}
	}

	// not found
	if len(matches) == 0 {
		return nil, nil
	}

	// find a cluster
	sort.Sort(rules.MatchByIndex(matches))

	// get borders of the matches
	end := matches[0].Right
	res.Index = matches[0].Left

	for i, m := range matches {
		if m.Left <= end+p.options.Distance {
			end = m.Right
		} else {
			matches = matches[:i]
			break
		}
	}

	res.Text = text[res.Index:end]

	// apply rules
	if p.options.MatchByOrder {
		sort.Sort(rules.MatchByOrder(matches))
	}

	ctx := &rules.Context{Text: res.Text}
	applied := false
	for _, applier := range matches {
		ok, err := applier.Apply(ctx, p.options, res.Time)
		if err != nil {
			return nil, err
		}
		applied = ok || applied
	}

	if !applied {
		return nil, nil
	}

	res.Time, err = ctx.Time(res.Time)
	if err != nil {
		return nil, errors.Wrap(err, "bind context")
	}

	return &res, nil
}

// Add adds  given rules to the main chain.
func (p *Parser) Add(r ...rules.Rule) {
	p.rules = append(p.rules, r...)
}

// Use adds give functions to middlewares.
func (p *Parser) Use(f ...func(string) (string, error)) {
	p.middleware = append(p.middleware, f...)
}

// SetOptions sets options object to use.
func (p *Parser) SetOptions(o *rules.Options) {
	p.options = o
}

// New returns Parser initialised with given options.
func New(o *rules.Options) *Parser {
	if o == nil {
		return &Parser{options: defaultOptions}
	}
	return &Parser{options: o}
}

// default options for internal usage
var defaultOptions = &rules.Options{
	Distance:     5,
	MatchByOrder: true,
}

// EN is a parser for English language
var EN *Parser

// RU is a parser for Russian language
var RU *Parser

// BR is a parser for Brazilian Portuguese language
var BR *Parser

// NL is a parser for Dutch language
var NL *Parser

func init() {
	EN = New(nil)
	EN.Add(en.All...)
	EN.Add(common.ISODate(rules.Override))
	EN.Add(common.SlashDMY(rules.Skip))

	RU = New(nil)
	RU.Add(ru.All...)
	RU.Add(common.All...)

	BR = New(nil)
	BR.Add(br.All...)
	BR.Add(common.All...)

	NL = New(nil)
	NL.Add(nl.All...)
	NL.Add(common.All...)
}

// tryStandardTimeFormats attempts to parse the text using Go's standard time formats
// Returns the parsed time and matched text if successful, nil otherwise
// If the parsed time doesn't have a timezone, it uses the base time's location
func tryStandardTimeFormats(text string, base time.Time) (*time.Time, string) {
	// Trim whitespace
	trimmed := text
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t' || trimmed[0] == '\n') {
		trimmed = trimmed[1:]
	}
	for len(trimmed) > 0 {
		last := len(trimmed) - 1
		if trimmed[last] == ' ' || trimmed[last] == '\t' || trimmed[last] == '\n' {
			trimmed = trimmed[:last]
		} else {
			break
		}
	}

	// Try common standard formats in order of specificity
	layouts := []string{
		time.RFC3339Nano,           // 2006-01-02T15:04:05.999999999Z07:00
		time.RFC3339,                // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05-07:00", // RFC3339 with timezone offset
		"2006-01-02T15:04-07:00",    // RFC3339 without seconds
		"2006-01-02T15:04:05Z",      // RFC3339 with Z
		"2006-01-02T15:04:05",       // ISO with T, no timezone
		"2006-01-02 15:04:05",       // ISO with space (24-hour)
		"2006-01-02 03:04:05 PM",    // ISO with space and PM (12-hour, hour 1-12)
		"2006-01-02 03:04:05 AM",    // ISO with space and AM (12-hour, hour 1-12)
		"2006-01-02 3:04:05 PM",     // ISO with space and PM (12-hour, single digit hour)
		"2006-01-02 3:04:05 AM",     // ISO with space and AM (12-hour, single digit hour)
		"2006-01-02 15:04",          // ISO without seconds (24-hour)
		"2006-01-02 03:04 PM",       // ISO without seconds, PM (12-hour)
		"2006-01-02 03:04 AM",       // ISO without seconds, AM (12-hour)
		"2006-01-02",                // Date only
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, trimmed)
		if err == nil {
			// If the parsed time doesn't have a timezone (location is UTC but no Z or offset),
			// use the base time's location
			if t.Location() == time.UTC && layout != time.RFC3339 && layout != time.RFC3339Nano {
				// Check if layout has timezone info
				hasTZ := false
				for _, tzLayout := range []string{time.RFC3339, time.RFC3339Nano, "Z07:00", "-07:00", "+07:00"} {
					if layout == tzLayout || contains(layout, "Z07:00") || contains(layout, "-07:00") || contains(layout, "+07:00") {
						hasTZ = true
						break
					}
				}
				if !hasTZ {
					// No timezone in layout, use base time's location
					t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), base.Location())
				}
			}
			return &t, trimmed
		}
	}

	return nil, ""
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
