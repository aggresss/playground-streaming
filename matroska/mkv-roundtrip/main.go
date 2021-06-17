package main

import (
	"fmt"
	"os"

	"github.com/at-wat/ebml-go/mkvcore"
)

func main() {
	r, err := os.Open("test.mkv")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	rs, err := mkvcore.NewSimpleBlockReader(r)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	for _, r := range rs {
		fmt.Printf("%#v\n", r.TrackEntry())
	}
}
