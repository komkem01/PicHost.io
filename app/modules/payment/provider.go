package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type ProviderName string

const (
	ProviderManual ProviderName = "manual"
	ProviderOmise  ProviderName = "omise"
)

var (
	ErrInvalidSignature = errors.New("payment: invalid webhook signature")
	ErrSecretRequired   = errors.New("payment: webhook secret is required")
)

type PaymentProvider interface {
	Name() ProviderName
	VerifyWebhookSignature(headers map[string]string, body []byte, secret string) error
}

type BaseProvider struct {
	providerName ProviderName
}

func (b *BaseProvider) Name() ProviderName {
	return b.providerName
}

// ManualPaymentProvider handles manual bank slip transfer payments.
type ManualPaymentProvider struct {
	BaseProvider
}

func NewManualPaymentProvider() *ManualPaymentProvider {
	return &ManualPaymentProvider{
		BaseProvider: BaseProvider{providerName: ProviderManual},
	}
}

func (p *ManualPaymentProvider) VerifyWebhookSignature(headers map[string]string, body []byte, secret string) error {
	return VerifyHMACorTokenSignature(headers, body, secret)
}

// OmisePaymentProvider handles Omise payment gateway webhooks and charges.
type OmisePaymentProvider struct {
	BaseProvider
}

func NewOmisePaymentProvider() *OmisePaymentProvider {
	return &OmisePaymentProvider{
		BaseProvider: BaseProvider{providerName: ProviderOmise},
	}
}

func (p *OmisePaymentProvider) VerifyWebhookSignature(headers map[string]string, body []byte, secret string) error {
	return VerifyHMACorTokenSignature(headers, body, secret)
}

// VerifyHMACorTokenSignature verifies either HMAC-SHA256 signature or direct X-Payment-Webhook-Token header.
func VerifyHMACorTokenSignature(headers map[string]string, body []byte, secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ErrSecretRequired
	}

	// 1. Direct shared token check
	token := getHeaderIgnoreCase(headers, "X-Payment-Webhook-Token")
	if token != "" && hmac.Equal([]byte(token), []byte(secret)) {
		return nil
	}

	// 2. HMAC-SHA256 signature check (X-Payment-Webhook-Signature or X-Signature or X-Omise-Signature)
	sigHeader := getHeaderIgnoreCase(headers, "X-Payment-Webhook-Signature")
	if sigHeader == "" {
		sigHeader = getHeaderIgnoreCase(headers, "X-Signature")
	}
	if sigHeader == "" {
		sigHeader = getHeaderIgnoreCase(headers, "X-Omise-Signature")
	}

	if sigHeader == "" {
		return ErrInvalidSignature
	}

	// Clean up prefix if present (e.g. "sha256=")
	sigHeader = strings.TrimPrefix(sigHeader, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !strings.EqualFold(sigHeader, expectedMAC) {
		return ErrInvalidSignature
	}

	return nil
}

func getHeaderIgnoreCase(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
