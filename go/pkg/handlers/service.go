package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
)

func (h *Handlers) PostService(w http.ResponseWriter, req *http.Request, data *types.PostServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostServiceResponse, error) {
	service := data.GetService()

	var serviceFormId *string
	if service.GetFormId() != "" {
		serviceFormId = &service.FormId
	}

	var serviceSurveyId *string
	if service.GetSurveyId() != "" {
		serviceSurveyId = &service.SurveyId
	}

	var serviceId string
	err := tx.QueryRow(`
		INSERT INTO dbtable_schema.services (name, cost, form_id, survey_id, created_sub)
		VALUES ($1, $2::integer, $3, $4, $5::uuid)
		RETURNING id
	`, service.GetName(), service.GetCost(), serviceFormId, serviceSurveyId, session.UserSub).Scan(&serviceId)

	if err != nil {
		var dbErr *pq.Error
		if errors.As(err, &dbErr) && dbErr.Constraint == "services_name_created_sub_key" {
			return nil, util.ErrCheck(util.UserError("A service with the same name already exists."))
		}
		return nil, util.ErrCheck(err)
	}

	return &types.PostServiceResponse{Id: serviceId}, nil
}

func (h *Handlers) PatchService(w http.ResponseWriter, req *http.Request, data *types.PatchServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchServiceResponse, error) {
	service := data.GetService()

	var serviceFormId *string
	if service.GetFormId() != "" {
		serviceFormId = &service.FormId
	}

	var serviceSurveyId *string
	if service.GetSurveyId() != "" {
		serviceSurveyId = &service.SurveyId
	}

	rows, err := tx.Query(`
		SELECT st.id
		FROM dbtable_schema.service_tiers st
		WHERE st.service_id = $1
	`, service.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	defer rows.Close()

	for rows.Next() {
		var tierId string
		err = rows.Scan(&tierId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}

		// delete any existing addons using tier ids dbtable_schema.service_tier_addons
		_, err = tx.Exec(`
			DELETE FROM dbtable_schema.service_tier_addons
			WHERE service_tier_id = $1
		`, tierId)
		if err != nil {
			return nil, util.ErrCheck(err)
		}
	}

	// delete any tiers with serviceId dbtable_schema.service_tiers
	_, err = tx.Exec(`
		DELETE FROM dbtable_schema.service_tiers
		WHERE service_id = $1
	`, service.GetId())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// update PatchService
	_, err = tx.Exec(`
		UPDATE dbtable_schema.services
		SET name = $2, form_id = $3, survey_id = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
	`, service.GetId(), service.GetName(), serviceFormId, serviceSurveyId, session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	// build tiers
	for _, tier := range service.GetTiers() {
		var tierId string

		var tierFormId *string
		if tier.GetFormId() != "" {
			tierFormId = &tier.FormId
		}

		var tierSurveyId *string
		if tier.GetSurveyId() != "" {
			tierSurveyId = &tier.SurveyId
		}

		err = tx.QueryRow(`
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
		`, tier.GetName(), service.GetId(), tier.GetMultiplier(), tierFormId, tierSurveyId, session.UserSub).Scan(&tierId)

		if err != nil {
			return nil, util.ErrCheck(err)
		}

		for _, addon := range tier.GetAddons() {
			_, err = tx.Exec(`
				INSERT INTO dbtable_schema.service_tier_addons (service_addon_id, service_tier_id, created_sub)
				VALUES ($1, $2, $3::uuid)
				ON CONFLICT (service_addon_id, service_tier_id) DO NOTHING
			`, addon.GetId(), tierId, session.UserSub)
			if err != nil {
				return nil, util.ErrCheck(err)
			}
		}
	}

	return &types.PatchServiceResponse{Success: true}, nil
}

func (h *Handlers) GetServices(w http.ResponseWriter, req *http.Request, data *types.GetServicesRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetServicesResponse, error) {
	var services []*types.IService

	err := tx.QueryRows(&services, `
		SELECT * FROM dbtable_schema.services
		WHERE created_sub = $1
	`, session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetServicesResponse{Services: services}, nil
}

func (h *Handlers) GetServiceById(w http.ResponseWriter, req *http.Request, data *types.GetServiceByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetServiceByIdResponse, error) {
	var services []*types.IService

	err := tx.QueryRows(&services, `
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

func (h *Handlers) DeleteService(w http.ResponseWriter, req *http.Request, data *types.DeleteServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteServiceResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
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

func (h *Handlers) DisableService(w http.ResponseWriter, req *http.Request, data *types.DisableServiceRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableServiceResponse, error) {
	ids := strings.Split(data.GetIds(), ",")

	for _, id := range ids {
		_, err := tx.Exec(`
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
