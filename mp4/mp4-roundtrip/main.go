package main

import (
	"fmt"
	"os"

	"github.com/abema/go-mp4"
)

func main() {
	file, err := os.Open("test.mp4")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// expand all boxes
	_, err = mp4.ReadBoxStructure(file, func(h *mp4.ReadHandle) (interface{}, error) {
		fmt.Println("depth", len(h.Path))

		// Box Type (e.g. "mdhd", "tfdt", "mdat")
		fmt.Println("type", h.BoxInfo.Type.String())

		// Box Size
		fmt.Println("size", h.BoxInfo.Size)

		if h.BoxInfo.IsSupportedType() {
			// Payload
			box, _, err := h.ReadPayload()
			if err != nil {
				return nil, err
			}
			str, err := mp4.Stringify(box, h.BoxInfo.Context)
			if err != nil {
				return nil, err
			}
			fmt.Println("payload", str)

			// Expands children
			return h.Expand()
		}
		return nil, nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}
