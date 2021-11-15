package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	const BrotherPort int = 54921
	BrotherIP := os.Args[1]
	resolution := os.Args[2]
	color := os.Args[3]

	if net.ParseIP(BrotherIP) == nil {
		fmt.Fprintf(os.Stderr, "Invalid IP address. Usage: %s <scanner ip>", os.Args[0])
		os.Exit(0)
	}

	log.Println("Valid IP address, opening socket...")

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", BrotherIP, BrotherPort))
	handleError(err)

	sendRequest(conn, resolution, color)

	defer conn.Close()
}

func sendRequest(conn net.Conn, _resolution string, _mode string) {

	reply := make([]byte, 128)

	resolution := checkResolution(_resolution)
	mode, compression := checkMode(_mode)

	log.Println("Reading scanner status...")

	_, err := conn.Read(reply)
	handleError(err)

	if "+OK 200" != string(reply[:7]) {
		handleError(fmt.Errorf("Invalid reply from scanner: %s", reply))
	}

	log.Println("Leasing options...")

	resp := []byte(fmt.Sprintf("\x1bI\nR=%d,%d\nM=%s\n\x80", resolution, resolution, mode))
	_, err = conn.Write(resp)
	handleError(err)

	_, err = conn.Read(reply)
	handleError(err)

	response := strings.Split(string(reply), ",")

	log.Println("Sending scan request...")

	scanRequestFormat := "\x1bX\nR=%v,%v\nM=%s\nC=%s\nJ=MID\nB=50\nN=50\nA=0,0,%v,%v\n\x80"

	resp = []byte(fmt.Sprintf(scanRequestFormat, response[1], response[1], mode, compression, response[5], response[6]))
	_, err = conn.Write(resp)
	handleError(err)
}

func checkResolution(_resolution string) int {
	resolution, err := strconv.Atoi(_resolution)

	if err != nil && resolution < 100 {
		resolution = 300
	}

	return resolution
}

func checkMode(_mode string) (string, string) {

	if _mode == "GRAY64" {
		return _mode, "NONE"
	} else {
		return "CGRAY", "JPEG"
	}
}

func handleError(err error) {

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
