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

var ch = make(chan string)
var methods = []string{
	"GET", "POST", "HEAD", "OPTIONS",
	"PUT", "PATCH", "TRACE", "DELETE",
}

var target string
var targetsPath string
var allowedStatusCodes string
var helpFlag bool

func parseArguments() {
	flag.StringVar(&target, "t", "", "Target URL")
	flag.StringVar(&targetsPath, "T", "", "List of targets [File]")
	flag.StringVar(&allowedStatusCodes, "s", "200", "Allowed status codes (default=200)")
	flag.BoolVar(&helpFlag, "h", false, "Show this help menu")

	flag.Parse()

	if helpFlag || (target == "" && targetsPath == "") {
		fmt.Print(banner)
		flag.Usage()
		os.Exit(0)
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
		_e := fmt.Sprintln("[!]", method, " Error reading request. ", err)
		// log.Fatal(_e)
		ch <- _e
		return
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		_e := fmt.Sprintln("[!]", method, " Error reading response. ", err)
		// log.Fatal(_e)
		ch <- _e
		return
	}

	if containsInt(allowedStatusCodes, resp.StatusCode) {
		ch <- fmt.Sprintln("[+] "+method+strings.Repeat(" ", 8-len(method))+":", resp.StatusCode, " | URL: ", url)
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
