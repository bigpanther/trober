package actions

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestTokenVerify(t *testing.T) {
	return
	//os.Setenv("FIREBASE_SA_CRED_FILE", "Path to firebase key")
	// cat filename | base64
	var encodedJson = ""
	os.Setenv("FIREBASE_SERVICE_ACCOUNT_JSON_ENCODED", encodedJson)
	client, err := firebaseClient()
	if err != nil {
		t.Fatalf("error getting firebase client: %v\n", err)
	}
	ctx := context.Background()
	var tokenToVerify = "..---"
	token, err := client.VerifyIDToken(ctx, tokenToVerify)
	if err != nil {
		t.Fatalf("error validating token: %v\n", err)
	}
	user, err := client.GetUser(ctx, token.Subject)
	if err != nil {
		t.Fatalf("error getting user: %v\n", err)
	}
	//Print the email always
	if user.Email != "test" {
		t.Errorf("found user %s %s %s", user.Email, user.UID, token.Subject)
	}
	fmt.Println(user.Email)
}
