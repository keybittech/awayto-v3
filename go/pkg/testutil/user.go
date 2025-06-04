package testutil

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TestUsersStruct struct {
	Client     *http.Client
	CookieData []http.Cookie
	*types.TestUser
}

func NewTestUser(userId, email, pass string) *TestUsersStruct {
	return &TestUsersStruct{
		CookieData: make([]http.Cookie, 0, 1),
		TestUser: &types.TestUser{
			TestEmail:  email,
			TestPass:   pass,
			TestUserId: userId,
			Profile: &types.IUserProfile{
				Email: email,
			},
		},
	}
}

func (tus *TestUsersStruct) GetUserSession(pool *pgxpool.Pool) (*types.ConcurrentUserSession, error) {
	var err error
	tus.Profile, err = tus.GetProfileDetails()
	if err != nil {
		return nil, util.ErrCheck(err)
	}
	session := &types.UserSession{
		UserSub: tus.Profile.GetSub(),
	}

	if len(tus.Profile.GetGroups()) > 0 {
		for code, group := range tus.Profile.GetGroups() {
			batch := util.NewBatchable(pool, "worker", "", 0)
			dbGroupReq := util.BatchQueryRow[types.IGroup](batch, `
				SELECT id, sub, external_id as "externalId"
				FROM dbtable_schema.groups
				WHERE code = $1
			`, code)
			subGroupReq := util.BatchQueryRow[types.IGroupRole](batch, `
				SELECT egr."externalId"
				FROM dbview_schema.enabled_roles er
				JOIN dbview_schema.enabled_group_roles egr ON egr."roleId" = er.id
				JOIN dbview_schema.enabled_groups eg ON eg.id = egr."groupId"
				WHERE er.name = $1 AND eg.name = $2
			`, tus.Profile.GetRoleName(), group.GetName())
			batch.Send(context.Background())

			dbGroup := *dbGroupReq
			subGroup := *subGroupReq

			session.GroupId = dbGroup.GetId()
			session.GroupPath = "/" + group.GetName()
			session.GroupCode = code
			session.GroupName = group.GetName()
			session.GroupExternalId = dbGroup.GetExternalId()
			session.GroupSub = dbGroup.GetSub()
			session.GroupAi = group.GetAi()
			session.SubGroupPath = session.GetGroupPath() + "/" + tus.Profile.GetRoleName()
			session.SubGroupName = tus.Profile.GetRoleName()
			session.SubGroupExternalId = subGroup.GetExternalId()
			break
		}
	}

	return types.NewConcurrentUserSession(session), nil
}

func (tus *TestUsersStruct) SetCookieData(cookies []*http.Cookie) {
	tus.CookieData = make([]http.Cookie, 0)
	for _, c := range cookies {
		tus.CookieData = append(tus.CookieData, *c)
	}
}

func (tus *TestUsersStruct) getUserClient() *http.Client {
	if tus.Client != nil {
		return tus.Client
	}
	if tus.CookieData == nil {
		log.Fatal("no cookie data to getUserClient with, did the user login?")
	}

	jar, _ := cookiejar.New(nil)
	appURL, _ := url.Parse(util.E_APP_HOST_URL)

	// convert regular struct data to pointer
	cookies := make([]*http.Cookie, 0, len(tus.CookieData))
	for _, c := range tus.CookieData {
		cookies = append(cookies, &c)
	}
	jar.SetCookies(appURL, cookies)
	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

func (tus *TestUsersStruct) apiRequest(method, path string, body []byte, queryParams map[string]string, responseObj proto.Message) error {
	reqURL := util.E_APP_HOST_URL + path

	if queryParams != nil && len(queryParams) > 0 {
		values := url.Values{}
		for k, v := range queryParams {
			values.Add(k, v)
		}
		reqURL = reqURL + "?" + values.Encode()
	}

	req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if body != nil && len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Add("X-Tz", "America/Los_Angeles")

	client := tus.getUserClient()
	resp, err := doAndRead(client, req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	if len(resp) > 0 && responseObj != nil {
		if err := protojson.Unmarshal(resp, responseObj); err != nil {
			return fmt.Errorf("error unmarshaling response body '%s' into responseObj: %w", string(resp), err)
		}
	}

	return nil
}

func (tus *TestUsersStruct) DoHandler(method, path string, body []byte, queryParams map[string]string, responseObj proto.Message) error {
	return tus.apiRequest(method, path, body, queryParams, responseObj)
}

func (tus *TestUsersStruct) GetProfileDetails() (*types.IUserProfile, error) {
	getProfileDetailsResponse := &types.GetUserProfileDetailsResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/profile/details", nil, nil, getProfileDetailsResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get user profile details error: %v", err))
	}

	return getProfileDetailsResponse.GetUserProfile(), nil
}

func (tus *TestUsersStruct) GetServiceById(serviceId string) (*types.IService, error) {
	getServiceByIdResponse := &types.GetServiceByIdResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/services/"+serviceId, nil, nil, getServiceByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get service by id error: %v", err))
	}
	if getServiceByIdResponse.Service.Id == "" {
		return nil, errors.New("get service by id response has no id")
	}

	return getServiceByIdResponse.Service, nil
}

func (tus *TestUsersStruct) GetScheduleById(scheduleId string) (*types.ISchedule, error) {
	getScheduleByIdResponse := &types.GetScheduleByIdResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/schedules/"+scheduleId, nil, nil, getScheduleByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get schedule by id error: %v", err))
	}
	if getScheduleByIdResponse.Schedule.Id == "" {
		return nil, errors.New("get schedule by id response has no id")
	}

	return getScheduleByIdResponse.Schedule, nil
}

func (tus *TestUsersStruct) GetMasterScheduleById(groupScheduleId string) (*types.IGroupSchedule, error) {
	getMasterScheduleByIdResponse := &types.GetGroupScheduleMasterByIdResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/group/schedules/master/"+groupScheduleId, nil, nil, getMasterScheduleByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get master schedule by id error: %v", err))
	}
	if getMasterScheduleByIdResponse.GroupSchedule.ScheduleId == "" {
		return nil, errors.New("get master schedule by id response has no schedule id")
	}

	return getMasterScheduleByIdResponse.GroupSchedule, nil
}

func (tus *TestUsersStruct) GetDateSlots(masterScheduleId string) ([]*types.IGroupScheduleDateSlots, error) {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	dateSlotsUrl := "/api/v1/group/schedules/" + masterScheduleId + "/date/" + firstOfMonth
	dateSlotsResponse := &types.GetGroupScheduleByDateResponse{}
	err := tus.apiRequest(http.MethodGet, dateSlotsUrl, nil, nil, dateSlotsResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get group date slots request, error: %v", err))
	}

	if len(dateSlotsResponse.GroupScheduleDateSlots) == 0 {
		return nil, errors.New(fmt.Sprintf("no date slots available to schedule %v", dateSlotsResponse))
	}

	return dateSlotsResponse.GroupScheduleDateSlots, nil
}

func (tus *TestUsersStruct) GetQuoteById(quoteId string) (*types.IQuote, error) {
	getQuoteByIdResponse := &types.GetQuoteByIdResponse{}
	err := tus.apiRequest(http.MethodGet, "/api/v1/quotes/"+quoteId, nil, nil, getQuoteByIdResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error get quote by id error: %v", err))
	}
	if getQuoteByIdResponse.Quote.Id == "" {
		return nil, errors.New("get quote by id response has no id")
	}

	return getQuoteByIdResponse.Quote, nil
}

func (tus *TestUsersStruct) PatchGroupUser(userSub, roleId string) error {
	patchGroupUserRequestBytes, err := protojson.Marshal(&types.PatchGroupUserRequest{
		UserSub: userSub,
		RoleId:  roleId,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling patch group user %s %s %v", userSub, roleId, err))
	}

	patchGroupUserResponse := &types.PatchGroupUserResponse{}
	err = tus.apiRequest(http.MethodPatch, "/api/v1/group/users", patchGroupUserRequestBytes, nil, patchGroupUserResponse)
	if err != nil {
		return errors.New(fmt.Sprintf("error patch group user request, sub: %s error: %v", userSub, err))
	}
	if !patchGroupUserResponse.Success {
		return errors.New("attach user internal was unsuccessful")
	}

	return nil
}

func (tus *TestUsersStruct) PatchGroupAssignments(roleFullName, actionName string) error {
	actions := make([]*types.IAssignmentAction, 1)
	actions[0] = &types.IAssignmentAction{
		Name: actionName,
	}
	assignmentActions := make(map[string]*types.IAssignmentActions)
	assignmentActions[roleFullName] = &types.IAssignmentActions{
		Actions: actions,
	}

	patchGroupAssignmentsBytes, err := protojson.Marshal(&types.PatchGroupAssignmentsRequest{
		Assignments: assignmentActions,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling patch group assignments %v %v", err, assignmentActions))
	}
	patchGroupAssignmentsResponse := &types.PatchGroupAssignmentsResponse{}
	err = tus.apiRequest(http.MethodPatch, "/api/v1/group/assignments", patchGroupAssignmentsBytes, nil, patchGroupAssignmentsResponse)
	if err != nil {
		return errors.New(fmt.Sprintf("error patch group assignments request: %v", err))
	}
	if !patchGroupAssignmentsResponse.Success {
		return errors.New(fmt.Sprintf("patch group assignments  was unsuccessful %v", patchGroupAssignmentsResponse))
	}

	return nil
}

func (tus *TestUsersStruct) PostSchedule(scheduleRequest *types.PostScheduleRequest) (*types.ISchedule, error) {
	scheduleRequestBytes, err := protojson.Marshal(scheduleRequest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling schedule request: %v", err))
	}

	scheduleResponse := &types.PostScheduleResponse{}
	err = tus.apiRequest(http.MethodPost, "/api/v1/schedules", scheduleRequestBytes, nil, scheduleResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post schedule request error: %v", err))
	}

	return tus.GetScheduleById(scheduleResponse.Id)
}

func (tus *TestUsersStruct) PostGroupSchedule(scheduleId string) error {
	scheduleRequestBytes, err := protojson.Marshal(&types.PostGroupScheduleRequest{
		ScheduleId: scheduleId,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error marshalling group schedule request: %v", err))
	}

	err = tus.apiRequest(http.MethodPost, "/api/v1/group/schedules", scheduleRequestBytes, nil, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("error post group schedule request error: %v", err))
	}

	return nil
}

func (tus *TestUsersStruct) PostQuote(serviceTierId string, slot *types.IGroupScheduleDateSlots, serviceForm, tierForm *types.IProtoFormVersionSubmission) (*types.IQuote, error) {
	if serviceForm == nil {
		serviceForm = &types.IProtoFormVersionSubmission{}
	}
	if tierForm == nil {
		tierForm = &types.IProtoFormVersionSubmission{}
	}

	postQuoteRequest := &types.PostQuoteRequest{
		SlotDate:                     slot.StartDate,
		ScheduleBracketSlotId:        slot.ScheduleBracketSlotId,
		ServiceTierId:                serviceTierId,
		ServiceFormVersionSubmission: serviceForm,
		TierFormVersionSubmission:    tierForm,
	}

	postQuoteBytes, err := protojson.Marshal(postQuoteRequest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling post quote request %v", err))
	}

	postQuoteResponse := &types.PostQuoteResponse{}
	err = tus.apiRequest(http.MethodPost, "/api/v1/quotes", postQuoteBytes, nil, postQuoteResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post quote request error: %v", err))
	}
	if postQuoteResponse.Quote.Id == "" {
		return nil, errors.New("no post quote id")
	}

	return postQuoteResponse.Quote, nil
}

func (tus *TestUsersStruct) PostBooking(bookingRequests []*types.IBooking) ([]*types.IBooking, error) {
	postBookingRequest := &types.PostBookingRequest{
		Bookings: bookingRequests,
	}

	postBookingBytes, err := protojson.Marshal(postBookingRequest)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshalling post booking request %v", err))
	}

	postBookingResponse := &types.PostBookingResponse{}
	err = tus.apiRequest(http.MethodPost, "/api/v1/bookings", postBookingBytes, nil, postBookingResponse)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error post booking request error: %v", err))
	}
	if len(postBookingResponse.Bookings) == 0 {
		return nil, errors.New("no bookings were created")
	}

	return postBookingResponse.Bookings, nil
}
