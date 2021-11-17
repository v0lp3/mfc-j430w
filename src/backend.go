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

	if status != "+OK 200" {
		HandleError(fmt.Errorf("invalid reply from scanner: %s", status))
	}

	log.Println("Leasing options...")

	request := []byte(fmt.Sprintf("\x1bI\nR=%d,%d\nM=%s\n\x80", resolution, resolution, mode))
	sendPacket(socket, request)

	offer := readPacket(socket)

	if !adf {
		log.Println("Disabling automatic document feeder (ADF)")

		request = []byte("\x1bD\nADF\n\x80")
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
		planeHeight = 294
	}

	width = mmToPixels(planeWidth, dpiX)
	height = mmToPixels(planeHeight, dpiY)

	requestFormat := "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%d,%d\n\x80"
	request = []byte(fmt.Sprintf(requestFormat, dpiX, dpiY, mode, compression, width, height))

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

	if compression != "JPEG" {

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

	const headerLen int = 12

	pages := make([][]byte, 0)
	page := make([]byte, 0)

	currentPage := 1
	i := 0

headersLoop:
	for {
		if data[i] == 0x82 {
			log.Println("Parsed page", currentPage)
			pages = append(pages, page)

			if len(data) > i+10 && data[i+10] == 0x80 {
				break headersLoop
			}

			page = make([]byte, 0)

			currentPage++

			i += headerLen - 2
			continue headersLoop
		}

		payloadLen := binary.LittleEndian.Uint16(data[i+headerLen-2 : i+headerLen])
		chunkSize := int(payloadLen) + headerLen

		page = append(page, data[i+headerLen:i+chunkSize]...)

		i += chunkSize
	}

	return pages
}
