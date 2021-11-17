package main

import (
	"image/color"
	"log"
	"net"
	"os"
)

func sendPacket(socket net.Conn, packet []byte) {
	_, err := socket.Write(packet)
	HandleError(err)
}

func readPacket(socket net.Conn) string {
	reply := make([]byte, 64)

	_, err := socket.Read(reply)
	HandleError(err)

	return string(reply)
}

func HandleError(err error) {

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func getCompressionMode(_mode string) (string, string) {

	if _mode == "GRAY64" {
		return _mode, "NONE"
	} else {
		return "CGRAY", "JPEG"
	}
}

func colorToGray(value byte) color.Gray {
	return color.Gray{Y: value}
}
