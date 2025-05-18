package util

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNullConn_Read(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.Read(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Read(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NullConn.Read(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestNullConn_Write(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.Write(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Write(%v) error = %v, wantErr %v", tt.args.b, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NullConn.Write(%v) = %v, want %v", tt.args.b, got, tt.want)
			}
		})
	}
}

func TestNullConn_Close(t *testing.T) {
	tests := []struct {
		name    string
		n       NullConn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_LocalAddr(t *testing.T) {
	tests := []struct {
		name string
		n    NullConn
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.LocalAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullConn.LocalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullConn_RemoteAddr(t *testing.T) {
	tests := []struct {
		name string
		n    NullConn
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.RemoteAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullConn.RemoteAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullConn_SetDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_SetReadDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetReadDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetReadDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNullConn_SetWriteDeadline(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name    string
		n       NullConn
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.SetWriteDeadline(tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("NullConn.SetWriteDeadline(%v) error = %v, wantErr %v", tt.args.t, err, tt.wantErr)
			}
		})
	}
}

func TestNewNullConn(t *testing.T) {
	tests := []struct {
		name string
		want net.Conn
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullConn(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullConn() = %v, want %v", got, tt.want)
			}
		})
	}
}
