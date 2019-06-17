package xorm

import (
	"reflect"
	"testing"

	"xorm.io/core"
	"github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/assert"
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

func TestGetIndexColName(t *testing.T) {
	t.Run("Index", func(t *testing.T) {
		s := "CREATE INDEX test2_mm_idx ON test2 (major);"
		colNames := getIndexColName(s)
		assert.Equal(t, []string{"major"}, colNames)
	})

	t.Run("Multicolumn indexes", func(t *testing.T) {
		s := "CREATE INDEX test2_mm_idx ON test2 (major, minor);"
		colNames := getIndexColName(s)
		assert.Equal(t, []string{"major", "minor"}, colNames)
	})

	t.Run("Indexes and ORDER BY", func(t *testing.T) {
		s := "CREATE INDEX test2_mm_idx ON test2 (major  NULLS FIRST, minor DESC NULLS LAST);"
		colNames := getIndexColName(s)
		assert.Equal(t, []string{"major", "minor"}, colNames)
	})

	t.Run("Combining Multiple Indexes", func(t *testing.T) {
		s := "CREATE INDEX test2_mm_cm_idx ON public.test2 USING btree (major, minor) WHERE ((major <> 5) AND (minor <> 6))"
		colNames := getIndexColName(s)
		assert.Equal(t, []string{"major", "minor"}, colNames)
	})

	t.Run("unique", func(t *testing.T) {
		s := "CREATE UNIQUE INDEX test2_mm_uidx ON test2 (major);"
		colNames := getIndexColName(s)
		assert.Equal(t, []string{"major"}, colNames)
	})

	t.Run("Indexes on Expressions", func(t *testing.T) {})
}
