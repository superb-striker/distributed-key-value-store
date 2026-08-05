package store

import (
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestPutGet(t *testing.T) {
	s := openTestStore(t)
	if err := s.Put("foo", []byte("bar")); err != nil {
		t.Fatalf("Put: %v", err)
	}
	v, err := s.Get("foo")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(v) != "bar" {
		t.Fatalf("got %q, want %q", v, "bar")
	}
}

func TestGetMissingKey(t *testing.T) {
	s := openTestStore(t)
	_, err := s.Get("nope")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestOverwrite(t *testing.T) {
	s := openTestStore(t)
	s.Put("k", []byte("v1"))
	s.Put("k", []byte("v2"))
	v, _ := s.Get("k")
	if string(v) != "v2" {
		t.Fatalf("got %q, want v2", v)
	}
}

func TestDelete(t *testing.T) {
	s := openTestStore(t)
	s.Put("k", []byte("v"))
	if err := s.Delete("k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get("k"); err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound after delete, got %v", err)
	}
}

func TestDeleteMissingKeyIsNotAnError(t *testing.T) {
	s := openTestStore(t)
	if err := s.Delete("never-existed"); err != nil {
		t.Fatalf("Delete on missing key should be idempotent, got: %v", err)
	}
}

func TestPersistenceAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	s1, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	s1.Put("durable", []byte("value"))
	s1.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	v, err := s2.Get("durable")
	if err != nil {
		t.Fatalf("Get after reopen: %v", err)
	}
	if string(v) != "value" {
		t.Fatalf("got %q, want %q", v, "value")
	}
}
