package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"google.golang.org/protobuf/encoding/protojson"
)

type IntegrationTest struct {
	UserIds       map[string]int
	UserSessions  []*types.UserSession
	Connections   map[string]net.Conn
	Roles         []*types.IRole
	DefaultRole   *types.IRole
	Group         *types.IGroup
	GroupService  *types.IGroupService
	GroupSchedule *types.IGroupSchedule
	UserSchedule  *types.ISchedule
	Quote         *types.IQuote
	Booking       *types.IBooking
}

var integrationTest *IntegrationTest

func init() {
	err := flag.Set("log", "debug")
	if err != nil {
		log.Fatal(err)
	}
}

func reset(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
}

func TestMain(m *testing.M) {
	integrationTest = &IntegrationTest{}

	go main()
	setupSockServer()

	exitCode := m.Run()
	os.Exit(exitCode)
}

func Test_main(t *testing.T) {
	if !testing.Short() {
		t.Run("server runs for 5 seconds", func(t *testing.T) {
			time.Sleep(5 * time.Second)
		})
	} else {
		time.Sleep(2 * time.Second)
	}
}

func TestIntegrationUser(t *testing.T) {
	userId := int(time.Now().UnixNano())

	registerKeycloakUserViaForm(userId)

	session, connection := getUser(userId)

	integrationTest.UserIds = map[string]int{session.UserSub: userId}
	integrationTest.UserSessions = []*types.UserSession{session}
	integrationTest.Connections = map[string]net.Conn{session.UserSub: connection}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationGroup(t *testing.T) {
	admin := integrationTest.UserSessions[0]
	userId := integrationTest.UserIds[admin.UserSub]
	token, err := getKeycloakToken(userId)
	if err != nil {
		t.Fatal(err)
	}
	groupRequest := &types.PostGroupRequest{
		Name:           "the_test_group_" + strconv.Itoa(userId),
		DisplayName:    "The Test Group #" + strconv.Itoa(userId),
		Ai:             true,
		Purpose:        "integration testing group",
		AllowedDomains: "",
	}
	requestBytes, err := protojson.Marshal(groupRequest)
	if err != nil {
		t.Fatal(err)
	}

	postGroupResponse := &types.PostGroupResponse{}
	err = apiRequest(token, http.MethodPost, "/api/v1/group", requestBytes, nil, postGroupResponse)
	if err != nil {
		t.Fatal(err)
	}

	if postGroupResponse.Id == "" {
		t.Fatal("integration failed to make group")
	}

	integrationTest.Group = &types.IGroup{
		Id:             postGroupResponse.Id,
		Name:           groupRequest.Name,
		DisplayName:    groupRequest.DisplayName,
		Ai:             groupRequest.Ai,
		Purpose:        groupRequest.Purpose,
		AllowedDomains: groupRequest.AllowedDomains,
	}

	println(fmt.Sprintf("Integration Test: %+v", integrationTest))
}

func TestIntegrationRoles(t *testing.T) {

}

func TestIntegrationGroupService(t *testing.T) {

}

func TestIntegrationGroupSchedule(t *testing.T) {

}

func TestIntegrationOnboarding(t *testing.T) {

}

func TestIntegrationUserSchedule(t *testing.T) {

}

func TestIntegrationQuotes(t *testing.T) {

}

func TestIntegrationBookings(t *testing.T) {

}

func TestIntegrationExchange(t *testing.T) {

}

// func BenchmarkBoolFormat(b *testing.B) {
// 	b.ReportAllocs()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
//
// 		// For true
// 		var v bool = true
// 		var expected rune = 't'
// 		var actual rune
//
// 		// Zero-allocation way to get first char of bool
// 		if v {
// 			actual = 't'
// 		} else {
// 			actual = 'f'
// 		}
//
// 		if actual != expected {
// 			b.Fail()
// 		}
// 	}
// }
//
// func BenchmarkBoolAllocate(b *testing.B) {
// 	b.ReportAllocs()
// 	reset(b)
// 	for c := 0; c < b.N; c++ {
// 		if false {
// 			_ = "t"
// 		}
// 	}
// }
