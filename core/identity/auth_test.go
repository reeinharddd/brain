package identity

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManagerLoginAndAuthenticate(t *testing.T) {
	t.Run("bootstrap credentials create a session", func(t *testing.T) {
		manager := NewManager(Config{
			Mode:              ModeBootstrap,
			Required:          true,
			BootstrapEmail:    "owner@brain.local",
			BootstrapPassword: "secret123",
			BootstrapName:     "Brain Owner",
			BootstrapRole:     RoleOwner,
			SessionTTL:        time.Hour,
		})

		session, err := manager.Login(context.Background(), "owner@brain.local", "secret123")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if session == nil {
			t.Fatal("expected session, got nil")
		}
		if session.User.Email != "owner@brain.local" {
			t.Fatalf("expected email owner@brain.local, got %q", session.User.Email)
		}
		if session.User.Role != RoleOwner {
			t.Fatalf("expected owner role, got %q", session.User.Role)
		}
		if len(session.User.Capabilities) == 0 {
			t.Fatal("expected capabilities to be populated")
		}

		validated, err := manager.Authenticate(context.Background(), session.Token)
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if validated.Token != session.Token {
			t.Fatalf("expected token %q, got %q", session.Token, validated.Token)
		}
	})

	t.Run("invalid bootstrap credentials are rejected", func(t *testing.T) {
		manager := NewManager(Config{
			Mode:              ModeBootstrap,
			Required:          true,
			BootstrapEmail:    "owner@brain.local",
			BootstrapPassword: "secret123",
			SessionTTL:        time.Hour,
		})

		_, err := manager.Login(context.Background(), "owner@brain.local", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestManagerAnonymousLogin(t *testing.T) {
	manager := NewManager(Config{
		Mode:           ModeBootstrap,
		Required:       false,
		AllowAnonymous: true,
		SessionTTL:     time.Hour,
	})

	session, err := manager.Login(context.Background(), "dev@brain.local", "devpass")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if session.User.Role != RoleOwner {
		t.Fatalf("expected anonymous development session to default to owner role, got %q", session.User.Role)
	}
	if got := manager.Status(context.Background(), session.Token); !got.Authenticated {
		t.Fatal("expected authenticated status for valid token")
	}
}

func TestManagerLogoutAndStatus(t *testing.T) {
	manager := NewManager(Config{
		Mode:              ModeBootstrap,
		Required:          true,
		BootstrapEmail:    "admin@brain.local",
		BootstrapPassword: "secret123",
		BootstrapRole:     RoleAdmin,
		SessionTTL:        time.Hour,
	})

	session, err := manager.Login(context.Background(), "admin@brain.local", "secret123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	status := manager.Status(context.Background(), session.Token)
	if !status.Authenticated {
		t.Fatal("expected authenticated status")
	}
	if len(status.AllowedSections) == 0 {
		t.Fatal("expected allowed sections in status")
	}
	if status.User == nil || status.User.Email != "admin@brain.local" {
		t.Fatalf("expected user email in status, got %#v", status.User)
	}

	if err := manager.Logout(context.Background(), session.Token); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if _, err := manager.Authenticate(context.Background(), session.Token); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound after logout, got %v", err)
	}

	status = manager.Status(context.Background(), session.Token)
	if status.Authenticated {
		t.Fatal("expected unauthenticated status after logout")
	}
}

func TestCapabilitiesAndSectionsByRole(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		wantCapLen int
		wantSecLen int
	}{
		{name: "owner", role: RoleOwner, wantCapLen: 11, wantSecLen: 9},
		{name: "admin", role: RoleAdmin, wantCapLen: 10, wantSecLen: 9},
		{name: "member", role: RoleMember, wantCapLen: 7, wantSecLen: 7},
		{name: "viewer", role: RoleViewer, wantCapLen: 3, wantSecLen: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := len(CapabilitiesForRole(tt.role)); got != tt.wantCapLen {
				t.Fatalf("expected %d capabilities, got %d", tt.wantCapLen, got)
			}
			if got := len(SectionsForRole(tt.role)); got != tt.wantSecLen {
				t.Fatalf("expected %d sections, got %d", tt.wantSecLen, got)
			}
		})
	}
}

func TestStatusWithoutSession(t *testing.T) {
	manager := NewManager(Config{Required: false})
	status := manager.Status(context.Background(), "")
	if status.Authenticated {
		t.Fatal("expected unauthenticated status")
	}
	if status.Required {
		t.Fatal("expected auth to be optional by default in this test")
	}
	if len(status.AllowedSections) == 0 {
		t.Fatal("expected allowed sections to be populated")
	}
}
