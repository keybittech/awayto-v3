package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/keybittech/awayto-v3/go/pkg/interfaces"
	"github.com/keybittech/awayto-v3/go/pkg/types"
)

func TestHandlers_PostGroupSeat(t *testing.T) {
	type args struct {
		w       http.ResponseWriter
		r       *http.Request
		data    *types.PostGroupSeatRequest
		session *types.UserSession
		tx      interfaces.IDatabaseTx
	}
	tests := []struct {
		name    string
		h       *Handlers
		args    args
		want    *types.PostGroupSeatResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.PostGroupSeat(tt.args.w, tt.args.r, tt.args.data, tt.args.session, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handlers.PostGroupSeat(%v, %v, %v, %v, %v) error = %v, wantErr %v", tt.args.w, tt.args.r, tt.args.data, tt.args.session, tt.args.tx, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handlers.PostGroupSeat(%v, %v, %v, %v, %v) = %v, want %v", tt.args.w, tt.args.r, tt.args.data, tt.args.session, tt.args.tx, got, tt.want)
			}
		})
	}
}
