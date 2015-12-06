// Copyright (c) 2014 The gomqtt Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package packet

import (
	"encoding/binary"
	"fmt"
)

// A PUBLISH Control Packet is sent from a Client to a Server or from Server to a Client
// to transport an Application Message.
type PublishPacket struct {
	// The Topic of the message.
	Topic []byte

	// The Payload of the message.
	Payload []byte

	// The QOS indicates the level of assurance for delivery of a message.
	QOS byte

	// If the RETAIN flag is set to true, in a PUBLISH Packet sent by a Client to a
	// Server, the Server MUST store the Application Message and its QOS, so that it can be
	// delivered to future subscribers whose subscriptions match its topic name.
	Retain bool

	// If the DUP flag is set to false, it indicates that this is the first occasion that the
	// Client or Server has attempted to send this MQTT PUBLISH Packet. If the DUP flag is
	// set to true, it indicates that this might be re-delivery of an earlier attempt to send
	// the Packet.
	Dup bool

	// Shared packet identifier.
	PacketID uint16
}

var _ Packet = (*PublishPacket)(nil)

// NewPublishPacket creates a new PUBLISH packet.
func NewPublishPacket() *PublishPacket {
	return &PublishPacket{}
}

// Type returns the packets type.
func (pm PublishPacket) Type() Type {
	return PUBLISH
}

// String returns a string representation of the packet.
func (pm PublishPacket) String() string {
	return fmt.Sprintf("PUBLISH: Topic=%q PacketID=%d QOS=%d Retained=%t Dup=%t Payload=%v",
		pm.Topic, pm.PacketID, pm.QOS, pm.Retain, pm.Dup, pm.Payload)
}

// Len returns the byte length of the encoded packet.
func (pm *PublishPacket) Len() int {
	ml := pm.len()
	return headerLen(ml) + ml
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (pm *PublishPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, flags, rl, err := headerDecode(src[total:], PUBLISH)
	total += hl
	if err != nil {
		return total, err
	}

	// read flags
	pm.Dup = ((flags >> 3) & 0x1) == 1
	pm.Retain = (flags & 0x1) == 1
	pm.QOS = (flags >> 1) & 0x3

	// check qos
	if !validQOS(pm.QOS) {
		return total, fmt.Errorf("PUBLISH/Decode: Invalid QOS (%d)", pm.QOS)
	}

	// check buffer length
	if len(src) < total+2 {
		return total, fmt.Errorf("PUBLISH/Decode: Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
	}

	n := 0

	// read topic
	pm.Topic, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	if pm.QOS != 0 {
		// check buffer length
		if len(src) < total+2 {
			return total, fmt.Errorf("PUBLISH/Decode: Insufficient buffer size. Expecting %d, got %d", total+2, len(src))
		}

		// read packet id
		pm.PacketID = binary.BigEndian.Uint16(src[total:])
		total += 2
	}

	// calculate payload length
	l := int(rl) - (total - hl)

	// read payload
	if l > 0 {
		pm.Payload = src[total : total+l]
		total += len(pm.Payload)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (pm *PublishPacket) Encode(dst []byte) (int, error) {
	total := 0

	// check topic length
	if len(pm.Topic) == 0 {
		return total, fmt.Errorf("PUBLISH/Encode: Topic name is empty")
	}

	flags := byte(0)

	// set dup flag
	if pm.Dup {
		flags |= 0x8 // 00001000
	} else {
		flags &= 247 // 11110111
	}

	// set retain flag
	if pm.Retain {
		flags |= 0x1 // 00000001
	} else {
		flags &= 254 // 11111110
	}

	// check qos
	if !validQOS(pm.QOS) {
		return 0, fmt.Errorf("PUBLISH/Encode: Invalid QOS %d", pm.QOS)
	}

	// set qos
	flags = (flags & 249) | (pm.QOS << 1) // 249 = 11111001

	// encode header
	n, err := headerEncode(dst[total:], flags, pm.len(), pm.Len(), PUBLISH)
	total += n
	if err != nil {
		return total, err
	}

	// write topic
	n, err = writeLPBytes(dst[total:], pm.Topic)
	total += n
	if err != nil {
		return total, err
	}

	// write packet id
	if pm.QOS != 0 {
		binary.BigEndian.PutUint16(dst[total:], pm.PacketID)
		total += 2
	}

	// write payload
	copy(dst[total:], pm.Payload)
	total += len(pm.Payload)

	return total, nil
}

// Returns the payload length.
func (pm *PublishPacket) len() int {
	total := 2 + len(pm.Topic) + len(pm.Payload)
	if pm.QOS != 0 {
		total += 2
	}

	return total
}
