package validator

import (
	"testing"
)

func Test_keySetFromKeys(t *testing.T) {
	type args struct {
		keys []Key
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if returns an error on nil keys",
			args: args{
				keys: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := keySetFromKeys(tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("keySetFromKeys() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
		})
	}
}
