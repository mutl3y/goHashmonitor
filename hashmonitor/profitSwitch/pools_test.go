package profitSwitch

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNewPoolsFile(t *testing.T) {
	tests := []struct {
		name string
		want *poolsFile
	}{
		{"", &poolsFile{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPoolsFile(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPoolsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_poolsFile_Write(t *testing.T) {
	var pathSep = string(os.PathSeparator)
	file := strings.Join([]string{".", "pools.txt"}, pathSep)
	f, err := os.OpenFile("pools.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("Can't find File %v, %v", file, err)
	}
	defer f.Close()

	type fields struct {
		PoolAddress    string
		WalletAddress  string
		RigID          string
		PoolPassword   string
		UseNicehash    bool
		UseTLS         bool
		TLSFingerprint string
		PoolWeight     int
	}
	type args struct {
		rwc io.WriteCloser
	}
	tests := []struct {
		name    string
		fields  fields
		r       io.WriteCloser
		wantErr bool
	}{
		{"", fields{
			PoolAddress:   "minexcash.com:7777",
			WalletAddress: "XCA1cLoiKCx79R156KpSR3BgUthyEnuuFGi96G5rFfsdjLQSe2shU2aBWNdKXDFMpTfjvmeNnVNHNLXvuKmiNSCJ8yczZktCGc",
			RigID:         "testing",
			PoolPassword:  "x",
			PoolWeight:    1,
		},
			f,
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolsFile()
			p.algo = "cryptonight_v8_double"
			p.pools = []poolsEntry{
				{
					PoolAddress:    tt.fields.PoolAddress,
					WalletAddress:  tt.fields.WalletAddress,
					RigID:          tt.fields.RigID,
					PoolPassword:   tt.fields.PoolPassword,
					UseNicehash:    tt.fields.UseNicehash,
					UseTLS:         tt.fields.UseTLS,
					TLSFingerprint: tt.fields.TLSFingerprint,
					PoolWeight:     tt.fields.PoolWeight,
				},
				{
					PoolAddress:    tt.fields.PoolAddress,
					WalletAddress:  tt.fields.WalletAddress,
					RigID:          tt.fields.RigID,
					PoolPassword:   tt.fields.PoolPassword,
					UseNicehash:    tt.fields.UseNicehash,
					UseTLS:         tt.fields.UseTLS,
					TLSFingerprint: tt.fields.TLSFingerprint,
					PoolWeight:     tt.fields.PoolWeight,
				},
			}

			if err := p.Write(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("poolsFile.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_poolsFile_ReadFile(t *testing.T) {
	var pathSep = string(os.PathSeparator)
	file := strings.Join([]string{".", "pools.txt"}, pathSep)
	rwc, err := os.OpenFile("pools.txt", os.O_RDONLY, 0666)
	if err != nil {
		t.Fatalf("Can't find File %v, %v", file, err)
	}

	tests := []struct {
		name    string
		rwc     io.ReadCloser
		wantErr bool
	}{
		{"", rwc, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolsFile()
			err = p.Read(tt.rwc)
			if (err != nil) != tt.wantErr {
				t.Errorf("poolsFile.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
