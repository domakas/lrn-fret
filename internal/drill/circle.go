package drill

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/domakas/lrn-fret/internal/theory"
)

// Circle: circle-of-fifths facts — neighbors, relative minors, signatures,
// and the I IV V of a key.
type Circle struct{}

func (Circle) ID() string   { return "fifths" }
func (Circle) Name() string { return "Circle of fifths" }
func (Circle) Desc() string { return "5ths, 4ths, relative minors, key signatures, I IV V of a key" }
func (Circle) FastMs() int  { return 5000 }

var circleKinds = []string{"fifth", "fourth", "rel", "sig", "ivv"}

func (Circle) Cards() []string {
	var out []string
	for _, kind := range circleKinds {
		for _, k := range theory.CircleMajors {
			out = append(out, fmt.Sprintf("cof:%s:%s", kind, k))
		}
	}
	return out
}

func (c Circle) Make(card string, _ *rand.Rand) *Question {
	parts := strings.Split(card, ":")
	kind, key := parts[1], theory.MustNote(parts[2])
	i := theory.CircleIndex(key.PC)
	q := &Question{
		Card:    card,
		Kind:    AnswerText,
		NoBoard: true,
		FastMs:  c.FastMs(),
		Circle:  &CircleFocus{Strong: []theory.PC{key.PC}},
	}
	switch kind {
	case "fifth":
		ans := key.Up(4, 7)
		q.Prompt = fmt.Sprintf("What is a 5th above %s? (one step clockwise)", key)
		q.CheckText, q.Answer, q.Expect = checkNote(ans.PC), ans.String(), ans.String()
	case "fourth":
		ans := key.Up(3, 5)
		q.Prompt = fmt.Sprintf("What is a 4th above %s? (one step counter-clockwise)", key)
		q.CheckText, q.Answer, q.Expect = checkNote(ans.PC), ans.String(), ans.String()
	case "rel":
		ans := theory.RelativeMinor(key)
		q.Prompt = fmt.Sprintf("Relative minor of %s major? (type the root, e.g. \"a\" or \"am\")", key)
		q.CheckText = func(in string) (bool, bool) {
			return checkNote(ans.PC)(strings.TrimSuffix(normalize(in), "m"))
		}
		q.Answer = ans.String() + "m"
		q.Expect = q.Answer
	case "sig":
		want := theory.Signature(i)
		q.Prompt = fmt.Sprintf("Key signature of %s major? (e.g. 0, 3#, 2b)", key)
		q.CheckText = func(in string) (bool, bool) {
			in = strings.ReplaceAll(normalize(in), " ", "")
			in = strings.NewReplacer("sharps", "#", "sharp", "#", "flats", "b", "flat", "b").Replace(in)
			if in == "" || strings.Trim(in, "0123456789#b") != "" {
				return false, false
			}
			return in == strings.ToLower(want), true
		}
		q.Answer, q.Expect = want, want
	case "ivv":
		iv, v := key.Up(3, 5), key.Up(4, 7)
		q.Prompt = fmt.Sprintf("I IV V in %s major? (type 3 roots — they're %s and its two neighbors)", key, key)
		q.CheckText = func(in string) (bool, bool) {
			fields := strings.Fields(strings.ReplaceAll(in, ",", " "))
			if len(fields) != 3 {
				return false, false
			}
			for j, want := range []theory.Note{key, iv, v} {
				pc, err := theory.ParseNote(fields[j])
				if err != nil {
					return false, false
				}
				if pc != want.PC {
					return false, true
				}
			}
			return true, true
		}
		q.Answer = fmt.Sprintf("%s %s %s", key, iv, v)
		q.Expect = q.Answer
		q.Words = 3
	}
	q.NoteInput = kind != "sig"
	return q
}
