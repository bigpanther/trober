package actions

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestTokenVerify(t *testing.T) {
	return
	os.Setenv("FIREBASE_SA_CRED_FILE", "Path to firebase key")
	client, err := firebaseClient()
	if err != nil {
		t.Fatalf("error getting firebase client: %v\n", err)
	}
	ctx := context.Background()
	var tokenToVerify = "..--ctRs-pp--cg"
	token, err := client.VerifyIDToken(ctx, tokenToVerify)
	if err != nil {
		t.Fatalf("error validating token: %v\n", err)
	}
	user, err := client.GetUser(ctx, token.Subject)
	if err != nil {
		t.Fatalf("error getting user: %v\n", err)
	}
	fmt.Println(user.Email)
}
