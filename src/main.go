package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	const brotherPort int = 54921

	brotherIP := flag.String("a", "192.168.0.157", "IP address of the Brother scanner")
	resolution := flag.Int("r", 300, "Resolution of the scan")
	color := flag.String("c", "CGRAY", "Color mode of the scan")
	adf := flag.Bool("m", false, "Enable scan of all pages from feeder")

	flag.Parse()

	if net.ParseIP(*brotherIP) == nil {
		HandleError(fmt.Errorf("Invalid IP address: %s", *brotherIP))
	}

	log.Println("Valid IP address, opening socket...")

	socket, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *brotherIP, brotherPort))

	HandleError(err)

	defer socket.Close()

	SendRequest(socket, *resolution, *color, *adf)
}

func SendRequest(socket net.Conn, resolution int, _mode string, adf bool) {

	mode, compression := GetCompressionMode(_mode)

	log.Println("Reading scanner status...")

	status := ReadPacket(socket)[:7]

	if "+OK 200" != status {
		HandleError(fmt.Errorf("Invalid reply from scanner: %s", status))
	}

	log.Println("Leasing options...")

	request := []byte(fmt.Sprintf("\x1bI\nR=%d,%d\nM=%s\n\x80", resolution, resolution, mode))
	SendPacket(socket, request)

	offer := ReadPacket(socket)

	if adf {
		log.Println("Enabling automatic document feeder (ADF)")

		request = []byte("\x1bD\nADF\n\x80")
		SendPacket(socket, request)

		ReadPacket(socket)
	}

	log.Println("Sending scan request...")

	offerOptions := strings.Split(offer, ",")

	requestFormat := "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%v,%v\n\x80"
	request = []byte(fmt.Sprintf(requestFormat, offerOptions[1], offerOptions[1], mode, compression, offerOptions[5], offerOptions[6]))
	SendPacket(socket, request)

	log.Println("Scanning started...")
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

func SendPacket(socket net.Conn, packet []byte) {
	_, err := socket.Write(packet)
	HandleError(err)
}

func ReadPacket(socket net.Conn) string {
	reply := make([]byte, 128)

	_, err := socket.Read(reply)
	HandleError(err)

	return string(reply)
}
