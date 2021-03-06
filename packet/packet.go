// Package packet implements functionality for encoding and decoding MQTT packets.
package packet

import "encoding/binary"

const (
	// QOSAtMostOnce defines that the message is delivered at most once, or it
	// may not be delivered at all.
	QOSAtMostOnce byte = iota

	// QOSAtLeastOnce defines that the message is always delivered at least once.
	QOSAtLeastOnce

	// QOSExactlyOnce defines that the message is always delivered exactly once.
	QOSExactlyOnce

	// QOSFailure indicates that there has been an error while subscribing
	// to a specific topic.
	QOSFailure = 0x80
)

// ID is the type used to store packet ids.
type ID uint16

// A GenericPacket is an MQTT control packet that can be encoded to a buffer or decoded
// from a buffer.
type GenericPacket interface {
	// Type returns the packets type.
	Type() Type

	// Len returns the byte length of the encoded packet.
	Len() int

	// Decode reads from the byte slice argument. It returns the total number of
	// bytes decoded, and whether there have been any errors during the process.
	Decode(src []byte) (int, error)

	// Encode writes the packet bytes into the byte slice from the argument. It
	// returns the number of bytes encoded and whether there's any errors along
	// the way. If there is an error, the byte slice should be considered invalid.
	Encode(dst []byte) (int, error)

	// String returns a string representation of the packet.
	String() string
}

// DetectPacket tries to detect the next packet in a buffer. It returns a length
// greater than zero if the packet has been detected as well as its Type.
func DetectPacket(src []byte) (int, Type) {
	// check for minimum size
	if len(src) < 2 {
		return 0, 0
	}

	// get type
	t := Type(src[0] >> 4)

	// get remaining length
	rl, n := binary.Uvarint(src[1:])
	if n <= 0 {
		return 0, 0
	}

	return 1 + n + int(rl), t
}

// GetID checks the packets type and returns its ID and true, or if it
// does not have a ID, zero and false.
func GetID(packet GenericPacket) (ID, bool) {
	switch packet.Type() {
	case PUBLISH:
		return packet.(*PublishPacket).ID, true
	case PUBACK:
		return packet.(*PubackPacket).ID, true
	case PUBREC:
		return packet.(*PubrecPacket).ID, true
	case PUBREL:
		return packet.(*PubrelPacket).ID, true
	case PUBCOMP:
		return packet.(*PubcompPacket).ID, true
	case SUBSCRIBE:
		return packet.(*SubscribePacket).ID, true
	case SUBACK:
		return packet.(*SubackPacket).ID, true
	case UNSUBSCRIBE:
		return packet.(*UnsubscribePacket).ID, true
	case UNSUBACK:
		return packet.(*UnsubackPacket).ID, true
	}

	return 0, false
}

// Fuzz is a basic fuzzing test that works with https://github.com/dvyukov/go-fuzz:
//
//		$ go-fuzz-build github.com/gomqtt/packet
//		$ go-fuzz -bin=./packet-fuzz.zip -workdir=./fuzz
func Fuzz(data []byte) int {
	// check for zero length data
	if len(data) == 0 {
		return 1
	}

	// detect packet
	_, mt := DetectPacket(data)

	// for testing purposes we will not cancel
	// on incomplete buffers

	// create a new packet
	pkt, err := mt.New()
	if err != nil {
		return 0
	}

	// decode it from the buffer.
	_, err = pkt.Decode(data)
	if err != nil {
		return 0
	}

	// everything was ok
	return 1
}
