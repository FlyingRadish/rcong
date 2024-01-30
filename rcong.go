package rcong

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type RCONConnection struct {
	conn     net.Conn
	password string
}

func Dial(ip string, port int, password string) (*RCONConnection, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	connection := &RCONConnection{conn: conn, password: password}
	//auth
	connection.Auth()
	return connection, nil
}

func (c *RCONConnection) Auth() (string, error) {
	packet := createPacket(DUMB_ID, SERVERDATA_AUTH, c.password)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, MAX_PACKET_SIZE)
	_, rErr := c.conn.Read(buf)
	if rErr != nil {
		return "", rErr
	}
	pkg, err := readPacket(buf)
	if err != nil {
		return "", errors.New("auth failed, wrong password")
	}

	if pkg.ID != DUMB_ID {
		return "", err
	}

	return pkg.Body, nil
}

func (c *RCONConnection) ExecCommand(command string) (string, error) {
	packet := createPacket(DUMB_ID, SERVERDATA_EXECCOMMAND, command)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, MAX_PACKET_SIZE)

	_, rErr := c.conn.Read(buf)
	if rErr != nil {
		return "", rErr
	}

	pkg, err := readPacket(buf)
	if err != nil {
		return "", err
	}
	return pkg.Body, nil
}

func (c *RCONConnection) Close() {
	c.conn.Close()
}

const (
	SERVERDATA_AUTH           int32 = 3
	SERVERDATA_AUTH_RESPONSE  int32 = 2
	SERVERDATA_EXECCOMMAND    int32 = 2
	SERVERDATA_RESPONSE_VALUE int32 = 0
	MAX_PACKET_SIZE           int32 = 4096

	DUMB_ID int32 = 0
)

type RCONPacket struct {
	Size int32
	ID   int32
	Type int32
	Body string
}

const HeaderLength = 10
const maximumPackageSize = 4096

func createPacket(id int32, pkgType int32, command string) []byte {
	commandBytes := []byte(command)
	// id:4  type:4  end:2
	size := int32(10 + len(commandBytes))

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, size)
	binary.Write(&buf, binary.LittleEndian, id)
	binary.Write(&buf, binary.LittleEndian, pkgType)
	buf.Write(commandBytes)
	buf.Write([]byte{0x00, 0x00})
	return buf.Bytes()
}

func readPacket(buf []byte) (RCONPacket, error) {
	packet := &RCONPacket{}
	packet.Size = int32(binary.LittleEndian.Uint32(buf[0:4]))
	packet.ID = int32(binary.LittleEndian.Uint32(buf[4:8]))
	packet.Type = int32(binary.LittleEndian.Uint32(buf[8:12]))
	bodyLength := packet.Size - 10
	packet.Body = string(buf[12 : 12+bodyLength])

	return *packet, nil
}
