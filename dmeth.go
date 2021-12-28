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
	parseArguments()
	enumMethods()
}

func enumMethods() {
	methods := []string{"GET", "POST", "HEAD", "OPTIONS", "PUT", "PATCH", "TRACE", "DELETE"}

	// Split status codes seperated by ","
	splittedCodes := strings.Split(allowedStatusCodes, ",")

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
