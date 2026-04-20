package testutil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

//		vaultResp := &types.GetVaultKeyResponse{}
//		// Reuse apiRequest to ensure headers (UA, TZ) match the session
//		err := tus.apiRequest(http.MethodGet, "/api/v1/vault/key", nil, nil, vaultResp)
//		if err != nil {
//			return err
//		}
//
//		keyBytes, err := base64.StdEncoding.DecodeString(vaultResp.Key)
//		if err != nil {
//			return err
//		}
//
//		tus.VaultKey = keyBytes
//		tus.VaultSessionId = vaultResp.Sid
//		return nil
//	}
func (tus *TestUsersStruct) GetSocketTicket() error {

	ticketResponse := &types.GetSocketTicketResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/sock/ticket", nil, nil, ticketResponse)
	if err != nil {
		return fmt.Errorf("could not make ticket request: %v", err)
	}

	ticket := ticketResponse.GetTicket()

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	tus.TestUser.TestTicket = ticket
	tus.TestUser.TestConnId = connId

	return nil
}

func (tus *TestUsersStruct) GetSocketConnection() error {
	dialer := websocket.Dialer{
		TLSClientConfig: TestTLS,
	}

	url := "wss://" + util.E_APP_HOST_NAME + "/sock?ticket=" + tus.GetTestTicket()

	sockConn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to dial socket tcp: %v", err)
	}

	tus.Socket = sockConn

	return nil
}

func (tus *TestUsersStruct) WriteSocketMessage(data []byte) error {
	return tus.Socket.WriteMessage(websocket.TextMessage, data)
}
