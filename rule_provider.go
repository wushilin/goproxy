package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"regexp"
	"strings"
	"time"
)

var timeMatcherRuleCache = NewTimeMatcherCache()

type Rule struct {
	minute           string
	hour             string
	dayOfMonth       string
	month            string
	dayOfWeek        string
	hostPattern      string
	hostPatternRegex *regexp.Regexp
}

type RuleSet struct {
	cache *CachedFileContent
	data  []Rule
}

func parseListsFromSource(dataSource io.Reader) []Rule {
	result := make([]Rule, 0)
	scanner := bufio.NewScanner(dataSource)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line[0] == '#' {
			continue
		}
		if len(line) == 0 {
			continue
		}
		minute, liner := extractToken(line)
		hour, liner := extractToken(liner)
		dayOfMonth, liner := extractToken(liner)
		month, liner := extractToken(liner)
		dayOfWeek, pattern := extractToken(liner)
		if !checkAllArgs(minute, hour, dayOfMonth, month, dayOfWeek, pattern) {
			log.Fatalf("Invalid rule line %s", line)
		}
		result = append(result, Rule{
			minute:           minute,
			hour:             hour,
			dayOfMonth:       dayOfMonth,
			month:            month,
			dayOfWeek:        dayOfWeek,
			hostPattern:      pattern,
			hostPatternRegex: regexp.MustCompile(pattern),
		})
	}
	return result
}

func (v *RuleSet) internalGet() []Rule {
	v.Reload()
	return v.data
}

func getNow(currentTime time.Time) []int {
	year, month, day := currentTime.Year(), currentTime.Month(), currentTime.Day()
	hr, min, sec := currentTime.Clock()
	dayOfWeek := currentTime.Weekday()
	return []int{year, int(month), day, hr, min, sec, int(dayOfWeek)}
}

func (v *RuleSet) MatchesTime(when time.Time, host string) (bool, Rule) {
	now := getNow(when)
	//var nowYear = now[0]
	var nowMonth = now[1]
	var nowDay = now[2]
	var nowHour = now[3]
	var nowMinute = now[4]
	//var nowSecond = now[5]
	var nowDayOfWeek = now[6]
	for _, nextRule := range v.internalGet() {
		if matches(nowMinute, nextRule.minute) && matches(nowHour, nextRule.hour) && matches(nowDay, nextRule.dayOfMonth) &&
			matches(nowMonth, nextRule.month) && matches(nowDayOfWeek, nextRule.dayOfWeek) {
			// now the time condition is matching
			if nextRule.hostPatternRegex.MatchString(host) {
				return true, nextRule
			}
		}
	}
	return false, Rule{}
}

func matches(what int, rule string) bool {
	matcher := timeMatcherRuleCache.Get(rule)
	return matcher.matches(what)
}

func (v *RuleSet) Reload() bool {
	dataBytes, changed := v.cache.Get()
	if !changed {
		return false
	}
	dataBytesReader := bytes.NewReader(dataBytes)
	v.data = parseListsFromSource(dataBytesReader)
	return true
}

func LoadRuleSetFrom(file string) *RuleSet {
	v := &RuleSet{NewCachedFile(file), nil}
	v.Reload()
	return v
}

func extractToken(input string) (string, string) {
	if len(input) == 0 {
		return "", ""
	}

	index := indexAny(input, " \t")
	if index == -1 {
		return input, ""
	}

	return input[:index], strings.TrimSpace(input[index+1:])
}
