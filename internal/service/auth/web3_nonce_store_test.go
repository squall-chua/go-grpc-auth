package auth

import (
	"context"
	"testing"
	"time"
)

func TestMemoryNonceStore_SaveAndConsume(t *testing.T) {
	store := NewMemoryNonceStore()
	ctx := context.Background()
	ns := "default"
	wallet := "0xabc"
	nonce := "n1"

	if err := store.Save(ctx, ns, wallet, nonce, time.Minute); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ok, err := store.Consume(ctx, ns, wallet, nonce)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if !ok {
		t.Fatalf("expected consume to succeed")
	}

	ok, err = store.Consume(ctx, ns, wallet, nonce)
	if err != nil {
		t.Fatalf("Consume (2nd): %v", err)
	}
	if ok {
		t.Fatalf("expected second consume to fail (single-use)")
	}
}

func TestMemoryNonceStore_NotFound(t *testing.T) {
	store := NewMemoryNonceStore()
	ctx := context.Background()
	ok, err := store.Consume(ctx, "default", "0xabc", "missing")
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if ok {
		t.Fatalf("expected consume to fail for missing nonce")
	}
}

func TestMemoryNonceStore_TTL(t *testing.T) {
	store := NewMemoryNonceStore()
	ctx := context.Background()
	ns := "default"
	wallet := "0xabc"
	nonce := "n1"

	if err := store.Save(ctx, ns, wallet, nonce, 50*time.Millisecond); err != nil {
		t.Fatalf("Save: %v", err)
	}

	time.Sleep(80 * time.Millisecond)

	ok, err := store.Consume(ctx, ns, wallet, nonce)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if ok {
		t.Fatalf("expected consume to fail after TTL")
	}
}

func TestMemoryNonceStore_DifferentWallets(t *testing.T) {
	store := NewMemoryNonceStore()
	ctx := context.Background()
	ns := "default"
	walletA := "0xaaa"
	walletB := "0xbbb"
	nonce := "n1"

	if err := store.Save(ctx, ns, walletA, nonce, time.Minute); err != nil {
		t.Fatalf("Save A: %v", err)
	}

	ok, err := store.Consume(ctx, ns, walletB, nonce)
	if err != nil {
		t.Fatalf("Consume B: %v", err)
	}
	if ok {
		t.Fatalf("expected consume with wrong wallet to fail")
	}

	ok, err = store.Consume(ctx, ns, walletA, nonce)
	if err != nil {
		t.Fatalf("Consume A: %v", err)
	}
	if !ok {
		t.Fatalf("expected consume with original wallet to succeed")
	}
}
