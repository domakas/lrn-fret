package drill

import (
	"fmt"
	"strings"

	"github.com/domakas/lrn-fret/internal/theory"
)

// The theory pages, one per module. Press ? in a drill to open them.

func (Notes) Theory() string {
	return `NOTE NAMES

The musical alphabet is A B C D E F G, then it repeats. With sharps and
flats there are 12 notes. One fret = one half step.

  C  C#/Db  D  D#/Eb  E  F  F#/Gb  G  G#/Ab  A  A#/Bb  B  (C)

There is NO sharp between B–C and E–F. Every other pair of natural
notes has one fret in between. A sharp and a flat on the same fret
are the same note (C# = Db).

Standard tuning, low to high: E A D G B e
  ("Eddie Ate Dynamite, Good Bye Eddie"). String 6 = low E, 1 = high e.

LANDMARKS
  • Fret 12 = the open string, an octave up. The board repeats there.
  • Fret 5 on any string = the next open string up (G string: fret 4).
  • Dots at 3 5 7 9 12 are your anchors.

NATURAL NOTES ON THE TWO ROOT STRINGS (learn these first)
  E string:  E 0  F 1  G 3  A 5  B 7  C 8  D 10  E 12
  A string:  A 0  B 2  C 3  D 5  E 7  F 8  G 10  A 12

OCTAVE SHAPES (use them to find a note everywhere)
  • From E or A string: two strings up, two frets up.   (E5 A → D7 A)
  • From D or G string: two strings up, three frets up (B string shift).
  • Low E string and high e string are the same notes.

HOW TO PRACTICE
  Name the note out loud before typing. Learn naturals first, then
  sharps/flats as "one fret above/below a natural you already know".
  Press Tab for the circle of fifths.`
}

func (Find) Theory() string {
	return Notes{}.Theory() + `

FINDING A NOTE FAST
  1. Go to the nearest landmark you know (open string, 5, 7, 12).
  2. Count up/down by half steps — remember B–C and E–F are one fret.
  3. Or jump from a string you know via an octave shape.

ANSWERING
  Move with h/l (or ←/→) and press enter, or type the fret number.`
}

func (Intervals) Theory() string {
	return `INTERVALS

An interval is the distance between two notes, counted in half steps
(frets). We name them by scale degree:

  semitones  0  1   2  3   4  5  6   7  8   9  10  11  12
  degree     1  b2  2  b3  3  4  b5  5  b6  6  b7  7   8 (octave)

The good news: on guitar each interval is a SHAPE. Learn the shape
once and it works in every key.

SHAPES WITH THE ROOT ON THE LOW E (or A) STRING
                fret:  -2   -1   R   +1   +2
  two strings up       b6    6   b7   7    8 (octave)
  one string up        b3    3   4    b5   5
  root string                    R         2

  • 5th  = one string up, two frets up  (the power chord)
  • 4th  = one string up, same fret
  • 3rd  = one string up, one fret back;  b3 = two frets back
  • b7   = two strings up, same fret
  • 8va  = two strings up, two frets up

THE B-STRING RULE
  The G→B pair is tuned a major 3rd (4 frets) instead of a 4th (5).
  Whenever a shape crosses from the G string onto the B string, move
  the note on the B string (and above) up ONE fret.

WHY IT MATTERS
  1 3 5 are a chord's tones. If you can see the 3rd and 5th around
  any root, you can build every triad and target notes on chord
  changes. The 3rd decides major (3) vs minor (b3).`
}

func (Spell) Theory() string {
	return `SPELLING TRIADS

A triad is three notes stacked in 3rds: root (1), third (3), fifth (5).

  major      1  3  5    = major 3rd (4 frets) + minor 3rd (3 frets)
  minor      1  b3 5    = minor 3rd (3) + major 3rd (4)
  diminished 1  b3 b5   = minor 3rd + minor 3rd

STEP 1 — skip a letter: every triad uses every other letter.
  C E G   D F A   E G B   F A C   G B D   A C E   B D F

STEP 2 — fix the accidentals. With no sharps/flats you get:
  C F G = major,  Dm Em Am = minor,  Bdim.
  For other chords adjust so the 3rd and 5th have the right size:
    D major: D F A → the 3rd must be 4 frets up → D F# A
    Eb:      E major is E G# B, flatten everything → Eb G Bb

STEP 3 — major ↔ minor: just lower or raise the 3rd.
    C = C E G   →   Cm = C Eb G

Spelling matters: Ebm is Eb Gb Bb (letters E G B), not Eb F# Bb.
The app accepts either, but shows the proper spelling.

TIP: the I IV V of any key are neighbours on the circle of fifths
(Tab). Their triads are major; their relative minors (inner ring) are
ii iii vi.`
}

func (Triads) Theory() string {
	var b strings.Builder
	b.WriteString(`TRIADS ON STRING SETS

A small triad shape = 3 notes on 3 neighbouring strings. You can play
any chord in any zone of the neck this way.

INVERSIONS — which chord tone is the lowest note
  root position   1 3 5  (root in the bass)
  1st inversion   3 5 1  (3rd in the bass)
  2nd inversion   5 1 3  (5th in the bass)

Going up the neck on one string set, the shapes cycle:
  root → 1st → 2nd → root (an octave up) ...
Each shape's notes move up to the NEXT chord tone.

C MAJOR ON EVERY STRING SET (frets listed low string → high string)
`)
	c := theory.Triad{Root: theory.MustNote("C"), Quality: theory.Major}
	for ss := theory.StringSet(1); ss <= 4; ss++ {
		fmt.Fprintf(&b, "  %-6s", ss)
		for inv := 0; inv < 3; inv++ {
			vs := theory.Voicings(c, ss, inv, 15)
			var shapes []string
			for _, v := range vs {
				shapes = append(shapes, fmt.Sprintf("%d-%d-%d", v[0].F, v[1].F, v[2].F))
			}
			fmt.Fprintf(&b, "  %s: %-14s", []string{"root", "1st", "2nd"}[inv], strings.Join(shapes, " / "))
		}
		b.WriteString("\n")
	}
	b.WriteString(`
MINOR = lower the 3rd by one fret. Find the 3 in each shape first: it
decides major/minor and it's the note to target on chord changes.

Remember the B-string rule: shapes that touch the B string look
different from the ones on the lower sets.

HOW TO PRACTICE
  Pick one string set. Play C major root → 1st → 2nd inversion up the
  neck, saying "1 3 5" for each note. Then do the same for C minor.`)
	return b.String()
}

func (Scales) Theory() string {
	return `SCALES AND POSITIONS

  minor pentatonic   1  b3  4  5  b7       (A: A C D E G)
  major pentatonic   1  2   3  5  6        (C: C D E G A)
  major scale        1  2   3  4  5  6  7  (whole whole half whole
                                            whole whole half)

RELATIVE SCALES: A minor pentatonic and C major pentatonic have the same
notes — the same 5 boxes, just a different root. On the circle of
fifths (Tab) they're the inner/outer pair (Am ↔ C).

POSITIONS (boxes)
  The 5 pentatonic boxes link up into one big pattern across the neck.
  Box 1 starts on the root on the low E string. Each next box starts on
  the next scale note. The major scale here uses 3 notes per string,
  which gives 7 positions.

LEARN THE DEGREES, NOT JUST THE DOTS
  Roots are marked R. For every other dot know if it's the b3, 5, b7…
  That's what links scales to chords:
    • A chord's 1 3 5 are inside the scale. Land on them over that chord.
    • In minor pentatonic the b3 and b7 give the "minor/bluesy" sound.
    • In the major scale the 3 and 7 are the "sweet" notes.

HOW TO PRACTICE
  Play a box slowly and say each degree out loud. Then play just the
  1 3 5 inside the box. Then just the b7s. Then move to the next box.`
}

func (Circle) Theory() string {
	return `THE CIRCLE OF FIFTHS

Press Tab to see it. The 12 major keys are on the outside, their
relative minors inside.

  • Clockwise = up a 5th (C → G → D …). Each step adds one sharp.
  • Counter-clockwise = up a 4th (C → F → Bb …). Each step adds a flat.
  • Inner ring = relative minor: same notes and key signature,
    3 frets (a minor 3rd) below. C ↔ Am, G ↔ Em.

KEY SIGNATURES
  sharps: G 1  D 2  A 3  E 4  B 5  F# 6     order: F C G D A E B
  flats:  F 1  Bb 2 Eb 3 Ab 4 Db 5 Gb 6     order: B E A D G C F

CHORDS OF A KEY — they sit together on the circle
  Take a key and its two neighbours: those are I, IV and V (major).
  The inner ring under those three: vi, ii and iii (minor).
    Key of G:  C  G  D   (IV I V)
               Am Em Bm  (ii vi iii)
  This is why I–IV–V and I–V–vi–IV sound right: the chords are next
  to each other.

ON THE FRETBOARD
  • 4th = next string up, same fret. Walking across the strings at one
    fret goes counter-clockwise around the circle (E A D G …, then the
    B-string shift).
  • 5th = next string up, two frets higher (the power chord).`
}

func (*Changes) Theory() string {
	return `PLAYING OVER CHORD CHANGES

When the chord changes, your target notes change. Aim for a CHORD TONE
of the new chord (1, 3 or 5) right on the change. The notes in between
can come from the scale.

CHORDS IN A MAJOR KEY
  I    ii   iii  IV   V    vi   vii°
  maj  min  min  maj  maj  min  dim
  Key of C:  C  Dm  Em  F  G  Am  Bdim
  Key of G:  G  Am  Bm  C  D  Em  F#dim

WHY THE 3RD?
  The 1 and 5 are often shared between chords. The 3rd changes, and
  it's what makes the change sound like it happened. Hitting the 3rd of
  each chord outlines the progression even with no band.

STAY IN ONE ZONE
  All 12 notes live inside any 5-fret window. Don't jump around the
  neck: find the next chord's tone NEAR the last note you played (it's
  shown as a dim marker). Good targets are usually 1–2 frets away.
    G → C:   B (3 of G) moves up a fret to C (1 of C)
    C → G:   E (3 of C) moves down a fret to D (5 of G)
  That small step is called voice leading.

HOW TO PRACTICE
  Loop a backing track (or just count 4 beats per chord). Play only
  the root of each chord. Then only the 3rds. Then only the 5ths.
  Then mix them, staying in one zone. Then move the zone.`
}
