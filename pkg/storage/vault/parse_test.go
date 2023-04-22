package vault

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	vault "github.com/hashicorp/vault/api"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault/testdata"
)

func Test_validateSecret(t *testing.T) {
	type args struct {
		secret *vault.KVSecret
	}
	tests := []struct {
		name    string
		args    args
		want    secretData
		wantErr bool
	}{
		{
			name: "Test if correctly parses valid PEM string and algorithm",
			args: args{
				secret: &vault.KVSecret{
					Data: map[string]interface{}{
						"algorithm": "RSA",
						"private":   testdata.RSAPem,
					},
				},
			},
			want: secretData{
				algorithm:  RSA,
				encodedKey: testdata.RSAPem,
			},
		},
		{
			name:    "Test if fails on nil secret",
			args:    args{secret: nil},
			wantErr: true,
		},
		{
			name: "Test if fails on missing 'private' field",
			args: args{
				secret: &vault.KVSecret{
					Data: map[string]interface{}{
						"algorithm": RSA,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSecret(tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSecret() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmp.AllowUnexported(secretData{})) {
				t.Errorf("parseSecret(): got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}
