package pq

import (
	"testing"
)

func TestSimpleParseURL(t *testing.T) {
	expected := "host=hostname.remote"
	str, err := ParseURL("***REMOVED***://hostname.remote")
	if err != nil {
		t.Fatal(err)
	}

	if str != expected {
		t.Fatalf("unexpected result from ParseURL:\n+ %v\n- %v", str, expected)
	}
}

func TestIPv6LoopbackParseURL(t *testing.T) {
	expected := "host=::1 port=***REMOVED***4"
	str, err := ParseURL("***REMOVED***://[::1]:***REMOVED***4")
	if err != nil {
		t.Fatal(err)
	}

	if str != expected {
		t.Fatalf("unexpected result from ParseURL:\n+ %v\n- %v", str, expected)
	}
}

func TestFullParseURL(t *testing.T) {
	expected := `dbname=database host=hostname.remote password=top\ secret port=***REMOVED***4 user=username`
	str, err := ParseURL("***REMOVED***://username:***REMOVED***@hostname.remote:***REMOVED***4/database")
	if err != nil {
		t.Fatal(err)
	}

	if str != expected {
		t.Fatalf("unexpected result from ParseURL:\n+ %s\n- %s", str, expected)
	}
}

func TestInvalidProtocolParseURL(t *testing.T) {
	_, err := ParseURL("http://hostname.remote")
	switch err {
	case nil:
		t.Fatal("Expected an error from parsing invalid protocol")
	default:
		msg := "invalid connection protocol: http"
		if err.Error() != msg {
			t.Fatalf("Unexpected error message:\n+ %s\n- %s",
				err.Error(), msg)
		}
	}
}

func TestMinimalURL(t *testing.T) {
	cs, err := ParseURL("***REMOVED***://")
	if err != nil {
		t.Fatal(err)
	}

	if cs != "" {
		t.Fatalf("expected blank connection string, got: %q", cs)
	}
}
