package progress

import (
	"math/rand/v2"
	"path/filepath"
	"testing"
)

func TestMastery(t *testing.T) {
	s := &Store{Cards: map[string]*Card{}}
	s.Record("a", true, 1000, 3000)
	if !s.Get("a").Mastered {
		t.Error("fast correct first answer should master the card")
	}
	s.Record("a", false, 1000, 3000)
	if s.Get("a").Mastered {
		t.Error("a miss should unmaster")
	}
	s.Record("a", true, 1000, 3000)
	if s.Get("a").Mastered {
		t.Error("one correct after a miss is not enough")
	}
	s.Record("a", true, 1000, 3000)
	if !s.Get("a").Mastered {
		t.Error("two fast correct in a row should master")
	}
	s.Record("b", true, 9000, 3000)
	s.Record("b", true, 9000, 3000)
	if s.Get("b").Mastered {
		t.Error("slow answers should not master")
	}
}

func TestPickIntroducesInOrder(t *testing.T) {
	s := &Store{Cards: map[string]*Card{}}
	rng := rand.New(rand.NewPCG(1, 2))
	cards := []string{"a", "b", "c"}
	if got := s.Pick(cards, 3000, nil, false, rng); got != "a" {
		t.Fatalf("first pick = %s; want a", got)
	}
	s.Record("a", false, 5000, 3000)
	got := map[string]int{}
	for i := 0; i < 200; i++ {
		got[s.Pick(cards, 3000, []string{"a"}, false, rng)]++
	}
	if got["b"] == 0 || got["c"] != 0 {
		t.Errorf("after one card, next new card should be b (never c): %v", got)
	}
}

func TestPickReviewOnlyAvoids(t *testing.T) {
	s := &Store{Cards: map[string]*Card{}}
	rng := rand.New(rand.NewPCG(1, 2))
	s.Record("a", true, 100, 3000)
	s.Record("b", false, 100, 3000)
	for i := 0; i < 50; i++ {
		if got := s.Pick([]string{"a", "b", "c"}, 3000, []string{"b"}, true, rng); got != "a" {
			t.Fatalf("got %s; want a (c unseen, b avoided)", got)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x", "progress.json")
	s, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	s.Record("k", true, 500, 3000)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	s2, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c := s2.Get("k"); c == nil || c.Attempts != 1 || !c.Mastered {
		t.Errorf("loaded card = %+v", c)
	}
}

func TestPickIntroducesWhenOnlySeenIsAvoided(t *testing.T) {
	s := &Store{Cards: map[string]*Card{}}
	rng := rand.New(rand.NewPCG(3, 4))
	s.Record("a", true, 9000, 3000)
	for i := 0; i < 50; i++ {
		if got := s.Pick([]string{"a", "b"}, 3000, []string{"a"}, false, rng); got != "b" {
			t.Fatalf("got %s; want b", got)
		}
	}
}
