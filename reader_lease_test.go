package expect

import (
	"context"
	"io"
	"sync"
	"testing"
)

func TestReaderLease(t *testing.T) {
	in, out := io.Pipe()
	defer out.Close()
	defer in.Close()

	rm := NewReaderLease(in)

	tests := []struct {
		title    string
		expected string
	}{
		{
			"Read cancels with deadline",
			"apple",
		},
		{
			"Second read has no bytes stolen",
			"banana",
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			tin, tout := io.Pipe()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				io.Copy(tout, rm.NewReader(ctx))
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := out.Write([]byte(test.expected))
				checkErr(t, "Write", err, nil)
			}()

			for i := 0; i < len(test.expected); i++ {
				p := make([]byte, 1)
				n, err := tin.Read(p)
				checkErr(t, "Read", err, nil)
				if g, w := n, 1; g != w {
					t.Fatalf("Read gave n==%d, want n==%d", g, w)
				}
				if g, w := p[0], test.expected[i]; g != w {
					t.Errorf("read %d gave %q, want %q", i, g, w)
				}
			}

			cancel()
			wg.Wait()
		})
	}
}
