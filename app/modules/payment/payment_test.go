package payment

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"github.com/google/uuid"
)

// --- Fakes ---------------------------------------------------------------
//
// Each fake embeds the corresponding (nil) interface so it satisfies the full
// interface without having to stub every method — only the methods a given
// test actually exercises are overridden. Calling an un-stubbed method will
// panic on the embedded nil interface, which is exactly what we want: it
// surfaces as a clear test failure if a test starts depending on behavior it
// didn't intend to.

type fakeUserEnt struct {
	entitiesinf.UserEntity
	getUserByIDFn func(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error)
}

func (f *fakeUserEnt) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	if f.getUserByIDFn != nil {
		return f.getUserByIDFn(ctx, id)
	}
	return &ent.UserEntity{ID: id}, nil
}

type fakePlanEnt struct {
	entitiesinf.PlanSettingEntity
	getPlanSettingByKeyFn func(ctx context.Context, key string) (*ent.PlanSettingEntity, error)
}

func (f *fakePlanEnt) GetPlanSettingByKey(ctx context.Context, key string) (*ent.PlanSettingEntity, error) {
	if f.getPlanSettingByKeyFn != nil {
		return f.getPlanSettingByKeyFn(ctx, key)
	}
	return &ent.PlanSettingEntity{PlanKey: key, IsEnabled: true, MonthlyPriceTHB: 100, DisplayName: key}, nil
}

type fakePaymentEnt struct {
	entitiesinf.PaymentTransactionEntity

	getOpenFn                 func(ctx context.Context, userID uuid.UUID) (*ent.PaymentTransactionEntity, error)
	updateStatusFn            func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentTransactionStatus) (*ent.PaymentTransactionEntity, error)
	createPaymentTxFn         func(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error)
	createPaymentTxIfNoOpenFn func(ctx context.Context, userID uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error)
	getByIDFn                 func(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error)
	confirmPaymentTxFn        func(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error)
	activatePendingFn         func(ctx context.Context, userID uuid.UUID, dur time.Duration) (*entitiesdto.ActivateEntitlementsResult, error)
}

func (f *fakePaymentEnt) GetOpenPaymentTransactionByUserID(ctx context.Context, userID uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	if f.getOpenFn != nil {
		return f.getOpenFn(ctx, userID)
	}
	return nil, sql.ErrNoRows
}

func (f *fakePaymentEnt) UpdatePaymentTransactionStatus(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentTransactionStatus) (*ent.PaymentTransactionEntity, error) {
	if f.updateStatusFn != nil {
		return f.updateStatusFn(ctx, id, in)
	}
	return &ent.PaymentTransactionEntity{ID: id, Status: in.Status}, nil
}

func (f *fakePaymentEnt) CreatePaymentTransaction(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
	if f.createPaymentTxFn != nil {
		return f.createPaymentTxFn(ctx, in)
	}
	return &ent.PaymentTransactionEntity{ID: uuid.New(), UserID: in.UserID, PlanKey: in.PlanKey, Status: in.Status}, nil
}

// CreatePaymentTransactionIfNoOpen simulates the entities-layer atomic
// check-then-insert: by default (no override) it behaves as if there were no
// open transaction and simply inserts. Tests that need to exercise the
// conflict/lazy-expiry paths stub createPaymentTxIfNoOpenFn directly, mirroring
// what the real transactional implementation would decide.
func (f *fakePaymentEnt) CreatePaymentTransactionIfNoOpen(ctx context.Context, userID uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
	if f.createPaymentTxIfNoOpenFn != nil {
		return f.createPaymentTxIfNoOpenFn(ctx, userID, in)
	}
	return &ent.PaymentTransactionEntity{ID: uuid.New(), UserID: in.UserID, PlanKey: in.PlanKey, Status: in.Status}, nil
}

func (f *fakePaymentEnt) GetPaymentTransactionByID(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return nil, sql.ErrNoRows
}

func (f *fakePaymentEnt) ConfirmPaymentTransaction(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error) {
	if f.confirmPaymentTxFn != nil {
		return f.confirmPaymentTxFn(ctx, in)
	}
	return nil, errors.New("confirmPaymentTxFn not stubbed")
}

func (f *fakePaymentEnt) ActivatePendingEntitlements(ctx context.Context, userID uuid.UUID, dur time.Duration) (*entitiesdto.ActivateEntitlementsResult, error) {
	if f.activatePendingFn != nil {
		return f.activatePendingFn(ctx, userID, dur)
	}
	return &entitiesdto.ActivateEntitlementsResult{}, nil
}

func newTestService(userEnt entitiesinf.UserEntity, planEnt entitiesinf.PlanSettingEntity, paymentEnt entitiesinf.PaymentTransactionEntity) *Service {
	cfg := &config.Config[Config]{Val: &Config{
		CheckoutTTLMinutes: 15,
		SubscriptionDays:   30,
	}}
	return newService(&Options{
		Config:     cfg,
		userEnt:    userEnt,
		planEnt:    planEnt,
		paymentEnt: paymentEnt,
	})
}

// --- CreateCheckout: open-checkout conflict (atomic entities-layer op) ----
//
// CreateCheckout no longer does its own "check, then create" against the
// entities layer — that race is now closed by delegating both steps to a
// single entities.CreatePaymentTransactionIfNoOpen transaction (user row
// locked FOR UPDATE, then the open-row check/expiry, then the insert, all in
// one DB transaction). These tests only verify that payment.Service correctly
// forwards the call and maps entitiesdto.ErrOpenPaymentTransactionExists to
// *ErrPaymentOpenExists; the atomicity itself is entities-layer behavior that
// requires a real Postgres transaction to observe and isn't covered here.

func TestCreateCheckout_OpenExistsConflict_ReturnsExistingPaymentID(t *testing.T) {
	userID := uuid.New()
	openID := uuid.New()

	paymentEnt := &fakePaymentEnt{
		createPaymentTxIfNoOpenFn: func(ctx context.Context, uid uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
			if uid != userID {
				t.Fatalf("expected userID %s to be forwarded, got %s", userID, uid)
			}
			return nil, &entitiesdto.ErrOpenPaymentTransactionExists{PaymentID: openID}
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	_, err := svc.CreateCheckout(context.Background(), CreateCheckoutInput{
		UserID:  userID,
		PlanKey: "basic",
	})

	var openErr *ErrPaymentOpenExists
	if !errors.As(err, &openErr) {
		t.Fatalf("expected *ErrPaymentOpenExists, got %v", err)
	}
	if openErr.PaymentID != openID {
		t.Fatalf("expected conflict to carry payment ID %s, got %s", openID, openErr.PaymentID)
	}
}

func TestCreateCheckout_NoOpenTransaction_CreatesNewCheckout(t *testing.T) {
	userID := uuid.New()

	var gotUserID uuid.UUID
	var gotIn entitiesdto.CreatePaymentTransaction
	paymentEnt := &fakePaymentEnt{
		createPaymentTxIfNoOpenFn: func(ctx context.Context, uid uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
			gotUserID = uid
			gotIn = in
			return &ent.PaymentTransactionEntity{ID: uuid.New(), UserID: in.UserID, PlanKey: in.PlanKey, Status: in.Status}, nil
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	row, err := svc.CreateCheckout(context.Background(), CreateCheckoutInput{
		UserID:  userID,
		PlanKey: "basic",
	})
	if err != nil {
		t.Fatalf("expected new checkout to succeed, got %v", err)
	}
	if row == nil {
		t.Fatalf("expected a new payment transaction to be created")
	}
	if gotUserID != userID {
		t.Fatalf("expected CreatePaymentTransactionIfNoOpen to be called with userID %s, got %s", userID, gotUserID)
	}
	if gotIn.UserID != userID || gotIn.PlanKey != "basic" || gotIn.Status != ent.PaymentStatusPending {
		t.Fatalf("unexpected CreatePaymentTransaction input: %+v", gotIn)
	}
}

func TestCreateCheckout_OtherEntitiesError_IsPropagatedUnwrapped(t *testing.T) {
	userID := uuid.New()
	boom := errors.New("boom: db unavailable")

	paymentEnt := &fakePaymentEnt{
		createPaymentTxIfNoOpenFn: func(ctx context.Context, uid uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
			return nil, boom
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	_, err := svc.CreateCheckout(context.Background(), CreateCheckoutInput{
		UserID:  userID,
		PlanKey: "basic",
	})
	var openErr *ErrPaymentOpenExists
	if errors.As(err, &openErr) {
		t.Fatalf("did not expect a conflict error for an unrelated failure")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected the underlying error to be propagated, got %v", err)
	}
}

// --- ConfirmPayment: activation branching ---------------------------------

func TestConfirmPayment_PaidAndVerified_ActivatesImmediately(t *testing.T) {
	userID := uuid.New()
	paymentID := uuid.New()
	email := "verified@example.com"

	paymentEnt := &fakePaymentEnt{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
			return &ent.PaymentTransactionEntity{
				ID:        id,
				UserID:    userID,
				Status:    ent.PaymentStatusPending,
				PlanKey:   "basic",
				AmountTHB: 100,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
		confirmPaymentTxFn: func(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error) {
			if in.NextStatus != ent.PaymentStatusPaid {
				t.Fatalf("expected NextStatus paid, got %s", in.NextStatus)
			}
			return &entitiesdto.ConfirmPaymentTransactionResult{
				Payment:   &ent.PaymentTransactionEntity{ID: in.PaymentID, UserID: userID, Status: ent.PaymentStatusPaid, PlanKey: "basic", AmountTHB: 100},
				User:      &ent.UserEntity{ID: userID, Email: &email},
				Activated: true,
			}, nil
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	amount := 100
	updated, activated, err := svc.ConfirmPayment(context.Background(), ConfirmPaymentInput{
		PaymentID:     &paymentID,
		Status:        ent.PaymentStatusPaid,
		PaidAmountTHB: &amount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !activated {
		t.Fatalf("expected activated=true for a verified user's paid transaction")
	}
	if updated.Status != ent.PaymentStatusPaid {
		t.Fatalf("expected status paid, got %s", updated.Status)
	}
}

func TestConfirmPayment_PaidButUnverified_WaitsForActivation(t *testing.T) {
	userID := uuid.New()
	paymentID := uuid.New()
	email := "unverified@example.com"

	paymentEnt := &fakePaymentEnt{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
			return &ent.PaymentTransactionEntity{
				ID:        id,
				UserID:    userID,
				Status:    ent.PaymentStatusPending,
				PlanKey:   "basic",
				AmountTHB: 100,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
		confirmPaymentTxFn: func(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error) {
			return &entitiesdto.ConfirmPaymentTransactionResult{
				Payment:   &ent.PaymentTransactionEntity{ID: in.PaymentID, UserID: userID, Status: ent.PaymentStatusPaid, PlanKey: "basic", AmountTHB: 100},
				User:      &ent.UserEntity{ID: userID, Email: &email}, // EmailVerifiedAt is nil
				Activated: false,
			}, nil
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	amount := 100
	updated, activated, err := svc.ConfirmPayment(context.Background(), ConfirmPaymentInput{
		PaymentID:     &paymentID,
		Status:        ent.PaymentStatusPaid,
		PaidAmountTHB: &amount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activated {
		t.Fatalf("expected activated=false while the user's email is unverified")
	}
	if updated.Status != ent.PaymentStatusPaid {
		t.Fatalf("expected status paid (approved), got %s", updated.Status)
	}
	if !isAwaitingVerification(updated) {
		t.Fatalf("expected the payment to be reported as awaiting verification")
	}
}

func TestConfirmPayment_AlreadyTerminal_IsIdempotentNoOp(t *testing.T) {
	userID := uuid.New()
	paymentID := uuid.New()

	paymentEnt := &fakePaymentEnt{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
			return &ent.PaymentTransactionEntity{
				ID:        id,
				UserID:    userID,
				Status:    ent.PaymentStatusPending,
				PlanKey:   "basic",
				AmountTHB: 100,
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}, nil
		},
		confirmPaymentTxFn: func(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error) {
			// Simulate a concurrent confirmation that already transitioned the row
			// to "failed" between our pre-check and the locked transaction.
			return &entitiesdto.ConfirmPaymentTransactionResult{
				Payment:         &ent.PaymentTransactionEntity{ID: in.PaymentID, UserID: userID, Status: ent.PaymentStatusFailed, PlanKey: "basic", AmountTHB: 100},
				AlreadyTerminal: true,
			}, nil
		},
	}

	svc := newTestService(&fakeUserEnt{}, &fakePlanEnt{}, paymentEnt)

	updated, activated, err := svc.ConfirmPayment(context.Background(), ConfirmPaymentInput{
		PaymentID: &paymentID,
		Status:    ent.PaymentStatusPaid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activated {
		t.Fatalf("expected activated=false for an already-terminal row")
	}
	if updated.Status != ent.PaymentStatusFailed {
		t.Fatalf("expected the pre-existing terminal status to be preserved, got %s", updated.Status)
	}
}

func TestIsAwaitingVerification(t *testing.T) {
	if isAwaitingVerification(nil) {
		t.Fatalf("nil row should not be awaiting verification")
	}
	if isAwaitingVerification(&ent.PaymentTransactionEntity{Status: ent.PaymentStatusPending}) {
		t.Fatalf("pending row should not be awaiting verification")
	}
	now := time.Now()
	if isAwaitingVerification(&ent.PaymentTransactionEntity{Status: ent.PaymentStatusPaid, ActivatedAt: &now}) {
		t.Fatalf("activated row should not be awaiting verification")
	}
	if !isAwaitingVerification(&ent.PaymentTransactionEntity{Status: ent.PaymentStatusPaid, ActivatedAt: nil}) {
		t.Fatalf("paid-but-unactivated row should be awaiting verification")
	}
}
