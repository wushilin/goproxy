package main

import (
	"bufio"
	"bytes"
	b64 "encoding/base64"
	"log"
	"strings"
)

type Secrets struct {
	cache *CachedFileContent
	data  map[string]string
}

func (v *Secrets) Reload() {
	dataBytes, changed := v.cache.Get()
	if !changed {
		return
	}
	dataBytesReader := bytes.NewReader(dataBytes)
	v.data = make(map[string]string)
	scanner := bufio.NewScanner(dataBytesReader)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		if line[0] == '#' || line[:2] == "//" || line[0] == ';' {
			continue
		}
		index := indexAny(line, " :")
		if index == -1 {
			log.Printf("Skipped invalid line %s", line)
			continue
		}
		username := strings.TrimSpace(line[:index])
		password := strings.TrimSpace(line[index+1:])

		v.data[username] = password
		log.Printf("Loaded user %s password %s", username, maskPassword(password))
	}
	log.Printf("Loaded %d users from userlist file %s", len(v.data), v.cache.FileName())

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func maskPassword(what string) string {
	resultarr := make([]rune, len(what))
	for i := 0; i < len(what); i++ {
		resultarr[i] = '*'
	}
	return string(resultarr)
}

func (v *Secrets) Authenticate(username, password string) bool {
	v.Reload()
	if actualPassword, ok := v.data[username]; ok {
		return actualPassword == password
	} else {
		return false
	}

}

func (v *Secrets) UserCount() int {
	v.Reload()
	return len(v.data)
}

func (v *Secrets) AuthenticateRaw(userinfo string) bool {
	bytes, err := b64.StdEncoding.DecodeString(userinfo)
	if err != nil {
		return false
	}
	userinfostr := string(bytes)
	index := strings.Index(userinfostr, ":")
	if index == -1 {
		return false
	}

	username := strings.TrimSpace(userinfostr[:index])
	password := strings.TrimSpace(userinfostr[index+1:])
	return v.Authenticate(username, password)
}

func LoadSecretsFrom(file string) *Secrets {
	v := &Secrets{NewCachedFile(file), make(map[string]string)}
	v.Reload()
	return v
}
