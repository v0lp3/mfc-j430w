package main

import (
	"flag"
	"fmt"
	"net"
)

func main() {
	const brotherPort int = 54921

	brotherIP := flag.String("a", "192.168.0.157", "IP address of the Brother scanner")
	resolution := flag.Int("r", 300, "Resolution of the scan")
	color := flag.String("c", "CGRAY", "Color mode of the scan")
	adf := flag.Bool("m", false, "Enable scan of all pages from feeder")
	name := flag.String("n", "scan.jpg", "Name of the output file")

	flag.Parse()

	if net.ParseIP(*brotherIP) == nil {
		HandleError(fmt.Errorf("invalid IP address: %s", *brotherIP))
	}

	rawImage, width, heigth := Scan(*brotherIP, brotherPort, *resolution, *color, *adf)

	SaveImage(rawImage, width, heigth, *name, *color)
}
