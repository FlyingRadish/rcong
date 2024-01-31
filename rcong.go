package rcong

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

type RCONConnection struct {
	conn       net.Conn
	ip         string
	port       int
	password   string
	retryCount int
	retryDelay int
}

func NewRCONConnection(ip string, port int, password string, retryCount int, retryDelay int) *RCONConnection {
	connection := &RCONConnection{
		ip:         ip,
		port:       port,
		password:   password,
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
	return connection
}

func (c *RCONConnection) Connect() error {
	c.Close()
	address := fmt.Sprintf("%s:%d", c.ip, c.port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	c.conn = conn
	// auth
	if len(c.password) > 0 {
		c.auth()
	}
	return nil
}

func (c *RCONConnection) auth() (string, error) {
	packet := createPacket(dumb_ID, serverDATA_AUTH, c.password)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, max_PACKET_SIZE)
	_, rErr := c.conn.Read(buf)
	if rErr != nil {
		return "", rErr
	}
	pkg, err := readPacket(buf)
	if err != nil {
		return "", errors.New("auth failed, wrong password")
	}

	if pkg.ID != dumb_ID {
		return "", err
	}

	return pkg.Body, nil
}

func (c *RCONConnection) ExecCommand(command string) (string, error) {
	return c.execCommandImp(command, c.retryCount)
}

func (c *RCONConnection) execCommandImp(command string, retryCount int) (string, error) {
	if c.conn == nil {
		if c.retryCount > 0 {
			if c.retryDelay > 0 {
				time.Sleep(time.Duration(c.retryDelay) * time.Second)
			}
			c.Connect()
			return c.ExecCommandImp(command, retryCount-1)
		} else {
			return "", fmt.Errorf("RCON connection is not established")
		}
	}
	resp, err := c.execute(command)
	if err != nil {
		if retryCount > 0 {
			if c.retryDelay > 0 {
				time.Sleep(time.Duration(c.retryDelay) * time.Second)
			}
			c.Connect()
			return c.ExecCommandImp(command, retryCount-1)
		} else {
			return "", err
		}
	}
	return resp, nil
}

func (c *RCONConnection) execute(command string) (string, error) {
	packet := createPacket(dumb_ID, serverDATA_EXECCOMMAND, command)
	_, wRrr := c.conn.Write(packet)
	if wRrr != nil {
		return "", wRrr
	}

	buf := make([]byte, max_PACKET_SIZE)

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
	if c.conn != nil {
		c.conn.Close()
	}
}

const (
	serverDATA_AUTH           int32 = 3
	serverDATA_AUTH_RESPONSE  int32 = 2
	serverDATA_EXECCOMMAND    int32 = 2
	serverDATA_RESPONSE_VALUE int32 = 0
	max_PACKET_SIZE           int32 = 4096

	dumb_ID int32 = 0
)

type rconPacket struct {
	Size int32
	ID   int32
	Type int32
	Body string
}

const headerLength = 10
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

func readPacket(buf []byte) (rconPacket, error) {
	packet := &rconPacket{}
	packet.Size = int32(binary.LittleEndian.Uint32(buf[0:4]))
	packet.ID = int32(binary.LittleEndian.Uint32(buf[4:8]))
	packet.Type = int32(binary.LittleEndian.Uint32(buf[8:12]))
	bodyLength := packet.Size - 10
	packet.Body = string(buf[12 : 12+bodyLength])

	return *packet, nil
}
