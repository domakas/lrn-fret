package progress

import (
	"math/rand/v2"
	"slices"
	"time"
)

// maxLearning is how many not-yet-mastered cards may be in rotation before
// new ones stop being introduced.
const maxLearning = 6

// Pick chooses the next card from `cards` (given in curriculum order).
//
//   - New cards are introduced in curriculum order while fewer than
//     maxLearning cards are still being learned.
//   - Otherwise a seen card is drawn, weighted towards misses, slow answers
//     and cards not seen for a while. Mastered cards still come back, rarely.
//   - `avoid` (recently asked cards) is skipped when there's anything else.
//
// reviewOnly disables introducing new cards (used by quick mix).
func (s *Store) Pick(cards []string, fastMs int, avoid []string, reviewOnly bool, rng *rand.Rand) string {
	var seen, unseen []string
	learning := 0
	for _, k := range cards {
		c := s.Cards[k]
		if c == nil {
			unseen = append(unseen, k)
			continue
		}
		seen = append(seen, k)
		if !c.Mastered {
			learning++
		}
	}
	notAvoided := func(ks []string) []string {
		return slices.DeleteFunc(slices.Clone(ks), func(k string) bool { return slices.Contains(avoid, k) })
	}
	fresh := notAvoided(seen)
	if len(unseen) > 0 && !reviewOnly {
		// introduce a new card at the start, sometimes while few are being
		// learned, and whenever everything seen was just asked
		if len(fresh) == 0 || (learning < maxLearning && rng.Float64() < 0.5) {
			if nu := notAvoided(unseen); len(nu) > 0 {
				return nu[0]
			}
		}
	}
	pool := fresh
	if len(pool) == 0 {
		pool = notAvoided(cards)
	}
	if len(pool) == 0 {
		pool = cards
	}
	weights := make([]float64, len(pool))
	total := 0.0
	for i, k := range pool {
		weights[i] = s.weight(k, fastMs)
		total += weights[i]
	}
	r := rng.Float64() * total
	for i, w := range weights {
		r -= w
		if r <= 0 {
			return pool[i]
		}
	}
	return pool[len(pool)-1]
}

func (s *Store) weight(key string, fastMs int) float64 {
	c := s.Cards[key]
	if c == nil {
		return 1
	}
	w := 0.3
	if !c.Mastered {
		w = 2
	}
	w += 4 * (1 - c.Accuracy())
	if avg := c.AvgMs(); avg > fastMs {
		w += min(3, float64(avg-fastMs)/float64(fastMs))
	}
	// a little boost for things not seen in a while
	if since := time.Since(c.Last); since > 24*time.Hour {
		w += min(2, since.Hours()/24/3)
	}
	return w
}

// Summary counts seen and mastered cards out of a list.
func (s *Store) Summary(cards []string) (seen, mastered int) {
	for _, k := range cards {
		if c := s.Cards[k]; c != nil {
			seen++
			if c.Mastered {
				mastered++
			}
		}
	}
	return
}
