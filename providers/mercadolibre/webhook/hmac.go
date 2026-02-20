package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zentry/sdk-mercadolibre/core/errors"
)

const defaultTimestampTolerance = 5 * time.Minute

func parseSignatureHeader(header string) (ts, hash string, err error) {
	if header == "" {
		return "", "", errors.NewError(errors.ErrCodeInvalidWebhook, "missing x-signature header")
	}

	parts := strings.Split(header, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			ts = kv[1]
		case "v1":
			hash = kv[1]
		}
	}

	if ts == "" || hash == "" {
		return "", "", errors.NewError(errors.ErrCodeInvalidWebhook, "invalid x-signature format: missing ts or v1")
	}
	return ts, hash, nil
}

func buildManifest(dataID, requestID, ts string) string {
	return fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, requestID, ts)
}

func computeHMAC(manifest, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyHMAC(computed, expected string) bool {
	return hmac.Equal([]byte(computed), []byte(expected))
}

func validateTimestamp(ts string, tolerance time.Duration) error {
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.NewError(errors.ErrCodeInvalidWebhook, "invalid timestamp in signature")
	}

	age := time.Since(time.Unix(tsInt, 0))
	if age < 0 {
		age = -age
	}
	if age > tolerance {
		return errors.NewError(errors.ErrCodeInvalidWebhook, "webhook timestamp outside tolerance window")
	}
	return nil
}
