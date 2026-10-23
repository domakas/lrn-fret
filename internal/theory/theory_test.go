package theory

import (
	"slices"
	"testing"
)

func TestParseNote(t *testing.T) {
	cases := map[string]PC{"c": 0, "C#": 1, "db": 1, "bb": 10, "B": 11, "e#": 5, "Fb": 4, "cb": 11, "g##": 9, "as": 10}
	for in, want := range cases {
		got, err := ParseNote(in)
		if err != nil || got != want {
			t.Errorf("ParseNote(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "h", "c#x", "1"} {
		if _, err := ParseNote(bad); err == nil {
			t.Errorf("ParseNote(%q) should fail", bad)
		}
	}
}

func TestPitchAt(t *testing.T) {
	cases := map[Pos]string{{6, 0}: "E", {6, 5}: "A", {5, 3}: "C", {3, 4}: "B", {2, 1}: "C", {1, 12}: "E", {4, 7}: "A"}
	for p, want := range cases {
		if got := PitchAt(p).Sharp(); got != want {
			t.Errorf("PitchAt(%v) = %s; want %s", p, got, want)
		}
	}
}

func TestTriadSpelling(t *testing.T) {
	cases := map[string][]string{
		"C":    {"C", "E", "G"},
		"Ebm":  {"Eb", "Gb", "Bb"},
		"F#":   {"F#", "A#", "C#"},
		"Bdim": {"B", "D", "F"},
		"Db":   {"Db", "F", "Ab"},
	}
	for name, want := range cases {
		var tr Triad
		switch {
		case len(name) > 3 && name[len(name)-3:] == "dim":
			tr = Triad{MustNote(name[:len(name)-3]), Dim}
		case name[len(name)-1] == 'm':
			tr = Triad{MustNote(name[:len(name)-1]), Minor}
		default:
			tr = Triad{MustNote(name), Major}
		}
		var got []string
		for _, n := range tr.Tones() {
			got = append(got, n.String())
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s tones = %v; want %v", name, got, want)
		}
	}
}

func TestVoicings(t *testing.T) {
	// C major root position on G B e: C on G5, E on B5, G on e3.
	vs := Voicings(Triad{MustNote("C"), Major}, 1, 0, 15)
	want := [3]Pos{{3, 5}, {2, 5}, {1, 3}}
	if !slices.Contains(vs, want) {
		t.Fatalf("C root pos on set 1 = %v; want contains %v", vs, want)
	}
	for _, v := range vs {
		if !MatchesVoicing(v[:], Triad{MustNote("C"), Major}, 1, 0) {
			t.Errorf("voicing %v should match", v)
		}
	}
	// every triad/set/inversion has at least one voicing within 15 frets
	for _, r := range Roots {
		for q := Major; q <= Minor; q++ {
			for ss := StringSet(1); ss <= 4; ss++ {
				for inv := 0; inv < 3; inv++ {
					if len(Voicings(Triad{MustNote(r), q}, ss, inv, 15)) == 0 {
						t.Errorf("no voicing for %s%s set %d inv %d", r, q.Suffix(), ss, inv)
					}
				}
			}
		}
	}
}

func TestMinorPentBox1(t *testing.T) {
	box := MinorPent.Box(MustNote("A").PC, 0)
	var frets []int
	for _, n := range box {
		frets = append(frets, n.Pos.F)
	}
	want := []int{5, 8, 5, 7, 5, 7, 5, 7, 5, 8, 5, 8}
	if !slices.Equal(frets, want) {
		t.Errorf("A minor pent box 1 frets = %v; want %v", frets, want)
	}
}

func TestBoxesAreInScale(t *testing.T) {
	for _, sc := range Scales {
		for _, k := range Keys {
			root := MustNote(k).PC
			for p := 0; p < sc.Positions(); p++ {
				box := sc.Box(root, p)
				for _, n := range box {
					if PitchAt(n.Pos) != root.Add(sc.Semis[n.Degree]) {
						t.Errorf("%s %s pos %d: %v not degree %s", k, sc.ID, p, n.Pos, sc.Degrees[n.Degree])
					}
					if n.Pos.F < 0 {
						t.Errorf("negative fret in %s %s pos %d", k, sc.ID, p)
					}
				}
				if MaxFret(box)-minFret(box) > 6 {
					t.Errorf("%s %s pos %d spans too wide: %v", k, sc.ID, p, box)
				}
			}
		}
	}
}

func TestKeyChords(t *testing.T) {
	k := Key{MustNote("Bb")}
	var got []string
	for d := 0; d < 7; d++ {
		got = append(got, k.Chord(d).Name())
	}
	want := []string{"Bb", "Cm", "Dm", "Eb", "F", "Gm", "Adim"}
	if !slices.Equal(got, want) {
		t.Errorf("Bb chords = %v; want %v", got, want)
	}
}
