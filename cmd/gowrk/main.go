package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/alinz/gowrk"
)

const usagestr = `
Usage: gowrk --url <url> [options]

gowork Options:
	--concurrent <value>  number of concurrent connections (default 1)
	--request <value>     number of total requests (default 1)
`

func usage() {
	fmt.Printf("%s\n", usagestr)
	os.Exit(0)
}

func main() {
	var url string
	var concurrent int
	var request int

	flag.StringVar(&url, "url", "", "full qualified url")
	flag.IntVar(&concurrent, "concurrent", 1, "number of concurrent connections")
	flag.IntVar(&request, "request", 1, "number of total requests")

	flag.Parse()

	if url == "" {
		usage()
	}

	gowrk.Start(url, concurrent, request)
}
