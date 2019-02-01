package postgres

import (
	"testing"
	"time"

	"github.com/clockworksoul/cog2/data/rest"
	"github.com/clockworksoul/cog2/dataaccess/errs"
)

func TestTokenGenerate(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_generate"})
	defer da.UserDelete("test_generate")
	expectNoErr(t, err)

	token, err := da.TokenGenerate("test_generate", 10*time.Minute)
	defer da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	t.Logf("%s\n", token.Token)

	if token.Duration != 10*time.Minute {
		t.Errorf("Duration mismatch: %v vs %v\n", token.Duration, 10*time.Minute)
	}

	if token.User != "test_generate" {
		t.Error("User mismatch")
	}

	if token.ValidFrom.Add(10*time.Minute) != token.ValidUntil {
		t.Error("Validity duration mismatch")
	}
}

func TestTokenRetrieveByUser(t *testing.T) {
	_, err := da.TokenRetrieveByUser("no-such-user")
	expectErr(t, err, errs.ErrNoSuchToken)

	err = da.UserCreate(rest.User{Username: "test_uretrieve", Email: "test_uretrieve"})
	defer da.UserDelete("test_uretrieve")
	expectNoErr(t, err)

	token, err := da.TokenGenerate("test_uretrieve", 10*time.Minute)
	defer da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	rtoken, err := da.TokenRetrieveByUser("test_uretrieve")
	expectNoErr(t, err)

	if token.Token != rtoken.Token {
		t.Error("token mismatch")
	}
}

func TestTokenRetrieveByToken(t *testing.T) {
	_, err := da.TokenRetrieveByToken("no-such-token")
	expectErr(t, err, errs.ErrNoSuchToken)

	err = da.UserCreate(rest.User{Username: "test_tretrieve", Email: "test_tretrieve"})
	defer da.UserDelete("test_tretrieve")
	expectNoErr(t, err)

	token, err := da.TokenGenerate("test_tretrieve", 10*time.Minute)
	defer da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	rtoken, err := da.TokenRetrieveByToken(token.Token)
	expectNoErr(t, err)

	if token.Token != rtoken.Token {
		t.Error("token mismatch")
	}
}

func TestTokenExpiry(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_expires", Email: "test_expires"})
	defer da.UserDelete("test_expires")
	expectNoErr(t, err)

	token, err := da.TokenGenerate("test_expires", 1*time.Second)
	defer da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	if token.IsExpired() {
		t.Error("Expected token to be unexpired")
	}

	time.Sleep(time.Second)

	if !token.IsExpired() {
		t.Error("Expected token to be expired")
	}
}

func TestTokenInvalidate(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_invalidate", Email: "test_invalidate"})
	defer da.UserDelete("test_invalidate")
	expectNoErr(t, err)

	token, err := da.TokenGenerate("test_invalidate", 10*time.Minute)
	defer da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	if !da.TokenEvaluate(token.Token) {
		t.Error("Expected token to be valid")
	}

	err = da.TokenInvalidate(token.Token)
	expectNoErr(t, err)

	if da.TokenEvaluate(token.Token) {
		t.Error("Expected token to be invalid")
	}
}
