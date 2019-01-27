package memory

import (
	"testing"
	"time"

	"github.com/clockworksoul/cog2/data/rest"
)

func TestTokenGenerate(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_generate"})
	defer da.UserDelete("test_generate")
	if err != nil {
		t.Error(err)
	}

	token, err := da.TokenGenerate("test_generate", 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s\n", token.Token)

	if token.Duration != 10*time.Minute {
		t.Error("Duration mismatch")
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
	if err == nil {
		t.Error("Expected an error")
	}

	err = da.UserCreate(rest.User{Username: "test_uretrieve"})
	defer da.UserDelete("test_uretrieve")
	if err != nil {
		t.Error(err)
	}

	token, err := da.TokenGenerate("test_uretrieve", 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	rtoken, err := da.TokenRetrieveByUser("test_uretrieve")
	if err != nil {
		t.Error(err)
	}

	if token != rtoken {
		t.Error("token mismatch")
	}
}

func TestTokenRetrieveByToken(t *testing.T) {
	_, err := da.TokenRetrieveByToken("no-such-token")
	if err == nil {
		t.Error("Expected an error")
	}

	err = da.UserCreate(rest.User{Username: "test_tretrieve"})
	defer da.UserDelete("test_tretrieve")
	if err != nil {
		t.Error(err)
	}

	token, err := da.TokenGenerate("test_tretrieve", 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	rtoken, err := da.TokenRetrieveByToken(token.Token)
	if err != nil {
		t.Error(err)
	}

	if token != rtoken {
		t.Error("token mismatch")
	}
}

func TestTokenExpiry(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_expires"})
	defer da.UserDelete("test_expires")
	if err != nil {
		t.Error(err)
	}

	token, err := da.TokenGenerate("test_expires", 1*time.Second)
	if err != nil {
		t.Error(err)
	}

	if token.IsExpired() {
		t.Error("Expected token to be unexpired")
	}

	time.Sleep(time.Second)

	if !token.IsExpired() {
		t.Error("Expected token to be expired")
	}
}

func TestTokenInvalidate(t *testing.T) {
	err := da.UserCreate(rest.User{Username: "test_invalidate"})
	defer da.UserDelete("test_invalidate")
	if err != nil {
		t.Error(err)
	}

	token, err := da.TokenGenerate("test_invalidate", 10*time.Minute)
	if err != nil {
		t.Error(err)
	}

	if !da.TokenEvaluate(token.Token) {
		t.Error("Expected token to be valid")
	}

	err = da.TokenInvalidate(token.Token)
	if err != nil {
		t.Error(err)
	}

	if da.TokenEvaluate(token.Token) {
		t.Error("Expected token to be invalid")
	}
}
