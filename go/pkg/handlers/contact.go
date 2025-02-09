package handlers

import (
	"av3api/pkg/clients"
	"av3api/pkg/types"
	"av3api/pkg/util"
	"errors"
	"net/http"
	"time"
)

func (h *Handlers) PostContact(w http.ResponseWriter, req *http.Request, data *types.PostContactRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PostContactResponse, error) {
	var contacts []*types.Contact

	err := tx.QueryRows(&contacts, `
		INSERT INTO dbtable_schema.contacts (name, email, phone, created_sub)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, phone
	`, data.GetName(), data.GetEmail(), data.GetPhone(), session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(contacts) == 0 {
		return nil, util.ErrCheck(errors.New("failed to insert contact"))
	}

	return &types.PostContactResponse{Id: contacts[0].GetId()}, nil
}

func (h *Handlers) PatchContact(w http.ResponseWriter, req *http.Request, data *types.PatchContactRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.PatchContactResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.contacts
		SET name = $2, email = $3, phone = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
		RETURNING id, name, email, phone
	`, data.GetId(), data.GetName(), data.GetEmail(), data.GetPhone(), session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchContactResponse{Success: true}, nil
}

func (h *Handlers) GetContacts(w http.ResponseWriter, req *http.Request, data *types.GetContactsRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetContactsResponse, error) {
	var contacts []*types.Contact

	err := tx.QueryRows(&contacts, `
		SELECT * FROM dbview_schema.enabled_contacts
	`)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetContactsResponse{Contacts: contacts}, nil
}

func (h *Handlers) GetContactById(w http.ResponseWriter, req *http.Request, data *types.GetContactByIdRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.GetContactByIdResponse, error) {
	var contacts []*types.Contact

	err := tx.QueryRows(&contacts, `
		SELECT * FROM dbview_schema.enabled_contacts
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(contacts) == 0 {
		return nil, util.ErrCheck(errors.New("contact not found"))
	}

	return &types.GetContactByIdResponse{Contact: contacts[0]}, nil
}

func (h *Handlers) DeleteContact(w http.ResponseWriter, req *http.Request, data *types.DeleteContactRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DeleteContactResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.contacts
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteContactResponse{Success: true}, nil
}

func (h *Handlers) DisableContact(w http.ResponseWriter, req *http.Request, data *types.DisableContactRequest, session *clients.UserSession, tx clients.IDatabaseTx) (*types.DisableContactResponse, error) {
	_, err := tx.Exec(`
		UPDATE dbtable_schema.contacts
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableContactResponse{Success: true}, nil
}
