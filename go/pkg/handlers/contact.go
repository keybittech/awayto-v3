package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/keybittech/awayto-v3/go/pkg/types"
	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func (h *Handlers) PostContact(w http.ResponseWriter, req *http.Request, data *types.PostContactRequest, session *types.UserSession, tx *sql.Tx) (*types.PostContactResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(tx, &contacts, `
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

func (h *Handlers) PatchContact(w http.ResponseWriter, req *http.Request, data *types.PatchContactRequest, session *types.UserSession, tx *sql.Tx) (*types.PatchContactResponse, error) {
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

func (h *Handlers) GetContacts(w http.ResponseWriter, req *http.Request, data *types.GetContactsRequest, session *types.UserSession, tx *sql.Tx) (*types.GetContactsResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(tx, &contacts, `
		SELECT * FROM dbview_schema.enabled_contacts
	`)

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.GetContactsResponse{Contacts: contacts}, nil
}

func (h *Handlers) GetContactById(w http.ResponseWriter, req *http.Request, data *types.GetContactByIdRequest, session *types.UserSession, tx *sql.Tx) (*types.GetContactByIdResponse, error) {
	var contacts []*types.Contact

	err := h.Database.QueryRows(tx, &contacts, `
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

func (h *Handlers) DeleteContact(w http.ResponseWriter, req *http.Request, data *types.DeleteContactRequest, session *types.UserSession, tx *sql.Tx) (*types.DeleteContactResponse, error) {
	_, err := tx.Exec(`
		DELETE FROM dbtable_schema.contacts
		WHERE id = $1
	`, data.GetId())

	if err != nil {
		return nil, util.ErrCheck(err)
	}

	return &types.DeleteContactResponse{Success: true}, nil
}

func (h *Handlers) DisableContact(w http.ResponseWriter, req *http.Request, data *types.DisableContactRequest, session *types.UserSession, tx *sql.Tx) (*types.DisableContactResponse, error) {
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
