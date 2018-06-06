package xorm

import (
	"reflect"
	"testing"

	"github.com/go-xorm/core"
	"github.com/jackc/pgx/stdlib"
)

func TestParsePostgres(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"***REMOVED***://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		{"***REMOVED***ql://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		{"postg://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", false},
		//{"***REMOVED***://auser:pass with space@localhost:5432/db?sslmode=disable", "db", true},
		//{"***REMOVED***:// auser : password@localhost:5432/db?sslmode=disable", "db", true},
		{"***REMOVED***://%20auser%20:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		//{"***REMOVED***://auser:***REMOVED***@localhost:5432/データベース?sslmode=disable", "データベース", true},
		{"dbname=db sslmode=disable", "db", true},
		{"user=auser password=password dbname=db sslmode=disable", "db", true},
		{"", "db", false},
		{"dbname=db =disable", "db", false},
	}

	driver := core.QueryDriver("***REMOVED***")

	for _, test := range tests {
		uri, err := driver.Parse("***REMOVED***", test.in)

		if err != nil && test.valid {
			t.Errorf("%q got unexpected error: %s", test.in, err)
		} else if err == nil && !reflect.DeepEqual(test.expected, uri.DbName) {
			t.Errorf("%q got: %#v want: %#v", test.in, uri.DbName, test.expected)
		}
	}
}

func TestParsePgx(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		valid    bool
	}{
		{"***REMOVED***://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		{"***REMOVED***ql://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		{"postg://auser:***REMOVED***@localhost:5432/db?sslmode=disable", "db", false},
		//{"***REMOVED***://auser:pass with space@localhost:5432/db?sslmode=disable", "db", true},
		//{"***REMOVED***:// auser : password@localhost:5432/db?sslmode=disable", "db", true},
		{"***REMOVED***://%20auser%20:***REMOVED***@localhost:5432/db?sslmode=disable", "db", true},
		//{"***REMOVED***://auser:***REMOVED***@localhost:5432/データベース?sslmode=disable", "データベース", true},
		{"dbname=db sslmode=disable", "db", true},
		{"user=auser password=password dbname=db sslmode=disable", "db", true},
		{"", "db", false},
		{"dbname=db =disable", "db", false},
	}

	driver := core.QueryDriver("pgx")

	for _, test := range tests {
		uri, err := driver.Parse("pgx", test.in)

		if err != nil && test.valid {
			t.Errorf("%q got unexpected error: %s", test.in, err)
		} else if err == nil && !reflect.DeepEqual(test.expected, uri.DbName) {
			t.Errorf("%q got: %#v want: %#v", test.in, uri.DbName, test.expected)
		}

		// Register DriverConfig
		drvierConfig := stdlib.DriverConfig{}
		stdlib.RegisterDriverConfig(&drvierConfig)
		uri, err = driver.Parse("pgx",
			drvierConfig.ConnectionString(test.in))
		if err != nil && test.valid {
			t.Errorf("%q got unexpected error: %s", test.in, err)
		} else if err == nil && !reflect.DeepEqual(test.expected, uri.DbName) {
			t.Errorf("%q got: %#v want: %#v", test.in, uri.DbName, test.expected)
		}

	}

}
