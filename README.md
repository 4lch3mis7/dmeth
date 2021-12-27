# dMeth - dMeth - A tool to discover allowed HTTP methods in a URL

## Installation
```
go install github.com/prasant-paudel/dmeth@latest
```

## Usage
Flag | Description          
-----|------------
-t   | Target URL   
-T   | LIst of targets
-s   | Allowed status codes
-h   | show help menu  

## Examples
```
dmeth -t https://google.com
```
```
dmeth -T target_urls.txt
```
```
dmeth -t https://google.com -s 200,300
```
```
dmeth -T target_urls.txt -s 200,300
```
