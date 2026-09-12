package safety

import (
	"context"
	"testing"
)

func TestSafetyRail_BlocksRmRf(t *testing.T) {
	r := New()
	d, err := r.OnBeforeAct(context.Background(), "s", "exec_command", []byte(`{"command":"rm -rf /"}`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Allow {
		t.Fatal("expected rm -rf to be blocked")
	}
}

func TestSafetyRail_AllowsGoTest(t *testing.T) {
	r := New()
	d, err := r.OnBeforeAct(context.Background(), "s", "exec_command", []byte(`{"command":"go test ./..."}`))
	if err != nil {
		t.Fatal(err)
	}
	if !d.Allow {
		t.Fatalf("go test should pass, got %+v", d)
	}
}

func TestSafetyRail_BlocksTraversal(t *testing.T) {
	r := New()
	d, err := r.OnBeforeAct(context.Background(), "s", "write_file", []byte(`{"rel_path":"../../etc/passwd","content":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	if d.Allow {
		t.Fatal("expected path traversal block")
	}
}
