package main

import (
	"fmt"
	"net"

	"github.com/google/uuid"
)

type connectionEvent struct {
	conn   net.Conn
	action string
	connId string
}

type MqttBroker struct {
	connections    map[string]net.Conn
	connectionChan chan connectionEvent
}

func newMqttBroker() *MqttBroker {
	broker := &MqttBroker{
		connections:    make(map[string]net.Conn),
		connectionChan: make(chan connectionEvent),
	}

	go broker.manageConnections()

	return broker
}

func (broker *MqttBroker) addConnection(conn net.Conn) {
	connId := uuid.New().String()

	broker.connectionChan <- connectionEvent{
		conn,
		"add",
		connId,
	}

	go broker.handleConnection(connId, conn)
}

func (broker *MqttBroker) removeConnection(connId string) {
	broker.connectionChan <- connectionEvent{
		conn:   nil,
		action: "remove",
		connId: connId,
	}
}

func (broker *MqttBroker) manageConnections() {
	for event := range broker.connectionChan {
		switch event.action {
		case "add":
			broker.connections[event.connId] = event.conn
			fmt.Println("Added Connection", event.connId)
		case "remove":
			delete(broker.connections, event.connId)
			fmt.Println("Removed connection", event.connId)
		}
	}
}

func (broker *MqttBroker) handleConnection(connId string, conn net.Conn) {
	b := make([]byte, 1024)
	// handle connection
	fmt.Println("Close Connection")
	n, err := conn.Read(b)

	conn.Write(createConnAckPacket(ConnackFlags{false}))

	if err != nil {
		fmt.Println("Read error")
		return
	}

	if n > 0 {
		fmt.Println("Found bytes", n)
		inspectBytes(b, n)
	} else {
		fmt.Println("zero bytes read")
	}

	conn.Close()
	broker.removeConnection(connId)
}

const PacketTypeReserved = 0
const PacketTypeConnect = 1
const PacketTypeConnack = 2
const PacketTypePublish = 3
const PacketTypePuback = 4
const PacketTypePubrec = 5
const PacketTypePubrel = 6
const PacketTypePubcomp = 7
const PacketTypeSubscribe = 8
const PacketTypeSuback = 9
const PacketTypeUnsubscribe = 10
const PacketTypeUnsuback = 11
const PacketTypePingreq = 12
const PacketTypePingresp = 13
const PacketTypeDisconnect = 14
const PacketTypeAuth = 15

var PacketTypeStrings = []string{
	"Reserved",    // 0
	"Connect",     // 1
	"Connack",     // 2
	"Publish",     // 3
	"Puback",      // 4
	"Pubrec",      // 5
	"Pubrel",      // 6
	"Pubcomp",     // 7
	"Subscribe",   // 8
	"Suback",      // 9
	"Unsubscribe", // 10
	"Unsuback",    // 11
	"Pingreq",     // 12
	"Pingresp",    // 13
	"Disconnect",  // 14
	"Auth",        // 15
}

func createConnAckPacket(ackFlags ConnackFlags) []byte {
	data := make([]byte, 4)

	data[0] = PacketTypeConnack << 4
	data[1] = 2
	data[2] = 1
	data[3] = 0

	return data
}

type ControlPacket struct {
	packetType uint8
	length     uint8
}

type ConnectVariableHeader struct {
	protocolVersion  byte
	connectFlags     byte
	keepAlive        uint16
	propertiesLength byte
	properties       map[byte][]byte
}

type ConnectPayload struct {
	clientId       string
	willProperties map[byte][]byte

	topic    string
	payload  []byte
	username string
	password string
}

type ConnectPacket struct {
	fixed    ControlPacket
	variable ConnectVariableHeader
	payload  ConnectPayload
}

type ConnackFlags struct {
	sessionPresent bool
}

type ConnackVariableHeader struct {
}

type ConnackPacket struct {
	fixex    ControlPacket
	variable ConnackVariableHeader
}

func readControlPacket(data []byte) ControlPacket {
	packetType := data[0] >> 4

	fmt.Println(PacketTypeStrings[packetType])

	fmt.Printf("Remaining Length %d\n", data[1])

	return ControlPacket{packetType, data[1]}
}

func inspectBytes(data []byte, n int) {
	for i := 0; i < 3; i++ {
		fmt.Printf("%b\n", data[i])
	}

	readControlPacket(data)
}

func main() {
	fmt.Println("MQTT Broker running on localhost:8080")

	broker := newMqttBroker()
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection Error")
		}

		broker.addConnection(conn)

	}
}
