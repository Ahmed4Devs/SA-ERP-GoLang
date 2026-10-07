package iam_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
	"erp/internal/iam/infrastructure/crypto"
	"erp/internal/iam/infrastructure/postgres"
	"erp/internal/iam/ports"
	"erp/internal/platform/transaction"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type pausedPasswordVerifier struct {
	ports.PasswordHasher
	verified chan struct{}
	resume   chan struct{}
}

func (h *pausedPasswordVerifier) Verify(password, hash string) (bool, error) {
	valid, err := h.PasswordHasher.Verify(password, hash)
	close(h.verified)
	<-h.resume
	return valid, err
}

func TestPasswordResetRejectsInFlightPasswordVerification(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	username := "race_" + uuid.NewString()[:8]
	user, err := mod.AuthService.SeedAdmin(ctx, username, username+"@example.test", "Race", "Original-password-123")
	if err != nil {
		t.Fatal(err)
	}
	hasher := &pausedPasswordVerifier{crypto.NewArgon2idHasher(), make(chan struct{}), make(chan struct{})}
	auth := services.NewAuthService(postgres.NewUserRepository(db), postgres.NewPasswordRepository(db), postgres.NewSessionRepository(db), postgres.NewLoginAttemptRepository(db), hasher, crypto.NewSessionTokenGenerator(), transaction.NewManager(db.Pool))
	result := make(chan error, 1)
	go func() {
		_, _, _, err := auth.Login(ctx, username, "Original-password-123", "127.0.0.1", "test")
		result <- err
	}()
	select {
	case <-hasher.verified:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	resetErr := mod.UserService.ResetPassword(ctx, user.ID, user.ID, "Replacement-password-123", user.Version)
	close(hasher.resume)
	if resetErr != nil {
		t.Fatal(resetErr)
	}
	select {
	case err := <-result:
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("old verification issued session: %v", err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	var sessions int
	if err := db.Pool.QueryRow(ctx, "SELECT count(*) FROM iam.sessions WHERE user_id = $1", user.ID).Scan(&sessions); err != nil {
		t.Fatal(err)
	}
	if sessions != 0 {
		t.Fatalf("old password created %d sessions", sessions)
	}
}

func TestPasswordResetAndSuspensionInvalidateMFAAuthentication(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx := context.Background()
	username := "reset_mfa_" + uuid.NewString()[:8]
	user, err := mod.AuthService.SeedAdmin(ctx, username, username+"@example.test", "MFA", "Original-password-123")
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := mod.MFAService.BeginEnrollment(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(enrollment.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := mod.MFAService.ConfirmEnrollment(ctx, user.ID, code); err != nil {
		t.Fatal(err)
	}
	user, err = mod.AuthService.GetUserByIdentifier(ctx, username)
	if err != nil {
		t.Fatal(err)
	}
	pending, _, err := mod.AuthService.CreateMFAChallenge(ctx, user, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	consumedToken, _, err := mod.AuthService.CreateMFAChallenge(ctx, user, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	consumed, err := mod.MFAService.CompleteLogin(ctx, consumedToken, enrollment.Codes[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := mod.UserService.ResetPassword(ctx, user.ID, user.ID, "Replacement-password-123", user.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.MFAService.CompleteLogin(ctx, pending, enrollment.Codes[1]); !errors.Is(err, domain.ErrMFAChallengeExpired) {
		t.Fatalf("pending MFA survives reset: %v", err)
	}
	if _, _, err := mod.AuthService.CreateMFAChallenge(ctx, user, "127.0.0.1", "test"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("stale verification created challenge: %v", err)
	}
	if _, _, _, err := mod.AuthService.CreateSessionForUser(ctx, consumed.UserID, consumed.UserVersion, "127.0.0.1", "test"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("completed MFA survives reset: %v", err)
	}
	// A fresh password can still complete MFA and produce a valid session.
	_, _, current, err := mod.AuthService.Login(ctx, username, "Replacement-password-123", "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrMFARequired) {
		t.Fatalf("fresh MFA login: %v", err)
	}
	fresh, _, err := mod.AuthService.CreateMFAChallenge(ctx, current, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	challenge, err := mod.MFAService.CompleteLogin(ctx, fresh, enrollment.Codes[1])
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := mod.AuthService.CreateSessionForUser(ctx, challenge.UserID, challenge.UserVersion, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := mod.AuthService.ValidateSession(ctx, token); err != nil {
		t.Fatal(err)
	}
	pending, _, err = mod.AuthService.CreateMFAChallenge(ctx, current, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	actor, err := mod.AuthService.SeedAdmin(ctx, "actor_"+uuid.NewString()[:8], uuid.NewString()+"@example.test", "Actor", "Actor-password-123")
	if err != nil {
		t.Fatal(err)
	}
	disabled, err := mod.UserService.ChangeStatus(ctx, actor.ID, current.ID, domain.UserStatusSuspended, current.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.UserService.ChangeStatus(ctx, actor.ID, current.ID, domain.UserStatusActive, disabled.Version); err != nil {
		t.Fatal(err)
	}
	if _, err := mod.MFAService.CompleteLogin(ctx, pending, enrollment.Codes[2]); !errors.Is(err, domain.ErrMFAChallengeExpired) {
		t.Fatalf("pending MFA survives suspension/reactivation: %v", err)
	}
}

func TestLogoutAllRejectsPreviouslyVerifiedAuthentication(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx := context.Background()
	username := "logout_" + uuid.NewString()[:8]
	user, err := mod.AuthService.SeedAdmin(ctx, username, username+"@example.test", "Logout", "Original-password-123")
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := mod.AuthService.Login(ctx, username, "Original-password-123", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := mod.AuthService.LogoutAll(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := mod.AuthService.ValidateSession(ctx, token); err == nil {
		t.Fatal("session survives logout-all")
	}
	if _, _, _, err := mod.AuthService.CreateSessionForUser(ctx, user.ID, user.Version, "127.0.0.1", "test"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("old verification survives logout-all: %v", err)
	}
	if _, _, _, err := mod.AuthService.Login(ctx, username, "Original-password-123", "127.0.0.1", "test"); err != nil {
		t.Fatalf("fresh authentication failed: %v", err)
	}
}

func TestLogoutAllReportsDatabaseFailure(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx := context.Background()
	username := "logout_fail_" + uuid.NewString()[:8]
	user, err := mod.AuthService.SeedAdmin(ctx, username, username+"@example.test", "Logout failure", "Original-password-123")
	if err != nil {
		t.Fatal(err)
	}
	_, token, _, err := mod.AuthService.Login(ctx, username, "Original-password-123", "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(ctx, `CREATE FUNCTION iam.fail_auth_revocation_test() RETURNS trigger AS $$
 BEGIN RAISE EXCEPTION 'injected authentication revocation failure'; END; $$ LANGUAGE plpgsql;
 CREATE TRIGGER fail_auth_revocation_test BEFORE UPDATE ON iam.users FOR EACH ROW EXECUTE FUNCTION iam.fail_auth_revocation_test();`)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Pool.Exec(context.Background(), "DROP FUNCTION iam.fail_auth_revocation_test() CASCADE")
	response := performJSONRequest(t, mod.Handler.Routes(), "POST", "/logout-all", nil, &http.Cookie{Name: "erp_session", Value: token})
	if response.Code != 500 {
		t.Fatalf("logout-all hid failure: %d %s", response.Code, response.Body.String())
	}
	if len(response.Result().Cookies()) != 0 {
		t.Fatal("failed revocation discarded the current cookie")
	}
	var version int64
	if err := db.Pool.QueryRow(ctx, "SELECT version FROM iam.users WHERE id=$1", user.ID).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != user.Version {
		t.Fatal("failed revocation changed user version")
	}
	if _, _, err := mod.AuthService.ValidateSession(ctx, token); err != nil {
		t.Fatal("failed revocation unexpectedly deleted session")
	}
}

func TestMFAActivationInvalidatesPriorVerificationAndSerializesAttemptBudget(t *testing.T) {
	mod, db := setupIAMTest(t)
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	username := "mfa_budget_" + uuid.NewString()[:8]
	user, err := mod.AuthService.SeedAdmin(ctx, username, username+"@example.test", "MFA budget", "Original-password-123")
	if err != nil {
		t.Fatal(err)
	}
	enrollment, err := mod.MFAService.BeginEnrollment(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(enrollment.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := mod.MFAService.ConfirmEnrollment(ctx, user.ID, code); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := mod.AuthService.CreateSessionForUser(ctx, user.ID, user.Version, "127.0.0.1", "test"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("pre-MFA verification issued session: %v", err)
	}
	_, _, current, err := mod.AuthService.Login(ctx, username, "Original-password-123", "127.0.0.1", "test")
	if !errors.Is(err, domain.ErrMFARequired) {
		t.Fatalf("fresh login bypassed MFA: %v", err)
	}
	token, _, err := mod.AuthService.CreateMFAChallenge(ctx, current, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	const attempts = 12
	results := make(chan error, attempts)
	var workers sync.WaitGroup
	for i := 0; i < attempts; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := mod.MFAService.CompleteLogin(ctx, token, "INVALID-RECOVERY-CODE")
			results <- err
		}()
	}
	workers.Wait()
	close(results)
	verified, expired := 0, 0
	for err := range results {
		if errors.Is(err, domain.ErrMFAInvalidCode) {
			verified++
		} else if errors.Is(err, domain.ErrMFAChallengeExpired) {
			expired++
		} else {
			t.Fatalf("unexpected guess outcome: %v", err)
		}
	}
	if verified != 5 || expired != attempts-5 {
		t.Fatalf("concurrent attempt budget bypassed: verified=%d expired=%d", verified, expired)
	}
	if _, err := mod.MFAService.CompleteLogin(ctx, token, enrollment.Codes[0]); !errors.Is(err, domain.ErrMFAChallengeExpired) {
		t.Fatalf("exhausted challenge accepted valid code: %v", err)
	}
	fresh, _, err := mod.AuthService.CreateMFAChallenge(ctx, current, "127.0.0.1", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mod.MFAService.CompleteLogin(ctx, fresh, enrollment.Codes[0]); err != nil {
		t.Fatalf("fresh challenge cannot use unconsumed recovery code: %v", err)
	}
}
