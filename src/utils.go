package main

import (
	"image/color"
	"log"
	"net"
	"os"
)

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

func colorToGray(byte byte) color.Gray {
	return color.Gray{Y: byte}
}
