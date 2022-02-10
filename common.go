package main

func indexAny(input string, what string) int {
	for i := 0; i < len(input); i++ {
		for j := 0; j < len(what); j++ {
			if input[i] == what[j] {
				return i
			}
		}
	}
	return -1
}
