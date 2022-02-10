package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/auth"
)

var port = flag.Int("port", 3408, "Port number to bind")
var host = flag.String("host", "", "Host or IP to bind")
var blackListFile = flag.String("blacklist", "blacklist", "Blacklist schedule and host pattern file path (auto reload)\n* 13-20 * * 1-5 (?i)youtube blacklists YouTube ignore case betwee 1:00pm and 8:59pm on Weekdays\n")
var whiteListFile = flag.String("whitelist", "whitelist", "Whitelist schedule and host pattern file path (auto reload)\n* * * * 1-2,6,7 (?i)youtube whitelists YouTube on Monday, Tuesday, Saturday and Sunday\n")
var defaultAllow = flag.Bool("default-allow", true, "Default allow (or not) when when both whitelist & blacklist are present")
var helpFlag = flag.Bool("h", false, "Show this help")
var userFile = flag.String("userlist", "userlist", "Proxy user & password file in user:password (for each line) format (auto reload)")

var secrets *Secrets = nil

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func fileExists(file string) bool {
	stat, err := os.Stat(file)
	if err != nil {
		return false
	}

	if stat.IsDir() {
		return false
	}

	return true
}

func checkAllArgs(args ...string) bool {
	for _, val := range args {
		if val == "" {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()
	if *helpFlag {
		Usage()
		os.Exit(1)
	}

	if *whiteListFile == "whitelist" {
		if !fileExists(*whiteListFile) {
			log.Printf("Default whitelist file %s doesn't exists, not loading", *whiteListFile)
			*whiteListFile = ""
		} else {
			log.Printf("Using default whitelist file %s (Specify by -whitelist xxx)", *whiteListFile)
		}
	}

	if *blackListFile == "blacklist" {
		if !fileExists(*blackListFile) {
			log.Printf("Default blacklist file %s doesn't exists, not loading", *blackListFile)
			*blackListFile = ""
		} else {
			log.Printf("Using default blacklist file %s (Specify by -blacklist xxx)", *blackListFile)
		}
	}

	if *userFile == "userlist" {
		if !fileExists(*userFile) {
			log.Printf("Default userlist file %s doesn't exists, not loading", *userFile)
			*userFile = ""
		} else {
			log.Printf("Using default userlist file %s (Specify by -userlist xxx)", *userFile)
		}
	}

	if *blackListFile != "" && *whiteListFile == "" {
		*defaultAllow = true
	}
	if *blackListFile == "" && *whiteListFile != "" {
		*defaultAllow = false
	}

	bind := fmt.Sprintf("%s:%d", *host, *port)

	log.Println("Proxy bind:", "["+bind+"] (Specify by -host xxxx -port yyyy)")
	if *blackListFile != "" {
		fi, err := os.Stat(*blackListFile)
		if err != nil {
			log.Fatalf("File can't be read:%s", *blackListFile)
		}
		log.Printf("Blacklist file is `%s`, size %d", *blackListFile, fi.Size())
	} else {
		log.Printf("No blacklist rule file specified (Specify by -blacklist xxxx)")
	}

	if *whiteListFile != "" {
		fi, err := os.Stat(*whiteListFile)
		if err != nil {
			log.Fatalf("File can't be read: %s", *whiteListFile)
		}
		log.Printf("Whitelist file is `%s`, size %d", *whiteListFile, fi.Size())
	} else {
		log.Printf("No whitelist rule specified (-whitelist xxx)")
	}

	if *userFile != "" {
		log.Printf("Using Proxy Authentication userlist file from `%s`", *userFile)
		secrets = LoadSecretsFrom(*userFile)
	} else {
		log.Printf("No userlist file specified, disabling proxy authentication. (Specify by -userlist xxxx)")
	}
	defaultAction := "Accepted"
	if !*defaultAllow {
		defaultAction = "Rejected"
	}
	log.Printf("When no whitelist/blacklist rules matched, connection will be %s (Specify by cmd argument -default-allow true|false)", defaultAction)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false

	if secrets != nil {
		log.Printf("Enabling Proxy Authentication...")
		auth.ProxyBasic(proxy, "goproxy Authentication", func(user, pwd string) bool {
			if secrets.UserCount() > 0 {
				if secrets.Authenticate(user, pwd) {
					return true
				} else {
					log.Printf("Denied proxy request: Invalid username and password.")
					return false
				}
			} else {
				log.Printf("Ignored authentication since no valid users found!")
				return true
			}
		})
	} else {
		log.Printf("Disabling Proxy Authentication...")
	}

	var blackListRuleSet *RuleSet = nil
	var whiteListRuleSet *RuleSet = nil

	if *blackListFile != "" {
		blackListRuleSet = LoadRuleSetFrom(*blackListFile)
	}

	if *whiteListFile != "" {
		whiteListRuleSet = LoadRuleSetFrom(*whiteListFile)
	}

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			now := time.Now()
			if blackListRuleSet != nil {
				shouldReject, rule := blackListRuleSet.MatchesTime(now, r.Host)
				if shouldReject {
					log.Printf("%s %s %s Denied by rule %+v", r.RemoteAddr, r.Method, r.URL, rule)
					return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden,
						"This request is not allowed")
				}
			}

			if whiteListRuleSet != nil {
				shouldAccept, rule := whiteListRuleSet.MatchesTime(now, r.Host)
				if shouldAccept {
					log.Printf("%s %s %s Accepted by rule %+v", r.RemoteAddr, r.Method, r.URL, rule)
					return r, nil
				}
			}

			if *defaultAllow {
				log.Printf("%s %s %s Accepted by default", r.RemoteAddr, r.Method, r.URL)
				return r, nil
			} else {
				log.Printf("%s %s %s Denied by default", r.RemoteAddr, r.Method, r.URL)
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden,
					"This request is not allowed")
			}
		})

	var Inspect goproxy.FuncHttpsHandler = func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		now := time.Now()
		if blackListRuleSet != nil {
			shouldReject, rule := blackListRuleSet.MatchesTime(now, host)
			if shouldReject {
				log.Printf("%s %s %s Denied by rule %+v", ctx.Req.RemoteAddr, ctx.Req.Method, host, rule)
				return goproxy.RejectConnect, host
			}
		}

		if whiteListRuleSet != nil {
			shouldAccept, rule := whiteListRuleSet.MatchesTime(now, host)
			if shouldAccept {
				log.Printf("%s %s %s Accepted by rule %+v", ctx.Req.RemoteAddr, ctx.Req.Method, host, rule)
				return goproxy.OkConnect, host
			}

		}
		if *defaultAllow {
			log.Printf("%s %s %s Accepted by default", ctx.Req.RemoteAddr, ctx.Req.Method, host)
			return goproxy.OkConnect, host
		} else {
			log.Printf("%s %s %s Denied by default", ctx.Req.RemoteAddr, ctx.Req.Method, host)
			return goproxy.RejectConnect, host
		}
	}
	proxy.OnRequest().HandleConnect(Inspect)
	log.Fatal(http.ListenAndServe(bind, proxy))
}
