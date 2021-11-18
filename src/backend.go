package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"log"
	"net"
	"os"
	"time"
)

func Scan(brotherIP string, brotherPort int, resolution int, color string, adf bool) ([][]byte, int, int) {
	log.Println("Valid IP address, opening socket...")

	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", brotherIP, brotherPort))

	HandleError(err)

	defer socket.Close()

	width, heigth := sendRequest(socket, resolution, color, adf)

	bytes, err := getScanBytes(socket)

	HandleError(err)

	return removeHeaders(bytes), width, heigth
}

func sendRequest(socket net.Conn, resolution int, _mode string, adf bool) (int, int) {

	mode, compression := getCompressionMode(_mode)

	log.Println("Reading scanner status...")

	status := readPacket(socket)[:7]

	if status != scanner.ready {
		HandleError(fmt.Errorf("invalid reply from scanner: %s", status))
	}

	log.Println("Leasing options...")

	request := []byte(fmt.Sprintf(formats.leaseRequest, resolution, resolution, mode))
	sendPacket(socket, request)

	offer := readPacket(socket)

	if !adf {
		log.Println("Disabling automatic document feeder (ADF)")

		request = []byte(formats.disableADF)
		sendPacket(socket, request)

		readPacket(socket)
	}

	log.Println("Sending scan request...")

	width, height := 0, 0
	planeWidth, planeHeight := 0, 0
	dpiX, dpiY := 0, 0
	adfStatus := 0

	fmt.Sscanf(offer[3:], "%d,%d,%d,%d,%d,%d,%d", &dpiX, &dpiY, &adfStatus, &planeWidth, &width, &planeHeight, &height)

	if planeHeight == 0 {
		planeHeight = scanner.A4height
	}

	width = mmToPixels(planeWidth, dpiX)
	height = mmToPixels(planeHeight, dpiY)

	request = []byte(fmt.Sprintf(formats.scanRequest, dpiX, dpiY, mode, compression, width, height))

	sendPacket(socket, request)

	log.Println("Scanning started...")

	return width, height
}

func getScanBytes(socket net.Conn) ([]byte, error) {
	log.Println("Getting packets...")

	packet := make([]byte, 2048)
	scanBytes := make([]byte, 0)

readPackets:
	for {
		socket.SetDeadline(time.Now().Add(time.Second * 10))
		bytes, err := socket.Read(packet)

		switch err := err.(type) {
		case net.Error:
			if err.Timeout() {
				break readPackets
			}

		case nil:
			scanBytes = append(scanBytes, packet[:bytes]...)

		default:
			HandleError(err)
		}
	}

	if (len(scanBytes)) < 1 {
		return scanBytes, fmt.Errorf("no data received")
	}

	return scanBytes, nil
}

func SaveImage(data []byte, width int, height int, name string, color string) {

	log.Println("Saving image...")

	_, compression := getCompressionMode(color)

	if compression != scanner.compression.jpeg {

		img := image.NewGray(image.Rectangle{
			image.Point{0, 0},
			image.Point{width, height},
		})

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				img.SetGray(x, y, colorToGray(data[(y*width+x)%len(data)]))
			}
		}

		file, err := os.Create(name)
		HandleError(err)

		png.Encode(file, img)

	} else {

		err := os.WriteFile(name, data, 0644)
		HandleError(err)
	}
}

func removeHeaders(data []byte) [][]byte {
	log.Println("Removing headers from bytes...")

	pages := make([][]byte, 0)
	page := make([]byte, 0)

	currentPage := 1
	i := 0

headersLoop:
	for {
		if data[i] == scanner.endPage {
			pages = append(pages, page)

			if len(data) > i+10 && data[i+10] == scanner.endScan {
				break headersLoop
			}

			page = make([]byte, 0)

			currentPage++

			i += scanner.headerLen - 2
			continue headersLoop
		}

		payloadLen := binary.LittleEndian.Uint16(data[i+scanner.headerLen-2 : i+scanner.headerLen])
		chunkSize := int(payloadLen) + scanner.headerLen

		page = append(page, data[i+scanner.headerLen:i+chunkSize]...)

		i += chunkSize
	}

	return pages
}
