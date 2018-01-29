package gowrk

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func countBytesReader(reader io.Reader) (int64, error) {
	var count int64

	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		switch err {
		case nil:
			count += int64(n)
		case io.EOF:
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
	url string
}

type result struct {
	size       int64
	statusCode int
	duration   time.Duration
	err        error
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

	result.size = size

	return result
}

func Start(url string, c, n int) {
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
		go func() {
			defer wg.Done()
			defer wg2.Done()

			for request := range wrk.requests {
				start := time.Now()
				result := wrk.sendRequest(request)
				result.duration = time.Since(start)
				wrk.results <- result
			}
		}()
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

		start := time.Now()

		for result := range wrk.results {
			if result.err != nil {
				errors++
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
		}

		if i > 0 {
			avgDuration = avgDuration / i
			avgSize = avgSize / i
		}

		totalTime = time.Since(start)
	}()

	go func() {
		req := &request{url: url}
		i := 0
		for i < n {
			fmt.Printf("\rCount: %d", i)
			wrk.requests <- req
			i++
		}
		fmt.Printf("\rFinished sending requests\n\n")
		close(wrk.requests)
	}()

	wg.Wait()

	fmt.Println("Concurrent: \t\t", c)
	fmt.Println("Request: \t\t", n)
	fmt.Println("URL: \t\t\t", url)
	fmt.Println("------------------")
	fmt.Println("Total time: \t\t", totalTime)
	fmt.Println("Min Duration: \t\t", min)
	fmt.Println("Max Duration: \t\t", max)
	fmt.Println("Average Duration: \t", time.Duration(avgDuration))
	fmt.Println("Average Size: \t\t", avgSize, "bytes")
	fmt.Println("Errors: \t\t", errors)
}
