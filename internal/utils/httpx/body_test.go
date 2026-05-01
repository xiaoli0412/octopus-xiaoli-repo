package httpx

import (
	"io"
	"strings"
	"testing"
)

func TestReadLimitedBodyRejectsInvalidLimit(t *testing.T) {
	_, err := ReadLimitedBody(strings.NewReader("ok"), 0, "")
	if err == nil || err.Error() != "invalid read limit" {
		t.Fatalf("ReadLimitedBody() error = %v, want invalid read limit", err)
	}
}

func TestReadLimitedBodyUsesDefaultTooLargeMessage(t *testing.T) {
	_, err := ReadLimitedBody(strings.NewReader("abcd"), 3, "")
	if err == nil || err.Error() != defaultTooLargeMessage {
		t.Fatalf("ReadLimitedBody() error = %v, want %q", err, defaultTooLargeMessage)
	}
}

func TestReadLimitedBodyUsesCustomTooLargeMessage(t *testing.T) {
	const want = "custom too large"
	_, err := ReadLimitedBody(strings.NewReader("abcd"), 3, want)
	if err == nil || err.Error() != want {
		t.Fatalf("ReadLimitedBody() error = %v, want %q", err, want)
	}
}

func TestReadLimitedBodyReturnsPayloadWithinLimit(t *testing.T) {
	got, err := ReadLimitedBody(strings.NewReader("abcd"), 4, "")
	if err != nil {
		t.Fatalf("ReadLimitedBody() error = %v", err)
	}
	if string(got) != "abcd" {
		t.Fatalf("ReadLimitedBody() = %q, want %q", string(got), "abcd")
	}
}

func TestReadLimitedBodyPropagatesReaderError(t *testing.T) {
	boom := io.ErrUnexpectedEOF
	_, err := ReadLimitedBody(errReader{err: boom}, 4, "")
	if err != boom {
		t.Fatalf("ReadLimitedBody() error = %v, want %v", err, boom)
	}
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
