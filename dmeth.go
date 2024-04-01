package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
==========================================================
dMeth - Concurrently discover allowed HTTP methods in URLs
==========================================================
     ==> https://github.com/prasant-paudel/dmeth <==

`
const usage = `USAGE:
dmeth [OPTIONS] [TARGET]

OPTIONS:
-s  string 	Allowed status codes
-b  string 	Blocked status codes (default=405,501)
-v       	Verbose (show all responses)
-h       	Show help menu 

EXAMPLES:
1. dmeth https://example.com
2. dmeth -s 200,301 -m post,delete target_urls.txt
3. echo "https://example.com" | dmeth -b 405 -m post
4. cat target_urls.txt | dmeth
`

var ch = make(chan string)
var methods = []string{
	"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS",
	"TRACE", "PATCH",
}

var (
	targetUrls         []string
	allowedStatusCodes string
	blockedStatusCodes string
	allowedMethods     string
	verbose            bool
	helpFlag           bool
)

type Response struct {
	Method     string
	URL        string
	StatusCode int
	Error      error
}

func parseArguments() {
	flag.StringVar(&allowedStatusCodes, "s", "", "Allowed status codes")
	flag.StringVar(&blockedStatusCodes, "b", "405,501", "Blocked status codes")
	flag.StringVar(&allowedMethods, "m", "all", "Allowed HTTP methods to look for")
	flag.BoolVar(&verbose, "v", false, "Blocked status codes")
	flag.BoolVar(&helpFlag, "h", false, "Show this help menu")
	flag.Parse()

	if allowedStatusCodes != "" {
		blockedStatusCodes = ""
	}

	targetUrls = getTargets()

	if helpFlag || len(targetUrls) == 0 {
		fmt.Print(colorPurple, banner, colorReset)
		fmt.Print(usage)
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
	parseArguments()
	fmt.Print(colorPurple, banner, colorReset)

	s := ParseStatusCodes(allowedStatusCodes)
	b := ParseStatusCodes(blockedStatusCodes)

	if len(b) > 0 && len(s) > 0 {
		log.Fatalln("[!] Flags -s and -b can not be used at the same time.")
	}

	resCh := make(chan Response)

	for _, m := range methods {
		requests := CreateRequests(m, targetUrls)
		go SendRequests(requests, resCh)
	}

	for range len(methods) * len(targetUrls) {
		res := <-resCh
		if verbose || containsInt(s, res.StatusCode) || (len(s) == 0 && !containsInt(b, res.StatusCode)) {
			fmt.Print(res.String())
		}
	}
}

func CreateRequests(method string, urls []string) (requests []*http.Request) {
	for _, url := range targetUrls {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Printf("[!] Error creating '%s' request for %s", method, url)
			continue
		}
		requests = append(requests, req)
	}
	return
}

func SendRequests(requests []*http.Request, outCh chan<- Response) {
	client := &http.Client{}

	for _, req := range requests {
		res, err := client.Do(req)
		if err != nil {
			outCh <- Response{
				Error: fmt.Errorf("[!] Error sending '%s' request for %s", req.Method, req.URL),
			}
		}
		outCh <- Response{
			Method:     res.Request.Method,
			URL:        res.Request.URL.String(),
			StatusCode: res.StatusCode,
		}
	}
}

func (r *Response) MethodColor() string {
	if r.Method == "GET" || r.Method == "HEAD" {
		return colorCyan
	} else if r.Method == "POST" {
		return colorGreen
	} else if r.Method == "DELETE" {
		return colorRed
	} else if r.Method == "PUT" || r.Method == "PATCH" || r.Method == "UPDATE" {
		return colorYellow
	} else if r.Method == "OPTIONS" {
		return colorPurple
	}
	return colorReset
}

func (res Response) String() string {
	if res.Error != nil {
		return res.Error.Error()
	}
	m := res.Method
	output := fmt.Sprintln("[+] "+m+strings.Repeat(" ", 8-len(m))+":", res.StatusCode, " | URL: ", res.URL)
	output = res.MethodColor() + output + colorReset
	return output
}

func getTargets() []string {
	// From stdin
	f, _ := os.Stdin.Stat()
	if f.Mode()&os.ModeCharDevice == 0 {
		return ReadFileByLines(os.Stdin.Name())
	}

	// From argument
	if len(os.Args) > 1 {
		t := os.Args[len(os.Args)-1]
		// If URL
		if strings.HasPrefix(t, "http") {
			return []string{t}
		}
		// If file
		return ReadFileByLines(t)
	}

	return nil
}

func ParseStatusCodes(s string) []int {
	var out []int
	if s != "" {
		splittedCodes := strings.Split(s, ",")
		for _, code := range splittedCodes {
			var _i int
			_, err := fmt.Sscan(code, &_i)
			if err != nil {
				log.Fatal("Invalid status code: ", code)
			}
			out = append(out, _i)
		}
	}
	return out
}

func containsInt(arr []int, e int) bool {
	for _, i := range arr {
		if i == e {
			return true
		}
	}
	return false
}

func ReadFileByLines(filePath string) []string {
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
