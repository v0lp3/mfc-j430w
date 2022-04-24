package main

import (
	"image/color"
	"log"
	"net"
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
	}
}

func getCompressionMode(_mode string) (string, string) {

	if _mode == scanner.mode.grayscale {
		return _mode, scanner.compression.none
	} else {
		return scanner.mode.color, scanner.compression.jpeg
	}
}

func mmToPixels(mm int, dpi int) int {
	return int(float32(mm*dpi) / scanner.mmInch)
}

func colorToGray(value byte) color.Gray {
	return color.Gray{Y: value}
}
