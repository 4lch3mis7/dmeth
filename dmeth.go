package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const colorReset = "\033[0m"

const colorRed = "\033[31m"
const colorGreen = "\033[32m"
const colorYellow = "\033[33m"
const colorBlue = "\033[34m"
const colorPurple = "\033[35m"
const colorCyan = "\033[36m"
const colorWhite = "\033[37m"

const bgYellow = "\033[43m"

const banner = `
          dP 8888ba.88ba             dP   dP       
          88 88  '8b  '8b            88   88       
    .d888b88 88   88   88 .d8888b. d8888P 88d888b. 
    88'  '88 88   88   88 88ooood8   88   88'  '88 
    88.  .88 88   88   88 88.  ...   88   88    88 
    '88888P8 dP   dP   dP '88888P'   dP   dP    dP 
========================================================
dMeth - A tool to discover allowed HTTP methods in a URL
========================================================
    ==> https://github.com/prasant-paudel/dmeth <==

`
const examples = `
Examples:
dmeth -t https://google.com
dmeth -T target_urls.txt
dmeth -t https://google.com -s 200,300
dmeth -T target_urls.txt -m post,delete
`

var ch = make(chan string)
var methods = []string{
	"GET", "POST", "HEAD", "OPTIONS",
	"PUT", "PATCH", "UPDATE", "TRACE", "DELETE",
}

var target string
var targetsPath string
var allowedStatusCodes string
var helpFlag bool
var allowedMethods string

func parseArguments() {
	flag.StringVar(&target, "t", "", "Target URL")
	flag.StringVar(&targetsPath, "T", "", "List of targets [File]")
	flag.StringVar(&allowedStatusCodes, "s", "200", "Allowed status codes")
	flag.StringVar(&allowedMethods, "m", "all", "Allowed HTTP methods to look for")
	flag.BoolVar(&helpFlag, "h", false, "Show this help menu")

	flag.Parse()

	if helpFlag || (target == "" && targetsPath == "") {
		fmt.Print(colorPurple, banner, colorReset)
		flag.Usage()
		fmt.Print(examples)
		os.Exit(0)
	}

	if allowedMethods != "all" {
		var _meths []string
		for _, m := range strings.Split(allowedMethods, ",") {
			m = strings.TrimSpace(m)
			m = strings.ToUpper(m)
			_meths = append(_meths, m)
		}
		methods = _meths
	}
}

func main() {
	var targetUrls []string

	parseArguments()

	// Single target
	if target != "" {
		targetUrls = append(targetUrls, target)
	}
	// Multiple targets
	if targetsPath != "" {
		for _, ln := range readLines(targetsPath) {
			targetUrls = append(targetUrls, ln)
		}
	}

	// Parse allowed status codes
	splittedCodes := strings.Split(allowedStatusCodes, ",")
	var allowedStatusCodes []int
	for _, code := range splittedCodes {
		var _i int
		_, err := fmt.Sscan(code, &_i)
		if err != nil {
			log.Fatal("Invalid status code: ", code)
		}
		allowedStatusCodes = append(allowedStatusCodes, _i)
	}

	// Run goroutines to check the status
	for _, url := range targetUrls {
		for _, method := range methods {
			go checkStatus(method, url, allowedStatusCodes)
		}
	}

	// Print the output from channels
	for range targetUrls {
		for range methods {
			fmt.Print(<-ch)
		}
	}

	close(ch)
}

func checkStatus(method string, url string, allowedStatusCodes []int) {
	method = strings.ToUpper(method)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		_e := fmt.Sprintln(colorRed+"[!]", method, " Error reading request. ", err, colorReset)
		// log.Fatal(_e)
		ch <- _e
		return
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		_e := fmt.Sprintln(colorRed+"[!]", method, " Error reading response. ", err, colorReset)
		// log.Fatal(_e)
		ch <- _e
		return
	}

	if containsInt(allowedStatusCodes, resp.StatusCode) {
		output := fmt.Sprintln("[+] "+method+strings.Repeat(" ", 8-len(method))+":", resp.StatusCode, " | URL: ", url)
		if method == "GET" || method == "HEAD" {
			output = colorCyan + output + colorReset
		} else if method == "POST" {
			output = colorGreen + output + colorReset
		} else if method == "DELETE" {
			output = colorRed + output + colorReset
		} else if method == "PUT" || method == "PATCH" || method == "UPDATE" {
			output = colorYellow + output + colorReset
		} else if method == "OPTIONS" {
			output = colorPurple + output + colorReset
		}

		ch <- output
	} else {
		ch <- ""
	}

}

func containsInt(arr []int, e int) bool {
	for _, i := range arr {
		if i == e {
			return true
		}
	}
	return false
}

func readLines(filePath string) []string {
	var lines []string
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	file.Close()
	return lines
}
