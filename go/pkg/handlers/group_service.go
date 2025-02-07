package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"net/http"
	"strings"
)

func (h *Handlers) PostGroupService(w http.ResponseWriter, req *http.Request, data *types.PostGroupServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostGroupServiceResponse, error) {
	_, err := tx.Exec(`
		INSERT INTO dbtable_schema.group_services (group_id, service_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (group_id, service_id) DO NOTHING
	`, session.GroupId, data.GetServiceId(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/services")

	return &types.PostGroupServiceResponse{}, nil
}

func (h *Handlers) GetGroupServices(w http.ResponseWriter, req *http.Request, data *types.GetGroupServicesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetGroupServicesResponse, error) {
	var groupServices []*types.IGroupService

	err := h.Database.QueryRows(&groupServices, `
		SELECT TO_JSONB(es) as service, egs.id, egs."groupId", egs."serviceId"
		FROM dbview_schema.enabled_group_services egs
		LEFT JOIN dbview_schema.enabled_services es ON es.id = egs."serviceId"
		WHERE egs."groupId" = $1
	`, session.GroupId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetGroupServicesResponse{GroupServices: groupServices}, nil
}

func (h *Handlers) DeleteGroupService(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteGroupServiceResponse, error) {
	for _, serviceId := range strings.Split(data.GetIds(), ",") {
		_, err := tx.Exec(`
			DELETE FROM dbtable_schema.group_services
			WHERE group_id = $1 AND service_id = $2
		`, session.GroupId, serviceId)

		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/services")

	return &types.DeleteGroupServiceResponse{Success: true}, nil
}
