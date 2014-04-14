package event

import (
	"testing"
)

func TestRoundTrip(t *testing.T) {
	e1, err := New()
	if err != nil {
		t.Fatal(err)
	}

	buf, err := e1.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	e2, err := Unmarshal(buf)
	if err != nil {
		t.Fatal(err)
	}

	if e1["uuid"] != e2["uuid"] {
		t.Fatal("uuids do not match: %s vs %s", e1["uuid"], e2["uuid"])
	}

	c1, err := e1.Created()
	if err != nil {
		t.Fatal(err)
	}

	c2, err := e2.Created()
	if err != nil {
		t.Fatal(err)
	}

	// comparing timestamps is not exact, we lose precision, c2 is truncated
	// so just check to the nearest ms
	d := c1.Sub(c2)
	if d / 1e6 != 0 {
		t.Fatalf("timestamps do not match: %s vs %s", c1, c2)
	}
}
