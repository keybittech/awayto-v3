package handlers

import (
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (h *Handlers) PostService(w http.ResponseWriter, req *http.Request, data *types.PostServiceRequest) (*types.PostServiceResponse, error) {
	session := h.Redis.ReqSession(req)
	service := data.GetService()
	tx, ongoing := h.Database.ReqTx(req)
	if tx == nil {
		return nil, util.ErrCheck(errors.New("bad post service tx"))
	}

	if !ongoing {
		defer tx.Rollback()
	}

	var serviceFormId *string
	if service.GetFormId() != "" {
		serviceFormId = &service.FormId
	}

	var serviceSurveyId *string
	if service.GetSurveyId() != "" {
		serviceSurveyId = &service.SurveyId
	}

	var serviceID string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.services (name, cost, form_id, survey_id, created_sub)
		VALUES ($1, $2::integer, $3, $4, $5::uuid)
		RETURNING id
	`, service.GetName(), service.GetCost(), serviceFormId, serviceSurveyId, session.UserSub).Scan(&serviceID)

	if err != nil {
		var dbErr *pq.Error
		if errors.As(err, &dbErr) && dbErr.Constraint == "services_name_created_sub_key" {
			return nil, util.ErrCheck(errors.New("you already have a service with the same name"))
		}
		return nil, util.ErrCheck(err)
	}

	for _, tier := range service.GetTiers() {
		var tierID string

		var tierFormId *string
		if tier.GetFormId() != "" {
			tierFormId = &tier.FormId
		}

		var tierSurveyId *string
		if tier.GetSurveyId() != "" {
			tierSurveyId = &tier.SurveyId
		}

		err := tx.QueryRow(`
			WITH input_rows(name, service_id, multiplier, form_id, survey_id, created_sub) as (VALUES ($1, $2::uuid, $3::decimal, $4::uuid, $5::uuid, $6::uuid)), ins AS (
				INSERT INTO dbtable_schema.service_tiers (name, service_id, multiplier, form_id, survey_id, created_sub)
				SELECT * FROM input_rows
				ON CONFLICT (name, service_id) DO NOTHING
				RETURNING id
			)
			SELECT id
			FROM ins
			UNION ALL
			SELECT st.id
			FROM input_rows
			JOIN dbtable_schema.service_tiers st USING (name, service_id)
		`, tier.GetName(), serviceID, tier.GetMultiplier(), tierFormId, tierSurveyId, session.UserSub).Scan(&tierID)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, addon := range tier.GetAddons() {
			_, err = tx.Exec(`
				INSERT INTO dbtable_schema.service_tier_addons (service_addon_id, service_tier_id, created_sub)
				VALUES ($1, $2, $3::uuid)
				ON CONFLICT (service_addon_id, service_tier_id) DO NOTHING
			`, addon.GetId(), tierID, session.UserSub)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	if !ongoing {
		tx.Commit()
	}

	return &types.PostServiceResponse{Id: serviceID}, nil
}

func (h *Handlers) PatchService(w http.ResponseWriter, req *http.Request, data *types.PatchServiceRequest) (*types.PatchServiceResponse, error) {
	session := h.Redis.ReqSession(req)
	service := data.GetService()

	_, err := h.Database.Client().Exec(`
		UPDATE dbtable_schema.services
		SET name = $2, updated_sub = $3, updated_on = $4 
		WHERE id = $1
	`, service.GetId(), service.GetName(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchServiceResponse{Success: true}, nil
}

func (h *Handlers) GetServices(w http.ResponseWriter, req *http.Request, data *types.GetServicesRequest) (*types.GetServicesResponse, error) {
	session := h.Redis.ReqSession(req)
	var services []*types.IService

	err := h.Database.QueryRows(&services, `
		SELECT * FROM dbtable_schema.services
		WHERE created_sub = $1
	`, session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetServicesResponse{Services: services}, nil
}

func (h *Handlers) GetServiceById(w http.ResponseWriter, req *http.Request, data *types.GetServiceByIdRequest) (*types.GetServiceByIdResponse, error) {
	var services []*types.IService

	err := h.Database.QueryRows(&services, `
		SELECT * FROM dbview_schema.enabled_services_ext
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(services) == 0 {
		return nil, util.ErrCheck(errors.New("service not found"))
	}

	return &types.GetServiceByIdResponse{Service: services[0]}, nil
}

func (h *Handlers) DeleteService(w http.ResponseWriter, req *http.Request, data *types.DeleteServiceRequest) (*types.DeleteServiceResponse, error) {
	session := h.Redis.ReqSession(req)
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := h.Database.Client().Exec(`
			DELETE FROM dbtable_schema.services
			WHERE id = $1
		`, id)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Redis.Client().Del(req.Context(), session.UserSub+"services/"+id)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"services")

	return &types.DeleteServiceResponse{Success: true}, nil
}

func (h *Handlers) DisableService(w http.ResponseWriter, req *http.Request, data *types.DisableServiceRequest) (*types.DisableServiceResponse, error) {
	session := h.Redis.ReqSession(req)
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := h.Database.Client().Exec(`
			UPDATE dbtable_schema.services
			SET enabled = false, updated_on = $2, updated_sub = $3
			WHERE id = $1
		`, id, time.Now().Local().UTC(), session.UserSub)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		h.Redis.Client().Del(req.Context(), session.UserSub+"services/"+id)
	}

	h.Redis.Client().Del(req.Context(), session.UserSub+"services")

	return &types.DisableServiceResponse{Success: true}, nil
}
