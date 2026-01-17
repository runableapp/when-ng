
# when-ng

when-ng is a fork of https://github.com/olebedev/when, tailored for use in Ruanble.app projects.

* Upstream commit used: `efeef445de938fe5bbaac704d9d1d9b69d66216d`
* Upstream branch used: `master`
* This was forked on: 2025-01-16

**Note:** 
- ⚠️ This repository is **not accepting pull requests (PRs)**.  
We welcome [issues and support requests](https://github.com/runableapp/when-ng/issues), but this fork is *not intended for upstream merging* or general open-source collaboration.  
- It is customized primarily for use within [Runable.app](https://runable.app/) projects.
- Support for other languages (Dutch: `nl`, Russian: `ru`, Chinese: `zh`, Brazilian Portuguese: `br`) is available *as is*, reflecting the state of those rules as of the last forked date. We are unable to accept feature requests or make improvements for these non-English language rules.


## Fork changes

This fork extends English parsing with the following additions:

1. **Year normalization**: "last year" and "next year" now normalize to January 1st, 00:00:00 of that year (instead of preserving the current month/day)
   - Example: `"last year"` → January 1st of last year at midnight

2. **Last weekday support**: Added support for "last <weekday>" phrases (e.g., "last monday", "last friday")
   - Example: `"last monday"` → The most recent Monday in the past

3. **MM/DD/YYYY date format**: Added `SlashMDY` rule to parse American-style slash dates
   - Example: `"5/20/2026"` → May 20th, 2026
   - Example: `"5/20/2026 at 23:42"` → May 20th, 2026 at 11:42 PM
   - Note: This takes precedence over the common DD/MM/YYYY format for English rules

4. **Standard date/time format parsing** (Common rule): The parser now tries Go's built-in date/time parsing first for standard formats (ISO 8601, RFC3339, etc.) before falling back to natural language parsing. The `ISODate` rule is available in common rules for all languages.
   - Example: `"2020-05-22T15:55-04:00"` → Parsed as RFC3339 format (2020-05-22T15:55:00-04:00)
   - Example: `"2026-01-16 04:00:00 AM"` → Parsed as ISO format with AM/PM (2026-01-16T04:00:00-05:00)
   - Supported formats include: RFC3339, RFC3339Nano, ISO 8601 (with/without timezone, with/without AM/PM)
   - If standard format parsing fails, the parser falls back to natural language parsing rules
   - Note: This is a common rule (available to all languages), not English-specific


## Examples

- **tonight at 11:10 pm**
- at **Friday afternoon**
- the deadline is **next tuesday 14:00**
- drop me a line **next wednesday at 2:25 p.m**
- it could be done at **11 am past tuesday**
- **last year**
- **last monday**  
  - _Note: For phrases like **"last monday"**, the parser interprets this as the Monday of the previous week. For example, if today is Friday, "last monday" refers to the Monday from the prior week—not the current week. If you want this week's Monday (the most recent Monday), use "this monday" or "past monday" instead. This behavior matches common interpretations of "last <weekday>" in English, meaning the weekday in the week before the current one._


Check [EN](https://github.com/olebedev/when/blob/master/rules/en) rules and tests of them, for more examples.

**Needed rule not found?**
Open [an issue](https://github.com/runableapp/when-nr/issues/new) with the case and it will be added asap.

### How it works

Usually, there are several rules added to the parser's instance for checking. Each rule has its own borders - length and offset in provided string. Meanwhile, each rule yields only the first match over the string. So, the library checks all the rules and extracts a cluster of matched rules which have distance between each other less or equal to [`options.Distance`](https://github.com/olebedev/when/blob/master/when.go#L141-L144), which is 5 by default. For example:

```
on next wednesday at 2:25 p.m.
   └──────┬─────┘    └───┬───┘
       weekday      hour + minute
```

So, we have a cluster of matched rules - `"next wednesday at 2:25 p.m."` in the string representation.

After that, each rule is applied to the context. In order of definition or in match order, if [`options.MatchByOrder`](https://github.com/olebedev/when/blob/master/when.go#L141-L144) is set to `true`(which it is by default). Each rule could be applied with given merge strategy. By default, it's an [Override](https://github.com/olebedev/when/blob/master/rules/rules.go#L13) strategy. The other strategies are not implemented yet in the rules. **Pull requests are welcome.**

### Supported Languages

- [EN](https://github.com/olebedev/when/blob/master/rules/en) - English

### ⚠️ Legacy Supported Languages 
(no further improvements planned)
- [RU](https://github.com/olebedev/when/blob/master/rules/ru) - Russian
- [BR](https://github.com/olebedev/when/blob/master/rules/br) - Brazilian Portuguese
- [ZH](https://github.com/olebedev/when/blob/master/rules/zh) - Chinese
- [NL](https://github.com/olebedev/when/blob/master/rules/nl) - Dutch

### Install

The project follows the official [release workflow](https://go.dev/doc/modules/release-workflow). It is recommended to refer to this resource for detailed information on the process.

To install the latest version:

```
$ go get github.com/runableapp/when-ng@latest
```

### Usage

See `tests/test_cases.go` also.




```go
w := when.New(nil)
w.Add(en.All...)
w.Add(common.All...)

text := "drop me a line in next wednesday at 2:25 p.m"
r, err := w.Parse(text, time.Now())
if err != nil {
	// an error has occurred
}
if  r == nil {
 	// no matches found
}

fmt.Println(
	"the time",
	r.Time.String(),
	"mentioned in",
	text[r.Index:r.Index+len(r.Text)],
)
```

#### Distance Option

```go
w := when.New(nil)
w.Add(en.All...)
w.Add(common.All...)

text := "February 23, 2019 | 1:46pm"

// With default distance (5):
// February 23, 2019 | 1:46pm
//            └───┬───┘
//           distance: 9 (1:46pm will be ignored)

r, _ := w.Parse(text, time.Now())
fmt.Printf(r.Time.String())
// "2019-02-23 09:21:21.835182427 -0300 -03"
// 2019-02-23 (correct)
//   09:21:21 ("wrong")

// With custom distance (10):
w.SetOptions(&rules.Options{
	Distance:     10,
	MatchByOrder: true})

r, _ = w.Parse(text, time.Now())
fmt.Printf(r.Time.String())
// "2019-02-23 13:46:21.559521554 -0300 -03"
// 2019-02-23 (correct)
//   13:46:21 (correct)
```

### State of the project

This project is maintained primarily to support the needs of Runable.app, and will be updated as our requirements evolve. We strive to accommodate any issues reported or enhancement requests submitted through GitHub issues.

### TODO

- [ ] language `en`: rules for [these examples](https://github.com/mojombo/chronic#examples).  See `tests/test_cases.txt`


### LICENSE

http://www.apache.org/licenses/LICENSE-2.0
