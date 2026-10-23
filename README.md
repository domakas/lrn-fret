# lrn-fret

**Learn the guitar fretboard in the terminal: one minute at a time.**

`lrn-fret` is a keyboard-only fretboard trainer for the gaps in your day: while a build runs,
while an agent works, between meetings. Open it in a small terminal split, answer a few
questions, switch away. It remembers what you know and keeps asking about what you don't.

It takes you from naming single notes all the way to following chord changes:
**notes → intervals → triads → scales → chord changes**, with a theory page for every
exercise and a circle of fifths always one key away.

```
Intervals  ✓ 12/14 · streak 5, avg 2.1s

Root C on the low E string. Find every octave (3) — frets 4–12

    0   1   2   3   4   5   6   7   8   9  10  11  12
 e ───║───│───│───│───│───│───│───│───│───│───│───│───│
 B ───║───│───│───│───│───│───│───│───│───│───│───│───│
 G ───║───│───│───│───│─R─│───│───│───│───│───│───│───│
 D ───║───│───│───│───│───│───│───│───│───│─R─│───│───│
 A ───║───│───│───│───│───│───│───│───│───│───│───│───│
 E ───║───│───│───│───│───│───│───│─R─│───│───│───│───│
                •       •       •       •           :

> █  found 2/3 · wasd move · space/enter pick

? theory · tab circle · / reveal · esc menu
```

## Install

Requires Go 1.25+.

```sh
go install github.com/domakas/lrn-fret/cmd/fret@latest
```

Or from a clone:

```sh
git clone https://github.com/domakas/lrn-fret && cd lrn-fret
go build -o fret ./cmd/fret
```

## Usage

```sh
fret              # menu
fret quick        # quick mix: 10 questions from your weakest cards across all drills
fret <drill>      # jump straight into a drill (ids below)
fret stats        # fretboard heatmap and progress per drill
fret circle       # circle of fifths
```

## Game modes

The drills build on each other. Start at the top; jump around once the basics stick.

| # | Drill | id | What you do |
|---|-------|----|-------------|
| 1 | **Note names** | `notes` | A spot on the neck is marked `?`, type its note: `c`, `f#`, `bb`. Starts with natural notes on the low E and A strings, then the rest of the neck, then sharps and flats. |
| 2 | **Find the note** | `find` | "Find G on the A string": move the cursor along the string, or type the fret. The cursor starts where your last answer was, the way your hand stays on the neck. |
| 3 | **Intervals** | `intervals` | A root is marked `R`. Find its 3, b3, 5, b7… within 4 frets. Octave questions ask for *every* octave in reach. Learn each interval as a shape that works in any key. |
| 4 | **Triad spelling** | `spell` | Type the 1 3 5 of a chord: `Ebm` → `eb gb bb`. |
| 5 | **Triad shapes** | `triads` | Play a triad in a given inversion on a 3-string set (G B e, D G B, A D G, E A D), or say whether a marked note is the 1, 3 or 5. |
| 6 | **Scale degrees** | `scales` | Minor pentatonic, major pentatonic and major scale (3 notes per string), in every position. Roots are marked. Name the degree of the `?`. |
| 7 | **Circle of fifths** | `fifths` | 5ths and 4ths, relative minors, key signatures, and the I IV V of a key. |
| 8 | **Chord changes** | `changes` | A progression loops (e.g. `G  [D]  Em  C`). Stay in one 5-fret zone and hit the 1, 3 or 5 of each chord before the 6-second timer runs out. Your last note stays on screen so you learn to move to the *nearest* chord tone. |

**Quick mix** (`fret quick`) pulls your 10 weakest cards from every drill you've started.
It's the best default for a one-minute break.

## Theory built in

- **`?`**: opens the theory page for the drill you're in: what the notes are, the
  shapes to look for, and how to practice. Examples: the interval shape grid with the B-string
  rule, C major triads on every string set, why the 3rd matters on chord changes.
- **`tab`**: opens the circle of fifths. Majors outside, relative minors inside, key signatures
  below. During a drill it highlights the key and chord in play, so you can see how the
  fretboard and the circle relate.

The clock stops while either is open.

## Keys

| Where | Keys |
|-------|------|
| Everywhere | `?` theory · `tab` circle of fifths · `esc` back · `ctrl+c` quit |
| Menu | `w`/`s` or arrows to move, `enter` start, `1`–`8` jump to a drill, `m` quick mix, `c` circle, `p` stats |
| Typed answers | `space` or `enter` submits. Multi-word answers (`eb gb bb`, `g c d`) submit on the space after the last word |
| Fretboard answers | `wasd` or arrows move, `space`/`enter` pick. For triads `space` marks a note, `enter` submits, `x` clears. Or type `string fret` (`4 7`; string 1 = high e), or just the fret when the string is given |
| During a drill | `/` reveal the answer (counts as a miss) |

## How it learns you

Every fact is a **card**: "string 4, fret 7 is A", "C major, 1st inversion on G B e", "the b3
above a root on the A string".

- A card counts as **known** after two fast, correct answers in a row, or one if you get it
  right fast the very first time. Things you already know go by quickly.
- New cards come in curriculum order while fewer than 6 are still being learned.
- Misses, slow answers and cards you haven't seen in a while come back more often. Known
  cards still come back now and then.
- `fret stats` shows the whole neck colored by how well you know each note, plus progress
  per drill.

Progress is saved after every answer to `~/.lrn-fret/progress.json`. Set `LRN_FRET_HOME` to
keep it somewhere else.

## Made for background use

- **Small footprint:** the board is ~55 columns wide, so it fits a tmux split or a side pane.
- **Pauses when you look away:** when the terminal loses focus the drill is hidden behind a
  "paused" card and the clock stops. It comes back when you return. This needs a terminal that
  reports focus (iTerm2, Ghostty, kitty, WezTerm, VS Code…). In tmux, add
  `set -g focus-events on` to `~/.tmux.conf`.
- **Stops at any time:** `esc` or `ctrl+c`, nothing is lost.

### macOS: holding `a`/`d` opens the accent picker

That's macOS press-and-hold, and the app can't intercept it. Turn it off for your terminal only:

```sh
defaults write <terminal bundle id> ApplePressAndHoldEnabled -bool false
# e.g. com.googlecode.iterm2, com.mitchellh.ghostty, com.apple.Terminal
```

Then restart the terminal. Arrow keys never trigger it.

## Development

```sh
go test ./...
go run ./cmd/fret
```

| Package | What's in it |
|---------|--------------|
| `internal/theory` | Notes, spelling, fretboard, triad voicings, scale positions, circle of fifths. No UI. |
| `internal/progress` | Card history, mastery, and the scheduler that picks the next card. |
| `internal/drill` | One file per drill. Each card becomes a `Question` with its own checker. Theory pages live in `theorytext.go`. |
| `internal/ui` | The [Bubble Tea](https://github.com/charmbracelet/bubbletea) app: menu, drills, theory, circle, stats. |

The tests check, among other things, that every card in every drill produces a question its
own answer passes.

To add a drill, implement `drill.Module` (cards in curriculum order, `Make`, `Theory`) and add it
to `drill.All()`.

## Roadmap

- Mic input: play the answer on the guitar (pitch detection)
- Adjustable chord-change tempo
- Diminished and 7th chords
- Other tunings, left-handed view
