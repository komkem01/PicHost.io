package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"github.com/google/uuid"
)

// Mock entities for unit testing Payment Service
type mockUserEnt struct {
	entitiesinf.UserEntity
	users map[uuid.UUID]*ent.UserEntity
}

func newMockUserEnt() *mockUserEnt {
	return &mockUserEnt{users: make(map[uuid.UUID]*ent.UserEntity)}
}

func (m *mockUserEnt) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (m *mockUserEnt) UpdateUserPlan(ctx context.Context, id uuid.UUID, plan entitiesdto.UpdateUserPlan) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	if plan.Plan != nil {
		u.Plan = ent.PlanType(*plan.Plan)
	}
	return u, nil
}

func (m *mockUserEnt) SetUserPlanExpiry(ctx context.Context, id uuid.UUID, expiresAt *time.Time, clearCancellation bool) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	u.PlanExpiresAt = expiresAt
	if clearCancellation {
		u.PlanCancelledAt = nil
	}
	return u, nil
}

type mockPlanEnt struct {
	entitiesinf.PlanSettingEntity
	plans map[string]*ent.PlanSettingEntity
}

func newMockPlanEnt() *mockPlanEnt {
	return &mockPlanEnt{plans: make(map[string]*ent.PlanSettingEntity)}
}

func (m *mockPlanEnt) GetPlanSettingByKey(ctx context.Context, key string) (*ent.PlanSettingEntity, error) {
	p, ok := m.plans[key]
	if !ok {
		return nil, fmt.Errorf("plan not found")
	}
	return p, nil
}

type mockPaymentEnt struct {
	entitiesinf.PaymentTransactionEntity
	txs map[uuid.UUID]*ent.PaymentTransactionEntity
}

func newMockPaymentEnt() *mockPaymentEnt {
	return &mockPaymentEnt{txs: make(map[uuid.UUID]*ent.PaymentTransactionEntity)}
}

func (m *mockPaymentEnt) CreatePaymentTransaction(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
	id := uuid.New()
	tx := &ent.PaymentTransactionEntity{
		ID:                id,
		UserID:            in.UserID,
		PlanKey:           in.PlanKey,
		AmountTHB:         in.AmountTHB,
		Currency:          in.Currency,
		Status:            in.Status,
		Provider:          in.Provider,
		CheckoutReference: in.CheckoutReference,
		PaymentURL:        in.PaymentURL,
		ExpiresAt:         in.ExpiresAt,
		Metadata:          in.Metadata,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	m.txs[id] = tx
	return tx, nil
}

func (m *mockPaymentEnt) GetPaymentTransactionByID(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	tx, ok := m.txs[id]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	return tx, nil
}

func (m *mockPaymentEnt) GetPaymentTransactionByCheckoutReference(ctx context.Context, checkoutReference string) (*ent.PaymentTransactionEntity, error) {
	for _, tx := range m.txs {
		if tx.CheckoutReference == checkoutReference {
			return tx, nil
		}
	}
	return nil, fmt.Errorf("transaction not found")
}

func (m *mockPaymentEnt) UpdatePaymentTransactionStatus(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentTransactionStatus) (*ent.PaymentTransactionEntity, error) {
	tx, ok := m.txs[id]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	tx.Status = in.Status
	if in.ProviderReference != nil {
		tx.ProviderReference = in.ProviderReference
	}
	if in.PaidAt != nil {
		tx.PaidAt = in.PaidAt
	}
	if in.ReviewReason != nil {
		tx.ReviewReason = in.ReviewReason
	}
	if in.ReviewedBy != nil {
		tx.ReviewedBy = in.ReviewedBy
	}
	if in.ReviewedAt != nil {
		tx.ReviewedAt = in.ReviewedAt
	}
	tx.UpdatedAt = time.Now()
	return tx, nil
}

func setupTestService() (*Service, *mockUserEnt, *mockPlanEnt, *mockPaymentEnt) {
	userEnt := newMockUserEnt()
	planEnt := newMockPlanEnt()
	paymentEnt := newMockPaymentEnt()

	conf := &config.Config[Config]{
		Val: &Config{
			WebhookSecret:      "test-secret-12345",
			CheckoutBaseURL:    "https://pichost.io",
			CheckoutTTLMinutes: 15,
			SubscriptionDays:   30,
		},
	}

	svc := newService(&Options{
		Config:     conf,
		userEnt:    userEnt,
		planEnt:    planEnt,
		paymentEnt: paymentEnt,
	})

	return svc, userEnt, planEnt, paymentEnt
}

// 1. Webhook Signature Verification Test
func TestWebhookSignatureVerification(t *testing.T) {
	secret := "my-secret-key"

	t.Run("empty secret returns ErrSecretRequired", func(t *testing.T) {
		err := VerifyHMACorTokenSignature(map[string]string{"X-Payment-Webhook-Token": "secret"}, []byte("{}"), "")
		if err != ErrSecretRequired {
			t.Errorf("expected ErrSecretRequired, got %v", err)
		}
	})

	t.Run("valid X-Payment-Webhook-Token header passes", func(t *testing.T) {
		headers := map[string]string{"X-Payment-Webhook-Token": secret}
		err := VerifyHMACorTokenSignature(headers, []byte("{}"), secret)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("valid HMAC-SHA256 signature passes", func(t *testing.T) {
		body := []byte(`{"status":"paid","checkout_reference":"REF123"}`)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		headers := map[string]string{"X-Signature": expectedSig}
		err := VerifyHMACorTokenSignature(headers, body, secret)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid HMAC-SHA256 signature fails", func(t *testing.T) {
		headers := map[string]string{"X-Signature": "invalid-signature"}
		err := VerifyHMACorTokenSignature(headers, []byte("{}"), secret)
		if err != ErrInvalidSignature {
			t.Errorf("expected ErrInvalidSignature, got %v", err)
		}
	})
}

// 2. Webhook Idempotency 5x Call Test (Exit Criteria: ยิง webhook ซ้ำ 5 ครั้ง → แพ็กเกจอัปเกรดครั้งเดียว)
func TestWebhookIdempotency_5TimesCall(t *testing.T) {
	svc, userEnt, planEnt, _ := setupTestService()
	ctx := context.Background()

	userID := uuid.New()
	userVerifiedAt := time.Now()
	userEnt.users[userID] = &ent.UserEntity{
		ID:              userID,
		Plan:            ent.PlanTypeFree,
		EmailVerifiedAt: &userVerifiedAt,
	}

	planEnt.plans["pro"] = &ent.PlanSettingEntity{
		PlanKey:        "pro",
		DisplayName:    "Pro Plan",
		MonthlyPriceTHB: 190,
		IsEnabled:      true,
	}

	// Step A: Create checkout
	tx, err := svc.CreateCheckout(ctx, CreateCheckoutInput{
		UserID:   userID,
		PlanKey:  "pro",
		Provider: "omise",
	})
	if err != nil {
		t.Fatalf("CreateCheckout failed: %v", err)
	}

	paidAmount := 190
	confirmInput := ConfirmPaymentInput{
		PaymentID:     &tx.ID,
		Status:        ent.PaymentStatusPaid,
		PaidAmountTHB: &paidAmount,
	}

	// Step B: Call 1 — initial payment confirmation
	updatedTx, upgraded, err := svc.ConfirmPayment(ctx, confirmInput)
	if err != nil {
		t.Fatalf("ConfirmPayment call 1 failed: %v", err)
	}
	if !upgraded {
		t.Errorf("expected upgraded == true on 1st call")
	}
	if updatedTx.Status != ent.PaymentStatusPaid {
		t.Errorf("expected status paid, got %s", updatedTx.Status)
	}

	// Check plan expiration set
	updatedUser, _ := userEnt.GetUserByID(ctx, userID)
	if updatedUser.Plan != ent.PlanTypePro {
		t.Errorf("expected user plan Pro, got %s", updatedUser.Plan)
	}
	initialExpiry := *updatedUser.PlanExpiresAt

	// Step C: Calls 2 to 5 — duplicate webhooks (retries)
	for i := 2; i <= 5; i++ {
		repeatTx, upgradedRepeat, errRepeat := svc.ConfirmPayment(ctx, confirmInput)
		if errRepeat != nil {
			t.Fatalf("ConfirmPayment call %d failed: %v", i, errRepeat)
		}
		if upgradedRepeat {
			t.Errorf("call %d: expected upgraded == false on repeat webhook, got true", i)
		}
		if repeatTx.Status != ent.PaymentStatusPaid {
			t.Errorf("call %d: expected status paid, got %s", i, repeatTx.Status)
		}
	}

	// Verify plan expiry wasn't extended 5x (must be equal to initial expiry)
	finalUser, _ := userEnt.GetUserByID(ctx, userID)
	if !finalUser.PlanExpiresAt.Equal(initialExpiry) {
		t.Errorf("expected plan expiration to be extended only once (%v), got %v", initialExpiry, finalUser.PlanExpiresAt)
	}
}

// 3. Refund Flow Test
func TestRefundPaymentFlow(t *testing.T) {
	svc, userEnt, planEnt, _ := setupTestService()
	ctx := context.Background()

	userID := uuid.New()
	userVerifiedAt := time.Now()
	userEnt.users[userID] = &ent.UserEntity{
		ID:              userID,
		Plan:            ent.PlanTypeFree,
		EmailVerifiedAt: &userVerifiedAt,
	}
	planEnt.plans["pro"] = &ent.PlanSettingEntity{
		PlanKey:        "pro",
		MonthlyPriceTHB: 190,
		IsEnabled:      true,
	}

	tx, _ := svc.CreateCheckout(ctx, CreateCheckoutInput{
		UserID:  userID,
		PlanKey: "pro",
	})

	t.Run("cannot refund pending transaction", func(t *testing.T) {
		_, err := svc.RefundPayment(ctx, RefundPaymentInput{PaymentID: tx.ID})
		if err != ErrPaymentNotPaidForRefund {
			t.Errorf("expected ErrPaymentNotPaidForRefund, got %v", err)
		}
	})

	// Pay the transaction first
	paidAmount := 190
	_, _, _ = svc.ConfirmPayment(ctx, ConfirmPaymentInput{
		PaymentID:     &tx.ID,
		Status:        ent.PaymentStatusPaid,
		PaidAmountTHB: &paidAmount,
	})

	t.Run("refund paid transaction succeeds", func(t *testing.T) {
		reason := "User requested refund"
		refunded, err := svc.RefundPayment(ctx, RefundPaymentInput{
			PaymentID: tx.ID,
			Reason:    &reason,
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if refunded.Status != ent.PaymentStatusRefunded {
			t.Errorf("expected status refunded, got %s", refunded.Status)
		}
	})

	t.Run("refunding already refunded transaction fails", func(t *testing.T) {
		_, err := svc.RefundPayment(ctx, RefundPaymentInput{PaymentID: tx.ID})
		if err != ErrPaymentAlreadyRefunded {
			t.Errorf("expected ErrPaymentAlreadyRefunded, got %v", err)
		}
	})
}

// 4. Provider Selection & Fallback Test
func TestProviderSelection(t *testing.T) {
	svc, userEnt, planEnt, _ := setupTestService()
	ctx := context.Background()

	userID := uuid.New()
	userVerifiedAt := time.Now()
	userEnt.users[userID] = &ent.UserEntity{
		ID:              userID,
		EmailVerifiedAt: &userVerifiedAt,
	}
	planEnt.plans["basic"] = &ent.PlanSettingEntity{
		PlanKey:        "basic",
		MonthlyPriceTHB: 90,
		IsEnabled:      true,
	}

	t.Run("default provider is manual", func(t *testing.T) {
		tx, err := svc.CreateCheckout(ctx, CreateCheckoutInput{
			UserID:  userID,
			PlanKey: "basic",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.Provider != "manual" {
			t.Errorf("expected provider manual, got %s", tx.Provider)
		}
	})

	t.Run("specified provider is saved", func(t *testing.T) {
		tx, err := svc.CreateCheckout(ctx, CreateCheckoutInput{
			UserID:   userID,
			PlanKey:  "basic",
			Provider: "omise",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.Provider != "omise" {
			t.Errorf("expected provider omise, got %s", tx.Provider)
		}
	})
}
