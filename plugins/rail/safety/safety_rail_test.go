package safety

import (
	"context"
	"testing"
)

func TestSafetyRail_BlocksRmRf(t *testing.T) {
	r := New()

	tests := []struct {
		cmd   string
		block bool
	}{
		// Should block
		{`rm -rf /`, true},
		{`rm -fr /`, true},
		{`rm -r -f /`, true},
		{`rm -f -r /`, true},
		{`rm --recursive --force /`, true},
		
		// Should allow
		{`rm file.txt`, false},
		{`rm -i x`, false},
		{`go test ./...`, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			args := []byte(`{"command":"` + tt.cmd + `"}`)
			d, err := r.OnBeforeAct(context.Background(), "s", "exec_command", args)
			if err != nil {
				t.Fatal(err)
			}
			if d.Allow == tt.block {
				t.Errorf("cmd %q: expected block=%v, got block=%v", tt.cmd, tt.block, !d.Allow)
			}
		})
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
