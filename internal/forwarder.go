package internal

import (
	"net"
)

// ForwardMessage forwards the OTLP-transformed message to the destination.
// Note: The message is already transformed to OTLP format by TransformToOTLP.
func ForwardMessage(data []byte, destConn *net.TCPConn) error {
	// Send to destination
	_, err := destConn.Write(data)
	return err
}
