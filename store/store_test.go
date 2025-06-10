package store

import (
	"testing"
	"time"
)

func TestSetAndTest(t *testing.T) {
	s := NewInMemoryStore()

	if res := s.Set("foo", "bar"); res != "OK" {
		t.Fatalf("Set returned %s, want OK", res)
	}

	if val := s.Get("foo"); val != "bar" {
		t.Fatalf("Get returned %s, want bar", val)
	}
}

func TestExpireAndTTL(t *testing.T) {
	s := NewInMemoryStore()

	s.Set("foo", "bar")

	if res := s.Expire("foo", 5); res != "OK" {
		t.Fatalf("Expire returned %s, want OK", res)
	}

	if ttl := s.GetTtl("foo"); ttl != 5 {
		t.Fatalf("GetTtl returned %d, want 5", ttl)
	}

	time.Sleep(6 * time.Second)

	if val := s.Get("foo"); val != "NULL" {
		t.Fatalf("Get returned %s, want NULL", val)
	}
}

func TestDelete(t *testing.T) {
	s := NewInMemoryStore()

	s.Set("foo", "bar")

	if res := s.Delete("foo"); res != "OK" {
		t.Fatalf("Delete returned %s, want OK", res)
	}
}
