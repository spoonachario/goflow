package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

type netflow struct {
	Templates map[uint16]netflowPacketTemplate
}

type netflowPacket struct {
	Header    netflowPacketHeader
	Templates map[uint16]netflowPacketTemplate
}

type netflowPacketHeader struct {
	Version  uint16
	Count    int16
	Uptime   int32
	Sequence int32
	Id       int32
}

type netflowPacketFlowsetId struct {
	FlowSetID uint16
}

type netflowPacketTemplate struct {
	FlowSetID  uint16
	Length     uint16
	ID         uint16
	FieldCount uint16
	Fields     map[uint16]templateField
}

type templateField struct {
	FieldType uint16
	Length    uint16
}

func main() {

	// Provides parsing for Netflow V9 Records
	// https://www.ietf.org/rfc/rfc3954.txt

	addr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("127.0.0.1"),
	}
	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return
	}

	nfpacket := netflowPacket{
		Templates: make(map[uint16]netflowPacketTemplate),
	}

	p := netflowPacketHeader{}
	// Buffer creates an array of bytes
	// We want to read the entire datagram in as UDP is type SOCK_DGRAM and "Read" can't be called more than once
	packet := make([]byte, 1500)
	// Read the max number of bytes in a datagram(1500) into a variable length slice of bytes, 'Buffer'
	conn.Read(packet)
	id := netflowPacketFlowsetId{}
	p.Version = binary.BigEndian.Uint16(packet[:2])
	id.FlowSetID = binary.BigEndian.Uint16(packet[20:22])

	nfpacket.Header = p
	switch id.FlowSetID {
	// Template flowset
	case 0:
		template := netflowPacketTemplate{
			Fields: make(map[uint16]templateField),
		}
		template.Length = binary.BigEndian.Uint16(packet[22:24])
		templateSlice := packet[22:template.Length]
		template.ID = binary.BigEndian.Uint16(templateSlice[2:4])

		// Get the number of Fields
		template.FieldCount = binary.BigEndian.Uint16(templateSlice[4:6])
		fmt.Printf("version: %v Fields %v\n", p.Version, template.FieldCount)

		// Start at the first fields and work through
		fieldStart := 6
		var read = uint16(0)
		for read < template.FieldCount {
			fieldTypeEnd := fieldStart + 2
			fieldType := binary.BigEndian.Uint16(templateSlice[fieldStart:fieldTypeEnd])
			fieldLengthEnd := fieldTypeEnd + 2
			fieldLength := binary.BigEndian.Uint16(templateSlice[fieldTypeEnd:fieldLengthEnd])

			// Create template FIELD struct
			field := templateField{
				FieldType: fieldType,
				Length:    fieldLength,
			}
			// Assign to fields, keyed by the type, to this template record
			template.Fields[fieldType] = field

			read++
			fieldStart = fieldLengthEnd
		}

		for t, template := range template.Fields {
			fmt.Printf("Type read: %v, Length: %v\n", t, template.Length)
		}

	}
}
