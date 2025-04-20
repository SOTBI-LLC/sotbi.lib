package templates

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sender_execTemplate(t *testing.T) {
	type args struct {
		data  any
		templ string
	}

	tests := []struct {
		name    string
		args    args
		wantRes string
		wantErr bool
	}{
		{
			name: "Test 1",
			args: args{
				data:  accountStatementDataNew,
				templ: templateAccountStatementNew,
			},
			wantRes: result,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := ExecTemplate(tt.args.data, tt.args.templ, "192.168.80.30")
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Contains(t, gotRes, tt.wantRes)
		})
	}
}
