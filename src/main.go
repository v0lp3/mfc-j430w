package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
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

	log.Println("Valid IP address, opening socket...")

	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *brotherIP, brotherPort))

	HandleError(err)

	defer socket.Close()

	width, heigth := SendRequest(socket, *resolution, *color, *adf)

	bytes := GetScanBytes(socket)

	rawImage := RemoveHeaders(bytes)

}

func SendRequest(socket net.Conn, resolution int, _mode string, adf bool) (int, int) {

	mode, compression := GetCompressionMode(_mode)

	log.Println("Reading scanner status...")

	status := ReadPacket(socket)[:7]

	if status != "+OK 200" {
		HandleError(fmt.Errorf("invalid reply from scanner: %s", status))
	}

	log.Println("Leasing options...")

	request := []byte(fmt.Sprintf("\x1bI\nR=%d,%d\nM=%s\n\x80", resolution, resolution, mode))
	SendPacket(socket, request)

	offer := ReadPacket(socket)

	if !adf {
		log.Println("Disabling automatic document feeder (ADF)")

		request = []byte("\x1bD\nADF\n\x80")
		SendPacket(socket, request)

		ReadPacket(socket)
	}

	log.Println("Sending scan request...")

	offerOptions := strings.Split(offer, ",")

	width, height := 0, 0
	fmt.Sscanf(offerOptions[4], "%d", &width)
	fmt.Sscanf(offerOptions[6], "%d", &height)

	requestFormat := "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%v,%v\n\x80"

	request = []byte(fmt.Sprintf(requestFormat, offerOptions[1], offerOptions[1], mode, compression, width, height))
	SendPacket(socket, request)

	log.Println("Scanning started...")

	return width, height
}

func GetScanBytes(socket net.Conn) []byte {
	log.Println("Getting packets...")

	packet := make([]byte, 2048)
	scanBytes := make([]byte, 0)

readPackets:
	for {
		socket.SetDeadline(time.Now().Add(time.Second * 5))
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

	return scanBytes
}

func HandleError(err error) {

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func GetCompressionMode(_mode string) (string, string) {

	if _mode == "GRAY64" {
		return _mode, "NONE"
	} else {
		return "CGRAY", "JPEG"
	}
}

func RemoveHeaders(scan []byte) []byte {
	log.Println("Removing headers from bytes...")

	const headerLen int = 12

	payloadLen := binary.LittleEndian.Uint16(scan[headerLen-2 : headerLen])
	chunkSize := int(payloadLen) + headerLen
	scanOutput := make([]byte, 0)

	for i := 0; i < len(scan)-chunkSize; i += chunkSize {
		scanOutput = append(scanOutput, scan[i+headerLen:i+chunkSize]...)
	}

	return scanOutput
}

func SendPacket(socket net.Conn, packet []byte) {
	_, err := socket.Write(packet)
	HandleError(err)
}

func ReadPacket(socket net.Conn) string {
	reply := make([]byte, 64)

	_, err := socket.Read(reply)
	HandleError(err)

	return string(reply)
}
