package handlers

import (
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
	"github.com/lib/pq"
)

func (h *Handlers) PostGroupService(info ReqInfo, data *types.PostGroupServiceRequest) (*types.PostGroupServiceResponse, error) {
	var groupServiceId string
	err := info.Tx.QueryRow(info.Ctx, `
		INSERT INTO dbtable_schema.group_services (group_id, service_id, created_sub)
		VALUES ($1, $2, $3::uuid)
		ON CONFLICT (group_id, service_id) DO NOTHING
		RETURNING id
	`, info.Session.GetGroupId(), data.ServiceId, info.Session.GetUserSub()).Scan(&groupServiceId)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/services")

	return &types.PostGroupServiceResponse{Id: groupServiceId}, nil
}

func (h *Handlers) GetGroupServices(info ReqInfo, data *types.GetGroupServicesRequest) (*types.GetGroupServicesResponse, error) {
	groupServices := util.BatchQuery[types.IGroupService](info.Batch, `
		SELECT TO_JSONB(es) as service, egs.id, egs."groupId", egs."serviceId"
		FROM dbview_schema.enabled_group_services egs
		LEFT JOIN dbview_schema.enabled_services es ON es.id = egs."serviceId"
		WHERE egs."groupId" = $1
	`, info.Session.GetGroupId())

	info.Batch.Send(info.Ctx)

	return &types.GetGroupServicesResponse{GroupServices: *groupServices}, nil
}

func (h *Handlers) DeleteGroupService(info ReqInfo, data *types.DeleteGroupServiceRequest) (*types.DeleteGroupServiceResponse, error) {
	serviceIds := strings.Split(data.Ids, ",")

	util.BatchExec(info.Batch, `
		DELETE FROM dbtable_schema.group_services
		WHERE group_id = $1 AND service_id = ANY($2)
	`, info.Session.GetGroupId(), pq.Array(serviceIds))

	info.Batch.Send(info.Ctx)
	info.Batch.Reset()

	h.DeleteService(info, &types.DeleteServiceRequest{Ids: data.Ids})

	h.Redis.Client().Del(info.Ctx, info.Session.GetUserSub()+"group/services")

	return &types.DeleteGroupServiceResponse{Success: true}, nil
}
