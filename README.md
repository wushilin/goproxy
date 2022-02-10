# goproxy
Go Http Proxy with Authentication, Schedule Control, and Portal Control


# Why this tool?

You may need to restrict my kids's youtube watch time in the after noon so they can focus on learning staff.
This tool allow you to configure time based rule + host based rule for both whitelisting and blacklisting.


# Building

```
$ go get
$ go build
$ ./goproxy -h
Usage of ./goproxy:
  -blacklist string
    	Blacklist schedule and host pattern file path (auto reload)
    	* 13-20 * * 1-5 (?i)youtube blacklists YouTube ignore case betwee 1:00pm and 8:59pm on Weekdays
    	 (default "blacklist")
  -default-allow
    	Default allow (or not) when when both whitelist & blacklist are present (default true)
  -h	Show this help
  -host string
    	Host or IP to bind
  -port int
    	Port number to bind (default 3408)
  -userlist string
    	Proxy user & password file in user:password (for each line) format (auto reload) (default "userlist")
  -whitelist string
    	Whitelist schedule and host pattern file path (auto reload)
    	* * * * 1-2,6,7 (?i)youtube whitelists YouTube on Monday, Tuesday, Saturday and Sunday
    	 (default "whitelist")
```


# Example rules
See examples folder

1. When whitelist is specified but not blacklist is specified, the proxy only allow the whitelisted URLs by 
the time specified, and blocks all other access

2. When blacklist is specified but not whitelist is specified, the proxy only block the blacklisted URLs by 
the time specified, and accepts all other access

3. When both blacklist and whitelists is specified, blocking is checked first, if no decision, whitelist is 
checked, if still no deciison, the -default-allow flag is applied

4. You can specify proxy user & password in the userlist file

5. All Config files can be updated on the fly, and will be loaded within 5 seconds on next request. You 
don't need to restart the proxy.

6. All non-HTTPs requests URLs are printed & logged, while HTTPs requests only log host & port
