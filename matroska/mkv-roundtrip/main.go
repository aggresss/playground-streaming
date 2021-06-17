package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/at-wat/ebml-go/mkvcore"
)

func main() {
	r, err := os.Open("test.mkv")
	if err != nil {
		panic(err)
	}

	rs, err := mkvcore.NewSimpleBlockReader(r)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	for _, r := range rs {
		fmt.Printf("%#v\n", r.TrackEntry())
		rb := r
		go func() {
			for {
				_, _, t, err := rb.Read()
				if err != nil {
					fmt.Println(err.Error())
					break
				}
				fmt.Println(t)
			}
		}()
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ch:
	}

}
