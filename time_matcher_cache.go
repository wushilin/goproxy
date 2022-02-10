package main

type TimeMatcherCache struct {
	memory map[string]Matcher
}

func (v TimeMatcherCache) Get(what string) Matcher {
	result := v.memory[what]
	if result != nil {
		return result
	}

	result = ParseMatcher(what)
	v.memory[what] = result
	return result
}

func NewTimeMatcherCache() TimeMatcherCache {
	memory := make(map[string]Matcher)
	return TimeMatcherCache{memory}
}
