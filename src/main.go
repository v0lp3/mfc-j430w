package main

import (
	"flag"
	"fmt"
	"net"
)

func main() {
	const brotherPort int = 54921

	brotherIP := flag.String("a", "", "IP address of the Brother scanner")
	resolution := flag.Int("r", 300, "Resolution of the scan")
	color := flag.String("c", "CGRAY", "Color mode of the scan (CGRAY, GRAY64)")
	name := flag.String("n", "scan.jpg", "Name of the output file")

	flag.Parse()

	if net.ParseIP(*brotherIP) == nil {
		HandleError(fmt.Errorf("invalid IP address: %s", *brotherIP))
	}

	rawImages, width, heigth := Scan(*brotherIP, brotherPort, *resolution, *color)

	for i, rawImage := range rawImages {
		if i == len(rawImages)-1 {
			SaveImage(rawImage, width, heigth, fmt.Sprintf("%s(%d)", *name, i), *color)

		} else {
			go SaveImage(rawImage, width, heigth, fmt.Sprintf("%s(%d)", *name, i), *color)
		}
	}
}
