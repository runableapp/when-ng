# Rules System Analysis

## Overview

The `when-ng` project uses a **rule-based parsing system** where you **must explicitly specify** which language rules to use. There is **no automatic language detection** - you choose the parser or rules at initialization.

## Two Ways to Use Rules

### 1. **Pre-initialized Language Parsers** (Recommended for single language)

The library provides pre-configured parsers for each language:

```go
import "github.com/runableapp/when-ng"

// Use the pre-initialized English parser
result, err := when.EN.Parse("tomorrow at 5pm", time.Now())

// Use the pre-initialized Russian parser
result, err := when.RU.Parse("завтра в 5 вечера", time.Now())

// Use the pre-initialized Brazilian Portuguese parser
result, err := when.BR.Parse("amanhã às 17h", time.Now())

// Use the pre-initialized Dutch parser
result, err := when.NL.Parse("morgen om 17:00", time.Now())
```

**Available pre-initialized parsers:**
- `when.EN` - English (includes `en.All` + `common.SlashDMY` with `Skip` strategy)
- `when.RU` - Russian (includes `ru.All` + `common.All`)
- `when.BR` - Brazilian Portuguese (includes `br.All` + `common.All`)
- `when.NL` - Dutch (includes `nl.All` + `common.All`)

These are initialized in the `init()` function in `when.go` (lines 169-185).

### 2. **Custom Parser with Manual Rule Selection** (For multi-language or custom rules)

Create your own parser and add specific rules:

```go
import (
    "github.com/runableapp/when-ng"
    "github.com/runableapp/when-ng/rules/en"
    "github.com/runableapp/when-ng/rules/common"
)

// Create a new parser
w := when.New(nil)

// Add English rules
w.Add(en.All...)

// Add common rules (like DD/MM/YYYY format)
w.Add(common.All...)

// Parse text
result, err := w.Parse("tomorrow at 5pm", time.Now())
```

**You can mix languages:**
```go
w := when.New(nil)
w.Add(en.All...)      // English rules
w.Add(ru.All...)      // Russian rules
w.Add(common.All...)   // Common rules

// This parser can now handle both English and Russian
result, err := w.Parse("tomorrow at 5pm", time.Now())
result, err := w.Parse("завтра в 5 вечера", time.Now())
```

## How Rules Work

### Rule Structure

Each rule is a `rules.Rule` interface that implements:
- `Find(string) *Match` - Finds matches in the text
- An `Applier` function that extracts date/time components

### Rule Execution Flow

1. **Standard Format Parsing** (NEW): First tries Go's built-in time parsing for ISO 8601, RFC3339, etc.
2. **Rule Matching**: All rules in the parser are checked against the input text
3. **Clustering**: Matches within `Distance` (default: 5 characters) are grouped together
4. **Rule Application**: Matched rules are applied in order (by definition or match order)
5. **Context Building**: Rules set values in a `Context` (Year, Month, Day, Hour, Minute, etc.)
6. **Time Construction**: Final time is built from the context

### Rule Strategies

Rules can use different strategies when applied:

- **`rules.Override`** (default): Overwrites existing values in context
- **`rules.Merge`**: Merges with existing values (not fully implemented)
- **`rules.Skip`**: Skips if context already has values

Example from `en.All`:
```go
var All = []rules.Rule{
    Weekday(rules.Override),      // Override strategy
    CasualDate(rules.Override),
    ISODate(rules.Override),
    // ...
}
```

## Language-Specific Rule Sets

Each language has an `All` variable containing all its rules:

- **`en.All`** - English rules (Weekday, CasualDate, HourMinute, etc.)
- **`ru.All`** - Russian rules
- **`br.All`** - Brazilian Portuguese rules
- **`nl.All`** - Dutch rules
- **`common.All`** - Common rules (like DD/MM/YYYY format)

## Examples

### Example 1: Single Language (English)
```go
// Option A: Use pre-initialized parser
result, err := when.EN.Parse("next wednesday at 2:25 p.m.", time.Now())

// Option B: Create custom parser
w := when.New(nil)
w.Add(en.All...)
w.Add(common.All...)
result, err := w.Parse("next wednesday at 2:25 p.m.", time.Now())
```

### Example 2: Multi-Language Support
```go
w := when.New(nil)
w.Add(en.All...)   // English
w.Add(ru.All...)   // Russian
w.Add(common.All...)

// Can parse both languages
result1, _ := w.Parse("tomorrow at 5pm", time.Now())           // English
result2, _ := w.Parse("завтра в 5 вечера", time.Now())        // Russian
```

### Example 3: Custom Rule Selection
```go
w := when.New(nil)
w.Add(en.Weekday(rules.Override))      // Only weekday rule
w.Add(en.HourMinute(rules.Override))   // Only hour:minute rule
// Skip other rules like CasualDate, etc.
```

## Important Notes

1. **No Automatic Language Detection**: You must explicitly choose which language rules to use
2. **Rule Order Matters**: Rules are applied in the order they're added (if `MatchByOrder` is true)
3. **Distance Clustering**: Rules must be within 5 characters (default) to be clustered together
4. **Standard Formats First**: ISO/RFC3339 formats are parsed first, before natural language rules
5. **Pre-initialized Parsers**: `when.EN`, `when.RU`, etc. are ready to use and already configured

## Current Test Example

Looking at `tests/test_cases.go`:
```go
w := when.New(nil)
w.Add(en.All...)        // Add all English rules
w.Add(common.All...)    // Add common rules
```

This creates a custom parser with English + common rules, rather than using the pre-initialized `when.EN` parser.
