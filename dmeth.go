package main

import (
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

var target string
var targets_path string
var allowed_status_codes string
var help_menu string

func parseArguments() {
	flag.StringVar(&target, "t", "", "Target URL")
	flag.StringVar(&targets_path, "T", "", "List of targets [File]")
	flag.StringVar(&allowed_status_codes, "s", "200", "Allowed status codes (default=200)")
	flag.StringVar(&help_menu, "h", "", "Show this help menu")

	flag.Parse()

	if target == "" && targets_path == "" {
		fmt.Print(banner)
		flag.Usage()
		os.Exit(0)
	}

}

func main() {
	parseArguments()
	enumMethods()
}

func enumMethods() {
	methods := []string{"GET", "POST", "HEAD", "OPTIONS", "PUT", "PATCH", "TRACE", "DELETE", "CONNECT"}

	// Split status codes seperated by ","
	splittedCodes := strings.Split(allowed_status_codes, ",")

	// Parse the splitted status codes into integers
	var whitelist []int
	for i := 0; i < len(splittedCodes); i++ {
		var code int
		_, err := fmt.Sscan(splittedCodes[i], &code)
		if err != nil {
			log.Fatal("Invalid status code: ", splittedCodes[i])
		}
		whitelist = append(whitelist, code)
	}

	// Iterate over list of methods
	for i := 0; i < len(methods); i++ {
		resp := req(methods[i], target)
		if containsInt(whitelist, resp.StatusCode) {
			fmt.Println("[+] ", methods[i], "\t:\t", resp.StatusCode, " --> ", target)
		}
	}
}

func req(method string, url string) *http.Response {
	method = strings.ToUpper(method)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response. ", err)
	}

	return resp

}

func containsInt(arr []int, e int) bool {
	for _, i := range arr {
		if i == e {
			return true
		}
	}
	return false
}
