package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

type (
	ParserOptions struct {
		Timeout  time.Duration
		Capacity int32
	}

	Parser struct {
		done     chan bool
		finish   chan int64
		capacity int32
		current  int32
		client   http.Client
		Result   int64
	}
)

func NewParser(opts *ParserOptions) *Parser {

	return &Parser{
		done:     make(chan bool, 0),
		finish:   make(chan int64, 0),
		capacity: opts.Capacity,
		client: http.Client{
			Timeout: opts.Timeout,
		},
	}

}

func (p *Parser) request(url string) (data []byte, err error) {

	resp, err := p.client.Get(url)
	if err != nil {
		return
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return

}

func (p *Parser) Finish() int64 {
	return <-p.finish
}

func (p *Parser) Count(url, substr string) {

	if atomic.AddInt32(&p.current, 1) > p.capacity {
		<-p.done
	}

	go func(p *Parser) {

		data, err := p.request(url)
		if err != nil {
			fmt.Println(err)
		}

		c := bytes.Count(data, []byte(substr))
		atomic.AddInt64(&p.Result, int64(c))

		fmt.Printf("Count for %s: %d\n\n", url, c)

		if atomic.AddInt32(&p.current, -1) == 0 {
			p.finish <- atomic.LoadInt64(&p.Result)
		}

		p.done <- true

	}(p)

}

func main() {

	opts := &ParserOptions{
		Timeout:  time.Duration(2 * time.Second),
		Capacity: 2,
	}

	parser := NewParser(opts)

	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}

			break

		}

		parser.Count(text[0:len(text)-1], "Go")

	}

	total := parser.Finish()
	fmt.Println("Total:", total)

}
