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

import "fmt"

// ConnackCode is the type representing the return code in the CONNACK packet.
type ConnackCode byte

// All existing ConnackCodes.
const (
	ConnectionAccepted ConnackCode = iota
	ErrInvalidProtocolVersion
	ErrIdentifierRejected
	ErrServerUnavailable
	ErrBadUsernameOrPassword
	ErrNotAuthorized
)

// Valid checks if the ConnackCode is valid.
func (cc ConnackCode) Valid() bool {
	return cc <= 5
}

// Error returns the corresponding error string for the ConnackCode.
func (cc ConnackCode) Error() string {
	switch cc {
	case ConnectionAccepted:
		return "Connection accepted"
	case ErrInvalidProtocolVersion:
		return "Connection Refused, unacceptable protocol version"
	case ErrIdentifierRejected:
		return "Connection Refused, identifier rejected"
	case ErrServerUnavailable:
		return "Connection Refused, Server unavailable"
	case ErrBadUsernameOrPassword:
		return "Connection Refused, bad user name or password"
	case ErrNotAuthorized:
		return "Connection Refused, not authorized"
	}

	return "Unknown error"
}

// The CONNACK Packet is the packet sent by the Server in response to a CONNECT Packet
// received from a Client.
type ConnackPacket struct {
	// The Session Present flag enables a Client to establish whether the Client and
	// Server have a consistent view about whether there is already stored Session state.
	SessionPresent bool

	// If a well formed CONNECT Packet is received by the Server, but the Server is unable
	// to process it for some reason, then the Server SHOULD attempt to send a CONNACK packet
	// containing the appropriate non-zero Connect return code.
	ReturnCode ConnackCode
}

var _ Packet = (*ConnackPacket)(nil)

// NewConnackPacket creates a new CONNACK packet.
func NewConnackPacket() *ConnackPacket {
	return &ConnackPacket{}
}

// Type returns the packets type.
func (cm ConnackPacket) Type() Type {
	return CONNACK
}

// String returns a string representation of the packet.
func (cm ConnackPacket) String() string {
	return fmt.Sprintf("CONNACK: SessionPresent=%t ReturnCode=%q", cm.SessionPresent, cm.ReturnCode)
}

// Len returns the byte length of the encoded packet.
func (cm *ConnackPacket) Len() int {
	return headerLen(2) + 2
}

// Decode reads from the byte slice argument. It returns the total number of bytes
// decoded, and whether there have been any errors during the process.
// The byte slice MUST NOT be modified during the duration of this
// packet being available since the byte slice never gets copied.
func (cm *ConnackPacket) Decode(src []byte) (int, error) {
	total := 0

	// decode header
	hl, _, rl, err := headerDecode(src, CONNACK)
	total += hl
	if err != nil {
		return total, err
	}

	// check remaining length
	if rl != 2 {
		return total, fmt.Errorf("CONNACK/Decode: Expected remaining length to be 2")
	}

	// read connack flags
	connackFlags := src[total]
	cm.SessionPresent = connackFlags&0x1 == 1
	total++

	// check flags
	if connackFlags&254 != 0 {
		return 0, fmt.Errorf("CONNACK/Decode: Bits 7-1 in acknowledge flags byte (1) are not 0")
	}

	// read return code
	cm.ReturnCode = ConnackCode(src[total])
	total++

	// check return code
	if !cm.ReturnCode.Valid() {
		return 0, fmt.Errorf("CONNACK/Decode: Invalid return code (%d)", cm.ReturnCode)
	}

	return total, nil
}

// Encode writes the packet bytes into the byte slice from the argument. It
// returns the number of bytes encoded and whether there's any errors along
// the way. If there is an error, the byte slice should be considered invalid.
func (cm *ConnackPacket) Encode(dst []byte) (int, error) {
	total := 0

	// encode header
	n, err := headerEncode(dst[total:], 0, 2, cm.Len(), CONNACK)
	total += n
	if err != nil {
		return total, err
	}

	// set session present flag
	if cm.SessionPresent {
		dst[total] = 1 // 00000001
	} else {
		dst[total] = 0 // 00000000
	}
	total++

	// check return code
	if !cm.ReturnCode.Valid() {
		return total, fmt.Errorf("CONNACK/Encode: Invalid return code (%d)", cm.ReturnCode)
	}

	// set return code
	dst[total] = byte(cm.ReturnCode)
	total++

	return total, nil
}
