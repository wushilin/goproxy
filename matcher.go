package main

import (
	"log"
	"strconv"
	"strings"
)

type Matcher interface {
	matches(what int) bool
}

type WildcardMatcher struct {
}

func (v WildcardMatcher) matches(what int) bool {
	return true
}

type SingleMatcher struct {
	value int
}

func (v SingleMatcher) matches(what int) bool {
	return v.value == what
}

type RangeMatcher struct {
	low  int
	high int
}

func (v RangeMatcher) matches(what int) bool {
	return v.low <= what && v.high >= what
}

type AndMatcher struct {
	matcher1 Matcher
	matcher2 Matcher
}

func (v AndMatcher) matches(what int) bool {
	return v.matcher1.matches(what) && v.matcher2.matches(what)
}

type OrMatcher struct {
	matcher1 Matcher
	matcher2 Matcher
}

func (v OrMatcher) matches(what int) bool {
	return v.matcher1.matches(what) || v.matcher2.matches(what)
}

type Composer struct {
	matcher Matcher
}

func (v Composer) Or(matcher Matcher) Composer {
	if v.matcher == nil {
		return Composer{matcher: matcher}
	}
	if matcher == nil {
		return Composer{matcher: v.matcher}
	}

	vMatcher := OrMatcher{v.matcher, matcher}
	return Composer{matcher: vMatcher}
}
func (v Composer) And(matcher Matcher) Composer {
	if v.matcher == nil {
		return Composer{matcher: matcher}
	}
	if matcher == nil {
		return Composer{matcher: v.matcher}
	}

	vMatcher := AndMatcher{v.matcher, matcher}
	return Composer{matcher: vMatcher}
}

func (v Composer) matches(what int) bool {
	return v.matcher.matches(what)
}

func NewComposer() Composer {
	return Composer{matcher: nil}
}

func ParseMatcher(rule string) Matcher {
	if rule == "*" {
		return WildcardMatcher{}
	}

	composer := NewComposer()
	tokens := strings.Split(rule, ",")
	for _, next := range tokens {
		index := strings.Index(next, "-")
		if index == -1 {
			composer = composer.Or(SingleMatcher{value: parseInt(next)})
		} else {
			rangeTokens := strings.Split(next, "-")
			if len(rangeTokens) != 2 {
				log.Fatal("Invalid range:", next)
			}
			lower, upper := parseInt(rangeTokens[0]), parseInt(rangeTokens[1])
			composer = composer.Or(RangeMatcher{low: lower, high: upper})
		}
	}
	return composer
}

func parseInt(what string) int {
	result, err := strconv.Atoi(what)
	if err != nil {
		log.Fatal("Invalid number:", what)
	}
	return result
}
