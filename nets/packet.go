package nets

import "log"

/*
 * client packet interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */


/*
 * Decode packet interface
 *
 * Original packet data format:
 * length(ushort)+data(bytes)
 * first 2byte storage 6 + len(data)
 */

//internal macro variables
const (
	SortHeadSize = 2
	PacketHeadSize = 6
	//CommandIdPosStart = 2
	//CommandIdPosEnd = 4
)

//client one pack info
type ClientPack struct {
	PlayerId int64
	KeyDataId string
	CommandId int
	BodySize int
	HeaderPack []byte
	BodyPack []byte
}

//packet info
type Packet struct {
}

//construct
func NewPacket() *Packet {
	this := &Packet{}
	return this
}

//get pack head size
func (p *Packet) GetPackHeadSize() int {
	return PacketHeadSize
}

//generate final packet
func (p *Packet) GenFinalPack(body []byte) []byte {
	var result = make([]byte, 0)
	if len(body) <= 0 {
		return result
	}
	//generate pack info
	bodySize := int32(len(body))
	packHeader := p.zipHeader(bodySize)
	packet := p.encodePacket(packHeader, body)
	return packet
}

//decode packet header
//return bodySize
func (p *Packet) DecodePackHeader(header[]byte) int {
	var (
		bodySize int
		packetSize int
	)
	if len(header) <= 0 {
		return bodySize
	}

	packetSize = p.readShort(header[0:SortHeadSize])
	bodySize = packetSize - PacketHeadSize
	//command := header[CommandIdPosStart:CommandIdPosEnd]
	//commandId = p.ReadShort(command)
	return bodySize
}


////////////////
//private func
///////////////

//generate final packet
func (p *Packet) encodePacket(header, body []byte) []byte {
	var (
		headerSize int
		bodySize int
		packetSize int
		x, y int
	)

	headerSize = len(header)
	bodySize = len(body)

	log.Println("Packet::encodePacket, headerSize:", headerSize, ", bodySize:", bodySize)

	//init final packet buff
	packetSize = headerSize + bodySize
	packet := make([]byte, packetSize)

	//var i, k int
	if headerSize <= 0 || bodySize <= 0 {
		packet = packet[0:0]
		return packet
	}

	//process header
	for x = 0; x < headerSize; x++ {
		packet[x] = header[x]
	}
	//process body
	for y = 0; y < bodySize; y++ {
		packet[x] = body[y]
		x++
	}
	//return packet[0:packetSize]
	return packet
}


//zip header
func (p *Packet) zipHeader(bodySize int32) []byte {
	header := make([]byte, PacketHeadSize)
	tempSize := PacketHeadSize + bodySize

	//packet header info
	//total len
	header[0] = byte((tempSize & 0xff00) >> 8)
	header[1] = byte(tempSize & 0xff)

	//command id
	//header[2] = byte((commandId & 0xff00) >> 8)
	//header[3] = byte(commandId & 0xff)

	return header
}

//write short int into 2bytes space
func (p *Packet) writeShort(value int32) []byte {
	shortHeader := make([]byte, SortHeadSize)
	shortHeader[0] = byte((value & 0xff00) >> 8)
	shortHeader[1] = byte(value & 0xff)
	return shortHeader
}

//read short int from 2 bytes space
func (p *Packet) readShort(header []byte) int {
	var value int
	if len(header) < SortHeadSize {
		return value
	}
	//use 2bytes storage int, need convert high position value!!!!
	value |= int(int32(header[0]) << 8)
	value |= int(header[1])
	return value
}

