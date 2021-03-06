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
	--concurrent <value>  tnumber of concurrent connections (default 1)
	--request <value>     number of total requests (default 1)
	--unique              atatch timestamp to eeach request to prevent caching
	--dump                dump all data request into csv file
`

func usage() {
	fmt.Printf("%s\n", usagestr)
	os.Exit(0)
}

func main() {
	var url string
	var concurrent int
	var request int
	var unique bool
	var dump bool

	flag.StringVar(&url, "url", "", "full qualified url")
	flag.IntVar(&concurrent, "concurrent", 1, "number of concurrent connections")
	flag.IntVar(&request, "request", 1, "number of total requests")
	flag.BoolVar(&unique, "unique", false, "atatch timestamp to eeach request to prevent caching")
	flag.BoolVar(&dump, "dump", false, "dump all data request into csv file")

	flag.Parse()

	if url == "" {
		usage()
	}

	gowrk.Start(url, concurrent, request, unique, dump)
}
