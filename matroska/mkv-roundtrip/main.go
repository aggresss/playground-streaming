package main

import (
	"flag"
	"fmt"
)

var (
	input = flag.String("input", "", "input file")
)

func readTrackInfo() error {
	return nil
}

func main() {
	flag.Parse()

	if err := readTrackInfo(); err != nil {
		fmt.Println(err.Error())
	}
}
