package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pprof "github.com/go-zen-chu/delve-debug-sample/pprof"
)

const loopWaitTime = 50 * time.Millisecond

func insertOdd(ch chan<- int) {
	for i := 1; i <= 100; i++ {
		ch <- 2*i - 1
		time.Sleep(loopWaitTime)
	}
}

func insertEven(ch chan<- int) {
	for i := 1; i <= 100; i++ {
		ch <- 2 * i
		time.Sleep(loopWaitTime)
	}
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	wg := &sync.WaitGroup{}
	ch := make(chan int, 100*100)
	var sb strings.Builder
	// odd number inserter
	wg.Add(1)
	go func() {
		defer wg.Done()
		insertOdd(ch)
	}()
	// even number inserter
	wg.Add(1)
	go func() {
		defer wg.Done()
		insertEven(ch)
	}()

	log.Println("goroutines dispatced")
	wg.Wait()
	log.Println("goroutines finished")
	close(ch)

	for k := range ch {
		sb.WriteString(",")
		sb.WriteString(strconv.Itoa(k))
	}
	fmt.Fprintf(w, "hello delve:%s", sb.String())
}

func Mux() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/foo", fooHandler)
	return m
}

func main() {
	log.Println("start server")

	// handle signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	// run pprof server
	wg.Add(1)
	pprofServer := &http.Server{
		Addr:    ":12001",
		Handler: pprof.Mux(),
	}
	go func() {
		defer wg.Done()
		if err := pprofServer.ListenAndServe(); err != nil {
			log.Printf("pprof server: %s", err)
		}
	}()
	// run application server
	wg.Add(1)
	appServer := &http.Server{
		Addr:    ":12000",
		Handler: Mux(),
	}
	go func() {
		defer wg.Done()
		if err := appServer.ListenAndServe(); err != nil {
			log.Printf("app server: %s", err)
		}
	}()
	// receive signal
	<-sigChan

	ctx := context.Background()
	if err := appServer.Shutdown(ctx); err != nil {
		log.Printf("finish app server: %s", err)
	}
	if err := pprofServer.Shutdown(ctx); err != nil {
		log.Printf("finish pprof server: %s", err)
	}
	wg.Wait()
	log.Println("end server")
}
