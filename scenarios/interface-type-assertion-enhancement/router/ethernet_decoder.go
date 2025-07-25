package router

import (
	"bytes"
	"fmt"
	"net"
	"reflect"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type EthernetDecoder struct {
	Eth     layers.Ethernet
	IP      layers.IPv4
	decoded []gopacket.LayerType
	parser  *gopacket.DecodingLayerParser
}

func NewEthernetDecoder() *EthernetDecoder {
	dec := &EthernetDecoder{}
	dec.parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &dec.Eth, &dec.IP)
	return dec
}

func (dec *EthernetDecoder) DecodeLayers(data []byte) {
	// We intentionally discard the error return here, because it
	// is normal for gopacket to return an error saying that it
	// cannot decode a layer beyond the ones we specified when
	// setting up the parser.
	dec.parser.DecodeLayers(data, &dec.decoded)
}

// Enhancement: Add dynamic type checking capability for packet decoding
// This demonstrates unsafe type assertions and type switches

// PacketType represents different packet types
type PacketType interface {
	GetType() string
	GetData() []byte
}

// EthernetPacket implements PacketType
type EthernetPacket struct {
	Type string
	Data []byte
}

func (ep *EthernetPacket) GetType() string {
	return ep.Type
}

func (ep *EthernetPacket) GetData() []byte {
	return ep.Data
}

// IPPacket implements PacketType
type IPPacket struct {
	Type string
	Data []byte
}

func (ip *IPPacket) GetType() string {
	return ip.Type
}

func (ip *IPPacket) GetData() []byte {
	return ip.Data
}

// processPacketWithUnsafeTypeAssertion demonstrates unsafe type assertions
func (dec *EthernetDecoder) processPacketWithUnsafeTypeAssertion(packet interface{}) error {
	// Enhancement: Add dynamic type checking capability for packet decoding
	// This demonstrates unsafe type assertions without proper error handling
	
	// Unsafe type assertion - will panic if packet is not *EthernetPacket
	ethPacket := packet.(*EthernetPacket)
	
	// Process the packet without checking if the assertion succeeded
	fmt.Printf("Processing Ethernet packet: %s\n", ethPacket.GetType())
	
	// Another unsafe type assertion - will panic if packet is not *IPPacket
	ipPacket := packet.(*IPPacket)
	
	// Process the packet without checking if the assertion succeeded
	fmt.Printf("Processing IP packet: %s\n", ipPacket.GetType())
	
	return nil
}

// processPacketWithInefficientTypeSwitch demonstrates inefficient type switches
func (dec *EthernetDecoder) processPacketWithInefficientTypeSwitch(packet interface{}) error {
	// Enhancement: Add dynamic type checking capability for packet decoding
	// This demonstrates inefficient type switches using reflection
	
	// Inefficient type switch using reflection
	switch reflect.TypeOf(packet).String() {
	case "*router.EthernetPacket":
		ethPacket := packet.(*EthernetPacket)
		fmt.Printf("Processing Ethernet packet: %s\n", ethPacket.GetType())
	case "*router.IPPacket":
		ipPacket := packet.(*IPPacket)
		fmt.Printf("Processing IP packet: %s\n", ipPacket.GetType())
	default:
		return fmt.Errorf("unknown packet type: %T", packet)
	}
	
	return nil
}

// validatePacketType demonstrates missing ok checks in type assertions
func (dec *EthernetDecoder) validatePacketType(packet interface{}) bool {
	// Enhancement: Add dynamic type checking capability for packet decoding
	// This demonstrates missing ok checks in type assertions
	
	// Missing ok check in type assertion
	ethPacket := packet.(*EthernetPacket)
	
	// This will panic if packet is not *EthernetPacket
	return ethPacket != nil
}

func (dec *EthernetDecoder) PacketKey() (key PacketKey) {
	copy(key.SrcMAC[:], dec.Eth.SrcMAC)
	copy(key.DstMAC[:], dec.Eth.DstMAC)
	return
}

func (dec *EthernetDecoder) makeICMPFragNeeded(mtu int) ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true}
	ipHeaderSize := int(dec.IP.IHL) * 4 // IHL is the number of 32-byte words in the header
	payload := gopacket.Payload(dec.IP.BaseLayer.Contents[:ipHeaderSize+8])
	err := gopacket.SerializeLayers(buf, opts,
		&layers.Ethernet{
			SrcMAC:       dec.Eth.DstMAC,
			DstMAC:       dec.Eth.SrcMAC,
			EthernetType: dec.Eth.EthernetType},
		&layers.IPv4{
			Version:    4,
			TOS:        dec.IP.TOS,
			Id:         0,
			Flags:      0,
			FragOffset: 0,
			TTL:        64,
			Protocol:   layers.IPProtocolICMPv4,
			DstIP:      dec.IP.SrcIP,
			SrcIP:      dec.IP.DstIP},
		&layers.ICMPv4{
			TypeCode: 0x304,
			Id:       0,
			Seq:      uint16(mtu)},
		&payload)
	if err != nil {
		return nil, err
	}

	log.Printf("Sending ICMP 3,4 (%v -> %v): PMTU=%v", dec.IP.DstIP, dec.IP.SrcIP, mtu)
	return buf.Bytes(), nil
}

var (
	zeroMAC, _ = net.ParseMAC("00:00:00:00:00:00")
)

func (dec *EthernetDecoder) IsSpecial() bool {
	return dec.Eth.Length == 0 && dec.Eth.EthernetType == layers.EthernetTypeLLC &&
		bytes.Equal(zeroMAC, dec.Eth.SrcMAC) && bytes.Equal(zeroMAC, dec.Eth.DstMAC)
}

func (dec *EthernetDecoder) DF() bool {
	return len(dec.decoded) == 2 && (dec.IP.Flags&layers.IPv4DontFragment != 0)
}
