// Package progress stores per-card results and decides what to ask next.
package progress

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const keepResults = 6

type Result struct {
	OK bool `json:"ok"`
	Ms int  `json:"ms"`
}

// Card is the history of one fact, e.g. "note:s4f7".
type Card struct {
	Attempts int       `json:"attempts"`
	Correct  int       `json:"correct"`
	Recent   []Result  `json:"recent"`
	Last     time.Time `json:"last"`
	// Mastered is sticky until the card is missed again.
	Mastered bool `json:"mastered"`
}

// Accuracy of the recent results, 0..1. Unseen cards report 0.
func (c *Card) Accuracy() float64 {
	if c == nil || len(c.Recent) == 0 {
		return 0
	}
	ok := 0
	for _, r := range c.Recent {
		if r.OK {
			ok++
		}
	}
	return float64(ok) / float64(len(c.Recent))
}

// AvgMs is the mean time of recent correct answers.
func (c *Card) AvgMs() int {
	if c == nil {
		return 0
	}
	sum, n := 0, 0
	for _, r := range c.Recent {
		if r.OK {
			sum += r.Ms
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / n
}

type Store struct {
	Cards map[string]*Card `json:"cards"`
	path  string
}

func DefaultPath() string {
	if d := os.Getenv("LRN_FRET_HOME"); d != "" {
		return filepath.Join(d, "progress.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".lrn-fret", "progress.json")
}

func Load(path string) (*Store, error) {
	s := &Store{Cards: map[string]*Card{}, path: path}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	if s.Cards == nil {
		s.Cards = map[string]*Card{}
	}
	return s, nil
}

// Save writes atomically so a ctrl-c mid-write never corrupts progress.
func (s *Store) Save() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) Get(key string) *Card { return s.Cards[key] }

// Record adds a result. fastMs is the time under which a correct answer counts
// as "known"; a card is mastered after two fast correct answers in a row, or
// one on the very first attempt (you already knew it).
func (s *Store) Record(key string, ok bool, ms, fastMs int) {
	c := s.Cards[key]
	if c == nil {
		c = &Card{}
		s.Cards[key] = c
	}
	c.Attempts++
	if ok {
		c.Correct++
	}
	c.Recent = append(c.Recent, Result{ok, ms})
	if len(c.Recent) > keepResults {
		c.Recent = c.Recent[len(c.Recent)-keepResults:]
	}
	c.Last = time.Now()

	fast := func(r Result) bool { return r.OK && r.Ms <= fastMs }
	n := len(c.Recent)
	switch {
	case !ok:
		c.Mastered = false
	case c.Attempts == 1 && fast(c.Recent[0]):
		c.Mastered = true
	case n >= 2 && fast(c.Recent[n-1]) && fast(c.Recent[n-2]):
		c.Mastered = true
	}
}
