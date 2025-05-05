package api

import (
	"fmt"
	"io"
	"net"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (a *API) InitTurnTCP(listenerPort, internalPort int) {

	listenerStr := fmt.Sprintf(":%d", listenerPort)

	turnTCPListener, err := net.Listen("tcp", listenerStr)
	if err != nil {
		panic(err)
	}

	defer turnTCPListener.Close()

	fmt.Println("Listening on", listenerStr)

	for {
		println("in accept")
		conn, err := turnTCPListener.Accept()
		println("did accept")
		if err != nil {
			fmt.Printf("Error accepting tcp %s connection: %s", listenerStr, err.Error())
			continue
		}

		fmt.Printf("Got TCP CONN %+v", conn)

		go HandleTurnTCPConnection(conn, internalPort)
	}
}

func HandleTurnTCPConnection(conn net.Conn, internalPort int) {
	defer conn.Close()

	println("did handle")
	turnTcp, err := net.Dial("tcp", fmt.Sprintf(":%d", internalPort))
	if err != nil {
		util.ErrCheck(err)
	}
	defer turnTcp.Close()

	go io.Copy(turnTcp, conn)
	io.Copy(conn, turnTcp)
}
