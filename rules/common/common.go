package common

import "github.com/runableapp/when-ng/rules"

var All = []rules.Rule{
	ISODate(rules.Override),
	SlashDMY(rules.Override),
}
