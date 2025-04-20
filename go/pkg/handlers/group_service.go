package handlers

import (
	"net/http"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/clients"
	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (h *Handlers) PostGroupService(w http.ResponseWriter, req *http.Request, data *types.PostGroupServiceRequest, session *types.UserSession, tx *clients.PoolTx) (*types.PostGroupServiceResponse, error) {
	var groupServiceId string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.group_services (group_id, service_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (group_id, service_id) DO NOTHING
		RETURNING id
	`, session.GroupId, data.GetServiceId(), session.UserSub).Scan(&groupServiceId)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/services")

	return &types.PostGroupServiceResponse{Id: groupServiceId}, nil
}

func (h *Handlers) GetGroupServices(w http.ResponseWriter, req *http.Request, data *types.GetGroupServicesRequest, session *types.UserSession, tx *clients.PoolTx) (*types.GetGroupServicesResponse, error) {
	var groupServices []*types.IGroupService

	err := h.Database.QueryRows(tx, &groupServices, `
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

func (h *Handlers) DeleteGroupService(w http.ResponseWriter, req *http.Request, data *types.DeleteGroupServiceRequest, session *types.UserSession, tx *clients.PoolTx) (*types.DeleteGroupServiceResponse, error) {

	serviceIds := strings.Split(data.Ids, ",")
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.group_services
		WHERE group_id = $1 AND service_id = ANY($2)
	`, session.GroupId, pq.Array(serviceIds))
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	_, err = h.DisableService(w, req, &types.DisableServiceRequest{Ids: data.Ids}, session, tx)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"group/services")

	return &types.DeleteGroupServiceResponse{Success: true}, nil
}
