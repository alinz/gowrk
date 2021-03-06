package gowrk

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

func countBytesReader(reader io.Reader) (int64, error) {
	var count int64

	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		switch err {
		case nil:
			fallthrough
		case io.EOF:
			count += int64(n)
			return count, nil
		default:
			return 0, err
		}
	}
}

func calcMax(a, b time.Duration) time.Duration {
	if a < b {
		return b
	}
	return a
}

func calcMin(a, b time.Duration) time.Duration {
	if a > b {
		return b
	}
	return a
}

type request struct {
	id  int
	url string
}

type result struct {
	id         int
	size       int64
	statusCode int
	duration   time.Duration
	err        error
	url        string
	threadID   int
}

type Wrk struct {
	client   *http.Client
	requests chan *request
	results  chan *result
}

func (w *Wrk) sendRequest(request *request) *result {
	result := &result{}
	resp, err := http.Get(request.url)
	if err != nil {
		result.err = err
		return result
	}

	size, err := countBytesReader(resp.Body)
	if err != nil {
		result.err = err
		return result
	}

	result.statusCode = resp.StatusCode
	result.size = size

	return result
}

func printMap(arr []map[string]interface{}, writer io.Writer) {
	to := tabwriter.NewWriter(writer, 0, 0, 3, ' ', tabwriter.TabIndent)
	for _, table := range arr {
		for key, value := range table {
			fmt.Fprintf(to, "%s:\t%v\n", key, value)
		}
	}
	to.Flush()
}

func Start(targetURL string, c, n int, unique, dump bool) {
	var dumpWriter *tabwriter.Writer

	if dump {
		dumpFile, err := os.Create("./dump.csv")
		dumpWriter = tabwriter.NewWriter(dumpFile, 0, 0, 3, ' ', tabwriter.TabIndent)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			dumpWriter.Flush()
			dumpFile.Close()
		}()
	}

	wrk := &Wrk{
		client:   &http.Client{},
		requests: make(chan *request, 1),
		results:  make(chan *result, c),
	}

	var wg sync.WaitGroup
	var wg2 sync.WaitGroup

	i := 0
	for i < c {
		wg.Add(1)
		wg2.Add(1)
		go func(id int) {
			defer wg.Done()
			defer wg2.Done()

			for request := range wrk.requests {
				start := time.Now()
				result := wrk.sendRequest(request)
				result.duration = time.Since(start)
				result.id = request.id
				result.url = request.url
				result.threadID = id
				wrk.results <- result
			}
		}(i)
		i++
	}

	go func() {
		wg2.Wait()
		close(wrk.results)
	}()

	var avgDuration int64
	var avgSize int64
	var errors int64
	var totalTime time.Duration
	var max time.Duration
	var min time.Duration

	wg.Add(1)
	go func() {
		defer wg.Done()
		var i int64
		var errorMessage string

		start := time.Now()

		if dump {
			fmt.Fprintf(dumpWriter, "id,\tthread,\tDuration,\tSize,\tStatus Code,\terror\n")
		}

		for result := range wrk.results {
			errorMessage = ""

			if result.err != nil {
				errors++
				errorMessage = result.err.Error()
			} else {
				if min == 0 {
					min = result.duration
				}

				max = calcMax(max, result.duration)
				min = calcMin(min, result.duration)
				avgDuration += int64(result.duration)
				avgSize += result.size
				i++
			}

			if dump {
				fmt.Fprintf(dumpWriter, "%d,\t%d,\t%s,\t%d,\t%d,\t%s\n", result.id, result.threadID, result.duration, result.size, result.statusCode, errorMessage)
			}
		}

		if i > 0 {
			avgDuration = avgDuration / i
			avgSize = avgSize / i
		}

		totalTime = time.Since(start)
	}()

	go func() {
		url, err := url.Parse(targetURL)
		if err != nil {
			log.Fatal(err)
		}

		query := url.Query()

		i := 0
		for i < n {
			fmt.Printf("\rProcessing: %%%d", int(float32(i)/float32(n)*100))
			if unique {
				query.Set("__", fmt.Sprintf("%d", time.Now().UnixNano()))
				url.RawQuery = query.Encode()
			}
			wrk.requests <- &request{id: i + 1, url: url.String()}
			i++
		}
		fmt.Printf("\rFinished sending requests\n\n")
		close(wrk.requests)
	}()

	wg.Wait()

	var output bytes.Buffer
	printMap([]map[string]interface{}{
		map[string]interface{}{"Concurrent": c},
		map[string]interface{}{"Request": n},
		map[string]interface{}{"URL": targetURL},
		map[string]interface{}{"Total time": totalTime},
		map[string]interface{}{"Min Duration": min},
		map[string]interface{}{"Max Duration": max},
		map[string]interface{}{"Average Duration": time.Duration(avgDuration)},
		map[string]interface{}{"Average Size": fmt.Sprintf("%d bytes", avgSize)},
		map[string]interface{}{"Errors": errors},
	}, &output)

	fmt.Println(string(output.Bytes()))
}
