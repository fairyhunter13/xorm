package hashkey

import "testing"

func TestGet(t *testing.T) {
	type args struct {
		sqlStr string
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name: "Remove all whitespace inside the sql string",
			args: args{
				sqlStr: "SELECT * FROM test",
			},
			wantRes: "SELECT*FROMtest",
		},
		{
			name: "Remove all whitespace characters inside the sql string",
			args: args{
				sqlStr: "SELECT * FROM test WHERE id = 5",
			},
			wantRes: "SELECT*FROMtestWHEREid=5",
		},
		{
			name: "Don't remove any characters inside quotes",
			args: args{
				sqlStr: "SELECT * FROM `test you` WHERE id = 5 AND name = 'Hello Test'",
			},
			wantRes: "SELECT*FROMtest youWHEREid=5ANDname=Hello Test",
		},
		{
			name: "Multi quotes in sql string for (edge cases)",
			args: args{
				sqlStr: "SELECT * FROM ``test you`` WHERE id = 5 AND name = 'Hello Test'",
			},
			wantRes: "SELECT*FROMtestyouWHEREid=5ANDname=Hello Test",
		},
		{
			name: "Multi quotes in sql string for appending",
			args: args{
				sqlStr: "SELECT * FROM `test you` WHERE id = 5 AND name = 'Hello Test''Appending Purposes'",
			},
			wantRes: "SELECT*FROMtest youWHEREid=5ANDname=Hello TestAppending Purposes",
		},
		{
			name: "Remove all ignored characters, ex colon and brackets",
			args: args{
				sqlStr: "INSERT INTO test(field1,field2) VALUES ('hello','name') ",
			},
			wantRes: "INSERTINTOtestfield1field2VALUEShelloname",
		},
		{
			name: "Remove all ignored characters, ex end statement",
			args: args{
				sqlStr: "INSERT INTO test(field1,field2) VALUES ('hello','name');",
			},
			wantRes: "INSERTINTOtestfield1field2VALUEShelloname",
		},
		{
			name: "Remove all control characters",
			args: args{
				sqlStr: "INSERT INTO test(field1,field2) VALUES ('hello','name');\x1A",
			},
			wantRes: "INSERTINTOtestfield1field2VALUEShelloname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRes := Get(tt.args.sqlStr); gotRes != tt.wantRes {
				t.Errorf("Get() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
