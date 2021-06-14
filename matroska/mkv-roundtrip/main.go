package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/at-wat/ebml-go"
	"github.com/at-wat/ebml-go/webm"
)

func main() {
	r, err := os.Open("test.mkv")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	var ret struct {
		Header  webm.EBMLHeader `ebml:"EBML"`
		Segment webm.Segment    `ebml:"Segment"`
	}
	if err := ebml.Unmarshal(r, &ret); err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	j, err := json.MarshalIndent(ret, "", "  ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("%s\n", string(j))

}
