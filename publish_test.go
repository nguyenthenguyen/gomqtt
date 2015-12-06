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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublishInterface(t *testing.T) {
	msg := NewPublishPacket()

	require.Equal(t, msg.Type(), PUBLISH)
	require.NotNil(t, msg.String())
}

func TestPublishPacketDecode1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 11,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishPacket()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, 7, int(msg.PacketID))
	require.Equal(t, []byte("surgemq"), msg.Topic)
	require.Equal(t, []byte("send me home"), msg.Payload)
	require.Equal(t, 1, int(msg.QOS))
	require.Equal(t, true, msg.Retain)
	require.Equal(t, true, msg.Dup)
}

func TestPublishPacketDecode2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishPacket()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, 0, int(msg.PacketID))
	require.Equal(t, []byte("surgemq"), msg.Topic)
	require.Equal(t, []byte("send me home"), msg.Payload)
	require.Equal(t, 0, int(msg.QOS))
	require.Equal(t, false, msg.Retain)
	require.Equal(t, false, msg.Dup)
}

func TestPublishPacketDecodeError1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		2, // <- too much
	}

	msg := NewPublishPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 6, // <- wrong qos
		0,
	}

	msg := NewPublishPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError3(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		0,
		// <- missing topic stuff
	}

	msg := NewPublishPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError4(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		2,
		0, // topic name MSB
		1, // topic name LSB
		// <- missing topic string
	}

	msg := NewPublishPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestPublishPacketDecodeError5(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		2,
		0, // topic name MSB
		1, // topic name LSB
		't',
		// <- missing packet id
	}

	msg := NewPublishPacket()
	_, err := msg.Decode(msgBytes)

	require.Error(t, err)
}

func TestPublishPacketEncode1(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 11,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishPacket()
	msg.Topic = []byte("surgemq")
	msg.QOS = QOSAtLeastOnce
	msg.Retain = true
	msg.Dup = true
	msg.PacketID = 7
	msg.Payload = []byte("send me home")

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestPublishPacketEncode2(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH << 4),
		21,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishPacket()
	msg.Topic = []byte("surgemq")
	msg.Payload = []byte("send me home")

	dst := make([]byte, msg.Len())
	n, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)
	require.Equal(t, msgBytes, dst[:n])
}

func TestPublishPacketEncodeError1(t *testing.T) {
	msg := NewPublishPacket()
	msg.Topic = []byte("") // <- empty topic

	dst := make([]byte, msg.Len())
	_, err := msg.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError2(t *testing.T) {
	msg := NewPublishPacket()
	msg.Topic = []byte("t")
	msg.QOS = 3 // <- wrong qos

	dst := make([]byte, msg.Len())
	_, err := msg.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError3(t *testing.T) {
	msg := NewPublishPacket()
	msg.Topic = []byte("t")

	dst := make([]byte, 1) // <- too small
	_, err := msg.Encode(dst)

	require.Error(t, err)
}

func TestPublishPacketEncodeError4(t *testing.T) {
	msg := NewPublishPacket()
	msg.Topic = make([]byte, 65536) // <- too big

	dst := make([]byte, msg.Len())
	_, err := msg.Encode(dst)

	require.Error(t, err)
}

func TestPublishEqualDecodeEncode(t *testing.T) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		23,
		0, // topic name MSB
		7, // topic name LSB
		's', 'u', 'r', 'g', 'e', 'm', 'q',
		0, // packet ID MSB
		7, // packet ID LSB
		's', 'e', 'n', 'd', ' ', 'm', 'e', ' ', 'h', 'o', 'm', 'e',
	}

	msg := NewPublishPacket()
	n, err := msg.Decode(msgBytes)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n)

	dst := make([]byte, msg.Len())
	n2, err := msg.Encode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n2)
	require.Equal(t, msgBytes, dst[:n2])

	n3, err := msg.Decode(dst)

	require.NoError(t, err)
	require.Equal(t, len(msgBytes), n3)
}

func BenchmarkPublishEncode(b *testing.B) {
	msg := NewPublishPacket()
	msg.Topic = []byte("t")
	msg.QOS = QOSAtLeastOnce
	msg.PacketID = 1
	msg.Payload = []byte("p")

	buf := make([]byte, msg.Len())

	for i := 0; i < b.N; i++ {
		_, err := msg.Encode(buf)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkPublishDecode(b *testing.B) {
	msgBytes := []byte{
		byte(PUBLISH<<4) | 2,
		6,
		0, // topic name MSB
		1, // topic name LSB
		't',
		0, // packet ID MSB
		1, // packet ID LSB
		'p',
	}

	msg := NewPublishPacket()

	for i := 0; i < b.N; i++ {
		_, err := msg.Decode(msgBytes)
		if err != nil {
			panic(err)
		}
	}
}
