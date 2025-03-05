package api

import (
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"fmt"
	"io"
	"net"
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

func (a *API) InitTurnUDP(listenerPort, internalPort int) {

	listenerStr := fmt.Sprintf(":%d", listenerPort)
	listenerAddr, err := net.ResolveUDPAddr("udp", listenerStr)
	if err != nil {
		panic(err)
	}

	internalStr := fmt.Sprintf(":%d", internalPort)
	internalAddr, err := net.ResolveUDPAddr("udp", internalStr)
	if err != nil {
		panic(err)
	}

	turnUDPListener, err := net.ListenUDP("udp", listenerAddr)
	if err != nil {
		panic(err)
	}
	defer turnUDPListener.Close()

	turnUDPInternal, err := net.DialUDP("udp", nil, internalAddr)
	if err != nil {
		panic(err)
	}
	defer turnUDPInternal.Close()

}

func HandleTurnUDPConnection(listenerConn, internalConn *net.UDPConn, listenerAddr, internalAddr *net.UDPAddr) {

	// buffer := make([]byte, 2048)

	for {
		// n, remoteAddr, err := listenerConn.ReadFromUDP(buffer)
	}

}
