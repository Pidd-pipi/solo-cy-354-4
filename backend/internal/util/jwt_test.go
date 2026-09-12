package util

import "testing"

func TestJWTRoundTrip(t *testing.T) {
	secret := "unit-test-secret"
	token, err := GenerateToken(secret, 7, "13800000000", "student", 1)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{name: "valid token", token: token, wantErr: false},
		{name: "tampered token", token: token + "x", wantErr: true},
		{name: "empty token", token: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseToken(secret, tt.token)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims.UserID != 7 || claims.Role != "student" {
				t.Fatalf("unexpected claims: %+v", claims)
			}
		})
	}
}
