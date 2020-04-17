package mysql

import "testing"

func TestGetType(t *testing.T) {
	type args struct {
		typeStr string
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name: "Mysql type int(11) unsigned",
			args: args{
				typeStr: "int(11) unsigned",
			},
			wantRes: "int unsigned",
		},
		{
			name: "Mysql type int unsigned",
			args: args{
				typeStr: "int unsigned",
			},
			wantRes: "int unsigned",
		},
		{
			name: "Mysql type BIGINT(10) UNSIGNED uppercase",
			args: args{
				typeStr: "BIGINT(10) UNSIGNED",
			},
			wantRes: "BIGINT UNSIGNED",
		},
		{
			name: "Mysql type BIGINT UNSIGNED uppercase",
			args: args{
				typeStr: "BIGINT UNSIGNED",
			},
			wantRes: "BIGINT UNSIGNED",
		},
		{
			name: "Mysql type SERIAL uppercase",
			args: args{
				typeStr: "SERIAL",
			},
			wantRes: "SERIAL",
		},
		{
			name: "Mysql type BIGSERIAL(21) uppercase",
			args: args{
				typeStr: "BIGSERIAL(21)",
			},
			wantRes: "BIGSERIAL",
		},
		{
			name: "Mysql type bigserial(21) unsigned",
			args: args{
				typeStr: "bigserial(21) unsigned",
			},
			wantRes: "bigserial unsigned",
		},
		{
			name: "Mysql type bigserial(21) unsigned zerofill",
			args: args{
				typeStr: "bigserial(21) unsigned zerofill",
			},
			wantRes: "bigserial unsigned",
		},
		{
			name: "Mysql type bigserial(21) unsigned zerofill uppercase",
			args: args{
				typeStr: "BIGSERIAL(21) UNSIGNED ZEROFILL",
			},
			wantRes: "BIGSERIAL UNSIGNED",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := GetType(tt.args.typeStr); gotRes != tt.wantRes {
				t.Errorf("GetType() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
