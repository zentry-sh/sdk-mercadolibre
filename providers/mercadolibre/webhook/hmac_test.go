package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/domain"
	"github.com/zentry/sdk-mercadolibre/pkg/logger"
)

func TestParseSignatureHeader_Valid(t *testing.T) {
	ts, hash, err := parseSignatureHeader("ts=1709123456,v1=abcdef0123456789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != "1709123456" {
		t.Errorf("expected ts '1709123456', got '%s'", ts)
	}
	if hash != "abcdef0123456789" {
		t.Errorf("expected hash 'abcdef0123456789', got '%s'", hash)
	}
}

func TestParseSignatureHeader_WithSpaces(t *testing.T) {
	ts, hash, err := parseSignatureHeader("ts=123 , v1=abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts != "123" {
		t.Errorf("expected ts '123', got '%s'", ts)
	}
	if hash != "abc" {
		t.Errorf("expected hash 'abc', got '%s'", hash)
	}
}

func TestParseSignatureHeader_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{"empty", ""},
		{"no equals", "garbage"},
		{"missing v1", "ts=123"},
		{"missing ts", "v1=abc"},
		{"no values", "ts=,v1="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseSignatureHeader(tt.header)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestBuildManifest(t *testing.T) {
	result := buildManifest("12345", "req-abc", "1709123456")
	expected := "id:12345;request-id:req-abc;ts:1709123456;"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestBuildManifest_EmptyFields(t *testing.T) {
	result := buildManifest("", "", "")
	expected := "id:;request-id:;ts:;"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestComputeHMAC_KnownVector(t *testing.T) {
	// Compute expected value manually
	secret := "test-secret-key"
	manifest := "id:12345;request-id:req-001;ts:1709123456;"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	got := computeHMAC(manifest, secret)
	if got != expected {
		t.Errorf("expected '%s', got '%s'", expected, got)
	}
}

func TestComputeHMAC_DifferentSecrets(t *testing.T) {
	manifest := "id:123;request-id:req;ts:999;"
	hash1 := computeHMAC(manifest, "secret-a")
	hash2 := computeHMAC(manifest, "secret-b")

	if hash1 == hash2 {
		t.Error("different secrets should produce different HMACs")
	}
}

func TestComputeHMAC_DifferentManifests(t *testing.T) {
	secret := "same-secret"
	hash1 := computeHMAC("id:1;request-id:a;ts:100;", secret)
	hash2 := computeHMAC("id:2;request-id:b;ts:200;", secret)

	if hash1 == hash2 {
		t.Error("different manifests should produce different HMACs")
	}
}

func TestVerifyHMAC_Match(t *testing.T) {
	hash := computeHMAC("test-manifest", "test-secret")
	if !verifyHMAC(hash, hash) {
		t.Error("identical hashes should verify as equal")
	}
}

func TestVerifyHMAC_Mismatch(t *testing.T) {
	hash1 := computeHMAC("manifest-a", "secret")
	hash2 := computeHMAC("manifest-b", "secret")
	if verifyHMAC(hash1, hash2) {
		t.Error("different hashes should not verify as equal")
	}
}

func TestValidateTimestamp_Current(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	err := validateTimestamp(ts, 5*time.Minute)
	if err != nil {
		t.Fatalf("expected nil for current timestamp, got: %v", err)
	}
}

func TestValidateTimestamp_WithinTolerance(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Add(-2*time.Minute).Unix(), 10)
	err := validateTimestamp(ts, 5*time.Minute)
	if err != nil {
		t.Fatalf("expected nil for timestamp within tolerance, got: %v", err)
	}
}

func TestValidateTimestamp_OutsideTolerance(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10)
	err := validateTimestamp(ts, 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for timestamp outside tolerance, got nil")
	}
}

func TestValidateTimestamp_FutureOutsideTolerance(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	err := validateTimestamp(ts, 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for future timestamp outside tolerance, got nil")
	}
}

func TestValidateTimestamp_Invalid(t *testing.T) {
	err := validateTimestamp("not-a-number", 5*time.Minute)
	if err == nil {
		t.Fatal("expected error for non-numeric timestamp, got nil")
	}
}

func TestEndToEnd_ValidateWithComputedHMAC(t *testing.T) {
	secret := "e2e-test-secret"
	dataID := "pay-99999"
	requestID := "req-e2e-001"
	ts := strconv.FormatInt(time.Now().Unix(), 10)

	manifest := buildManifest(dataID, requestID, ts)
	hash := computeHMAC(manifest, secret)
	signature := fmt.Sprintf("ts=%s,v1=%s", ts, hash)

	h := NewHandler(logger.Nop())

	err := h.Validate(domain.WebhookRequest{
		Body:      []byte(`{"action":"payment.created","data":{"id":"pay-99999"}}`),
		Signature: signature,
		RequestID: requestID,
		DataID:    dataID,
	}, secret)

	if err != nil {
		t.Fatalf("end-to-end validation failed: %v", err)
	}
}
