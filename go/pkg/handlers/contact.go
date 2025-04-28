package handlers

import (
	"errors"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostContact(info ReqInfo, data *types.PostContactRequest) (*types.PostContactResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &contacts, `
		INSERT INTO dbtable_schema.contacts (name, email, phone, created_sub)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, phone
	`, data.GetName(), data.GetEmail(), data.GetPhone(), info.Session.UserSub)
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	if len(contacts) == 0 {
		return nil, util.ErrCheck(errors.New("failed to insert contact"))
	}

	return &types.PostContactResponse{Id: contacts[0].GetId()}, nil
}

func (h *Handlers) PatchContact(info ReqInfo, data *types.PatchContactRequest) (*types.PatchContactResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.contacts
		SET name = $2, email = $3, phone = $4, updated_sub = $5, updated_on = $6
		WHERE id = $1
		RETURNING id, name, email, phone
	`, data.GetId(), data.GetName(), data.GetEmail(), data.GetPhone(), info.Session.UserSub, time.Now().Local().UTC())
	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.PatchContactResponse{Success: true}, nil
}

func (h *Handlers) GetContacts(info ReqInfo, data *types.GetContactsRequest) (*types.GetContactsResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &contacts, `
		SELECT * FROM dbview_schema.enabled_contacts
	`)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetContactsResponse{Contacts: contacts}, nil
}

func (h *Handlers) GetContactById(info ReqInfo, data *types.GetContactByIdRequest) (*types.GetContactByIdResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(info.Req.Context(), info.Tx, &contacts, `
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

func (h *Handlers) DeleteContact(info ReqInfo, data *types.DeleteContactRequest) (*types.DeleteContactResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		DELETE FROM dbtable_schema.contacts
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteContactResponse{Success: true}, nil
}

func (h *Handlers) DisableContact(info ReqInfo, data *types.DisableContactRequest) (*types.DisableContactResponse, error) {
	_, err := info.Tx.Exec(info.Req.Context(), `
		UPDATE dbtable_schema.contacts
		SET enabled = false, updated_on = $2, updated_sub = $3
		WHERE id = $1
	`, data.GetId(), time.Now().Local().UTC(), info.Session.UserSub)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DisableContactResponse{Success: true}, nil
}
