# Playbook: write a test

*When to read this:* before writing any test, or when someone says "add tests".
*Result you should have:* one table-driven test, fixture if needed, `make check` green.

## Checklist

1. Name: `Test<Unit>_<Behaviour>` in the same package (white-box) unless the
   public API boundary is the point — then `_test` package (black-box).
2. Table-driven ONLY. Shape:

```go
func TestNormalizeFingerprint(t *testing.T) {
    tests := []struct {
        name string
        in   Job
        want string
    }{
        {"trims spaces", Job{Company: " ACME ", Title: " Dev "}, "acme|dev|"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Fingerprint(tt.in)
            if got != tt.want {
                t.Errorf("Fingerprint(%+v) = %q, want %q", tt.in, got, tt.want)
            }
        })
    }
}
```

3. HTTP/HTML/JSON behaviour: save a real minimal fixture under `testdata/<topic>/`
   and feed it. Never hand-write fake HTML inline in the test file.
4. Server tests use `httptest.NewServer` on 127.0.0.1. No external network in tests.
5. Error paths are tested too: assert the error is non-nil AND its message
   contains the context you wrapped (`fetch <url>:`).
6. Run `go test ./internal/<pkg>/ -run <TestName> -v` first, then `make check`.
7. Commit: `test(<pkg>): cover <behaviour>`.

## If this fails

- Test is flaky (passes/fails on rerun): root-cause before proceeding — time,
  randomness, port reuse, and map ordering are usual causes. Never `time.Sleep`
  as a fix; use channels/polling with deadline.
- You are tempted to skip a test: do not; write the reason, then the fix.
