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

func Scan(brotherIP string, brotherPort int, resolution int, color string, adf bool, name string) {
	log.Println("Valid IP address, opening socket...")

	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", brotherIP, brotherPort))

	HandleError(err)

	defer socket.Close()

	sendRequest(socket, resolution, color, adf)

	processResponse(socket, name)

	HandleError(err)
}

func sendRequest(socket net.Conn, resolution int, _mode string, adf bool) {

	preferences.mode, preferences.encode = getCompressionMode(_mode)

	log.Println("Reading scanner status...")

	status := readPacket(socket)[:7]

	if status != scanner.ready {
		HandleError(fmt.Errorf("invalid reply from scanner: %s", status))
	}

	log.Println("Leasing options...")

	request := []byte(fmt.Sprintf(formats.leaseRequest, resolution, resolution, preferences.mode))
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

	fmt.Sscanf(offer[3:], "%d,%d,%d,%d,%d,%d,%d", &dpiX, &dpiY, &adfStatus, &planeWidth,
		&width, &planeHeight, &height)

	if planeHeight == 0 {
		planeHeight = scanner.A4height
	}

	preferences.width = mmToPixels(planeWidth, dpiX)
	preferences.height = mmToPixels(planeHeight, dpiY)

	request = []byte(fmt.Sprintf(formats.scanRequest, dpiX, dpiY, preferences.mode,
		preferences.encode, preferences.width, preferences.height))

	sendPacket(socket, request)

	log.Println("Scanning started...")
}

func processResponse(socket net.Conn, name string) {
	log.Println("Getting packets...")

	packet := make([]byte, 2048)
	scanBytes := make([]byte, 0)

	headerPos := 0
	currentPage := 1

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

			if len(scanBytes) > headerPos {
				if scanBytes[headerPos] == scanner.endPage {

					if len(scanBytes) > (headerPos+10) && scanBytes[headerPos+10] == scanner.endScan {
						processPacket(scanBytes, name, currentPage)
						break readPackets
					}

					go processPacket(scanBytes, name, currentPage)
					scanBytes = make([]byte, 0)

					headerPos = 0
					currentPage++

					continue readPackets
				}

				payloadLen := binary.LittleEndian.Uint16(scanBytes[headerPos+scanner.headerLen-2 : headerPos+scanner.headerLen])
				headerPos += int(payloadLen) + scanner.headerLen
			}

		default:
			HandleError(err)
		}
	}

	if (len(scanBytes)) < 1 {
		HandleError(fmt.Errorf("no data received"))
	}
}

func SaveImage(data []byte, name string) {

	log.Printf("Saving image %s...", name)

	_, compression := getCompressionMode(preferences.mode)
	width := preferences.width
	height := preferences.height

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

func processPacket(data []byte, name string, currentPage int) {
	log.Printf("Removing headers from page %d...\n", currentPage)

	i := 0
	page := make([]byte, 0)

headersLoop:
	for {
		if data[i] == scanner.endPage {
			break headersLoop
		}

		payloadLen := binary.LittleEndian.Uint16(data[i+scanner.headerLen-2 : i+scanner.headerLen])
		chunkSize := int(payloadLen) + scanner.headerLen

		page = append(page, data[i+scanner.headerLen:i+chunkSize]...)

		i += chunkSize
	}

	SaveImage(page, fmt.Sprintf("%s(%d)", name, currentPage))
}
