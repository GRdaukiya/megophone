package megophone

/*
The original Metaphone algorithm was published in 1990 as an improvement over
the Soundex algorithm. Like Soundex, it was limited to English-only use. The
Metaphone algorithm does not produce phonetic representations of an input word
or name; rather, the output is an intentionally approximate phonetic
representation. The approximate encoding is necessary to account for the way
speakers vary their pronunciations and misspell or otherwise vary words and
names they are trying to spell.

The Double Metaphone phonetic encoding algorithm is the second generation of
the Metaphone algorithm. Its implementation was described in the June 2000
issue of C/C++ Users Journal. It makes a number of fundamental design
improvements over the original Metaphone algorithm.

It is called "Double" because it can return both a primary and a secondary code
for a string; this accounts for some ambiguous cases as well as for multiple
variants of surnames with common ancestry. For example, encoding the name
"Smith" yields a primary code of SM0 and a secondary code of XMT, while the
name "Schmidt" yields a primary code of XMT and a secondary code of SMT--both
have XMT in common.

Double Metaphone tries to account for myriad irregularities in English of
Slavic, Germanic, Celtic, Greek, French, Italian, Spanish, Chinese, and other
origin. Thus it uses a much more complex ruleset for coding than its
predecessor; for example, it tests for approximately 100 different contexts of
the use of the letter C alone.

This script implements the Double Metaphone algorithm (c) 1998, 1999 originally
implemented by Lawrence Philips in C++. It was further modified in C++ by Kevin
Atkinson (http {//aspell.net/metaphone/). It was translated to C by Maurice
Aubrey <maurice@hevanet.com> for use in a Perl extension. A Python version was
created by Andrew Collins on January 12, 2007, using the C source
(http {//www.atomodo.com/code/double-metaphone/metaphone.py/view). It was also
translated to Go by Adele Dewey-Lopez <adele@seed.co> using Atkinson's C++ source,
with some further revisions.

  Updated 2007-02-14 - Found a typo in the 'gh' section (0.1.1)
  Updated 2007-12-17 - Bugs fixed in 'S', 'Z', and 'J' sections (0.2;
                       Chris Leong)
  Updated 2009-03-05 - Various bug fixes against the reference C++
                       implementation (0.3; Matthew Somerville)
  Updated 2012-07    - Fixed long lines, added more docs, changed names,
                       reformulated as objects, fixed a bug in 'G'
                       (0.4; Duncan McGreggor)
  Updated 2013-06    - Enforced unicode literals (0.5; Ian Beaver)
  Updated 2015-05	 - // TODO: Write out all the changes I've made
*/

import (
	"fmt"
	"strings"
)

type phoneticData struct {
	word            []rune
	cur             int
	isSlavoGermanic bool
	metaphone1      string
	metaphone2      string
}

func (p *phoneticData) beginsWith(matches ...string) bool {
	start := strings.LastIndex(string(p.word)[:p.cur], " ")
	if start != -1 {
		for _, str := range matches {

			if strings.Contains(string(p.word)[start:p.cur], " "+str) {
				return true
			}
		}
	}
	return false
}

func (p *phoneticData) endsWith(matches ...string) bool {
	end := strings.Index(string(p.word)[p.cur:], " ")
	if end != -1 {
		for _, str := range matches {

			if strings.Contains(string(p.word)[p.cur:p.cur+end+1], str+" ") {
				return true
			}
		}
	}
	return false
}

func (p *phoneticData) containsAny(matches ...string) bool {
	for _, str := range matches {

		if strings.Contains(string(p.word), str) {
			return true
		}
	}
	return false
}

func (p *phoneticData) matchesAny(pos int, matches ...string) bool {
	if len(matches) == 0 {
		return true
	}
	// out of bounds
	if p.cur+pos < 0 {
		return false
	}

	for i, str := range matches {
		size := len(matches[i])
		// bounds check
		if p.cur+pos+size <= len(p.word) {
			if string(p.word[p.cur+pos:p.cur+size+pos]) == str {
				return true
			}
		}
	}

	return false
}

func (p *phoneticData) add(phoneme ...string) {
	if len(phoneme) > 0 {
		p.metaphone1 += phoneme[0]
		if len(phoneme) > 1 {
			p.metaphone2 += phoneme[1]
		} else {
			p.metaphone2 += phoneme[0]
		}
	}
}

func (p *phoneticData) skip(skipBy int) {
	p.cur += skipBy
}

func (p *phoneticData) isVowel(pos int) bool {
	return p.matchesAny(pos, "a", "e", "i", "o", "u", "y")
}

func (p *phoneticData) b() {
	p.add("p")
	// skip double b
	if p.matchesAny(1, "b") {
		p.skip(1)
	}
}

func (p *phoneticData) ç() {
	p.add("s")
}

func (p *phoneticData) c() {
	switch {
	case !p.matchesAny(-1, " ") && !p.matchesAny(-2, " ") && !p.isVowel(-2) &&
		p.matchesAny(-1, "ach") && !p.matchesAny(2, "i") &&
		(!p.matchesAny(2, "e") || p.matchesAny(-2, "acher")):
		// various germanic
		p.add("k")
		p.skip(1)
	case p.matchesAny(-1, " caesar"):
		// special case: "caesar"
		p.add("s")
		p.skip(1)
	case p.matchesAny(0, "chia"):
		// italian "chianti"
		p.add("k")
		p.skip(1)
	case p.matchesAny(0, "ch"):
		// ch
		switch {

		case !p.matchesAny(-1, " ") && p.matchesAny(0, "chae"):
			// find "michael"
			p.add("k")
		case !p.matchesAny(0, "chore") &&
			p.matchesAny(-1, " charac", " charis", " chor", " chym", " chia", " chem", " chla"):
			// greek roots
			p.add("k")
		case p.containsAny(" van ", " von ", " sch") || p.matchesAny(-3, "psych") ||
			p.matchesAny(-2, "orches", "archit", "orchid") || p.matchesAny(2, "t", "s") ||
			(p.matchesAny(-1, "a", "e", "o", "u", " ") &&
				p.matchesAny(2, "l", "r", "n", "m", "b", "h", "f", "v", "w", " ")):
			// germanic greek or otherwise "ch" for "kh" sound
			// "architect" but not "arch", "orchestra" or "orchid"
			// e.g., "watchler", "wechsler", but not "tichner"
			p.add("k")
		case !p.matchesAny(-1, " "):
			if p.matchesAny(-p.cur, " mc") {
				// e.g. "McHugh"
				if p.isVowel(2) {
					p.add("kh") //*
				} else {
					p.add("k")
				}
			} else {
				p.add("x", "k")
			}
		default:
			p.add("x")
		}
		p.skip(1)
	case p.matchesAny(0, "cz") && !p.matchesAny(-2, "wicz"):
		// e.g. "czerny"
		p.add("s", "x")
		p.skip(1)
	case p.matchesAny(1, "cia"):
		// e.g. "focaccia"
		p.add("x")
		p.skip(2)
	case p.matchesAny(0, "cc") && !p.matchesAny(-p.cur, " m"):
		// double "c", but not if e.g. "McClellan"
		if p.matchesAny(2, "i", "e", "h") && !p.matchesAny(2, "hu") {
			// "bellocchio" but not "bacchus"
			if p.matchesAny(-2, " a") || p.matchesAny(-1, "uccee", "ucces") {
				// "accident" "acceed" or "success"
				p.add("ks")
			} else {
				p.add("x")
			}
			p.skip(2)
		} else {
			// Pierce's rule
			p.add("k")
			p.skip(1)
		}
	case p.matchesAny(0, "ck", "cg", "cq"):
		p.add("k")
		p.skip(1)
	case p.matchesAny(0, "ci", "ce", "cy"):
		if p.matchesAny(0, "cio", "cie", "cia") {
			// italian vs. english
			p.add("s", "x")
		} else {
			p.add("s")
		}
		p.skip(1)
	default:
		p.add("k")
		// "mac caffrey", "mac gregor"
		if p.matchesAny(1, " c", " g", " q") {
			p.skip(2)
		} else if p.matchesAny(1, "c", "k", "q") && !p.matchesAny(1, "ce", "ci") {
			p.skip(1)
		}
	}
}

func (p *phoneticData) d() {
	switch {
	case p.matchesAny(0, "dg"):
		if p.matchesAny(2, "i", "e", "y") {
			p.add("j")
			p.skip(2)
		} else {
			p.add("tk")
			p.skip(1)
		}
	case p.matchesAny(0, "dd", "dt"):
		p.add("t")
		p.skip(1)
	default:
		p.add("t")
	}
}

func (p *phoneticData) f() {
	if p.matchesAny(0, "ff") {
		p.add("f")
		p.skip(1)
	} else {
		p.add("f")
	}
}

func (p *phoneticData) g() {
	switch {
	case p.matchesAny(1, "h"):
		switch {
		case !p.matchesAny(-1, " ") && !p.isVowel(-1):
			p.add("k")
			p.skip(1)
		case p.matchesAny(-1, " "):
			if p.matchesAny(2, "i") {
				p.add("j")
			} else {
				p.add("k")
			}
			p.skip(1)
		case p.matchesAny(-2, "b", "h", "d") ||
			p.matchesAny(-3, "b", "h", "d") ||
			p.matchesAny(-4, "b", "h", "d"):
			// Parker's rule (with further refinements)
			// e.g., "hugh", "bough", "broughton", "drought"
			p.skip(1)
		default:
			// e.g., "laugh", "McLaughlin"
			if p.matchesAny(-1, "u") && p.matchesAny(-3, "c", "g", "l", "r", "t") {
				p.add("f")
			} else if !p.matchesAny(-1, "i") {
				p.add("k")
			}
			p.skip(1)
		}
	case p.matchesAny(1, "n"):
		if p.matchesAny(-2, " ") && p.isVowel(-1) && !p.isSlavoGermanic {
			p.add("kn", "n")
		} else if p.matchesAny(-1, " ") {
			p.add("n")
		} else {
			// not e.g., "cagney"
			if !p.matchesAny(2, "ey") && !p.matchesAny(1, "y") && !p.isSlavoGermanic {
				p.add("n", "kn")
			} else {
				p.add("kn")
			}
		}
		p.skip(1)
	case p.matchesAny(1, "li") && !p.isSlavoGermanic:
		// tagliaro
		p.add("kl", "l")
		p.skip(1)
	case p.matchesAny(1, " gy") ||
		p.matchesAny(1, "es", "ep", "eb", "el", "ey", "ib", "il", "in", "ie", "ei", "er"):
		p.add("k", "j")
		p.skip(1)
	case p.matchesAny(1, "er", "y") &&
		// -ger- -gy-
		!p.matchesAny(-3, "danger", "ranger", "manger") &&
		!p.matchesAny(-1, "e", "i", "rgy", "ogy"):
		p.add("k", "j")
		p.skip(1)
	case p.matchesAny(1, "e", "i", "y") || p.matchesAny(-1, "aggi", "oggi"):
		// italian e.g. "biaggi"
		if p.containsAny(" van ", " von ", " sch") || p.matchesAny(1, "et") {
			// obvious germanic
			p.add("k")
		} else if p.matchesAny(1, "ier") {
			// always soft if french ending
			p.add("j")
		} else {
			p.add("j", "k")
		}
		p.skip(1)
	default:
		p.add("k")
		if p.matchesAny(1, "g") {
			p.skip(1)
		}
	}
}

func (p *phoneticData) h() {
	if (p.matchesAny(-1, " ") || p.isVowel(-1)) && p.isVowel(1) {
		// only keep if first and before vowel, or between two vowels
		p.add("h")
		p.skip(1)
	}
	// will skip over "hh"
}

func (p *phoneticData) j() {
	switch {
	case p.matchesAny(0, "jose") || p.containsAny(" san "):
		// obvious spanish e.g. "jose" "san jacinto"
		if (p.matchesAny(-1, " ") && p.matchesAny(4, " ")) || p.containsAny(" san ") {
			p.add("h")
		} else {
			p.add("j", "h")
		}
	case p.matchesAny(-1, " "):
		p.add("j", "a")
	case p.isVowel(-1) && !p.isSlavoGermanic && p.matchesAny(1, "a", "o"):
		p.add("j", "h")
	case p.matchesAny(1, " "):
		// end of the word because of padding
		p.add("j", "")
	case !p.matchesAny(1, "l", "t", "k", "s", "n", "m", "b", "z") &&
		!p.matchesAny(-1, "s", "k", "l"):
		p.add("j")
	}

	if p.matchesAny(1, "j") {
		p.skip(1)
	}
}

func (p *phoneticData) k() {
	if !p.matchesAny(-1, " kn") {
		p.add("k")
	}
	if p.matchesAny(1, "k") {
		p.skip(1)
	}
}

func (p *phoneticData) l() {
	if p.matchesAny(1, "l") {
		if p.matchesAny(-1, "illo ", "illa ", "alle ") ||
			p.endsWith("as", "os", "a", "o") &&
				p.matchesAny(-1, "alle") {
			p.add("l", "")
		} else {
			p.add("l")
		}
		p.skip(1)
	} else if !p.matchesAny(-2, "colm", "coln") { //*
		// "malcolm", "lincoln"
		p.add("l")
	}
}

func (p *phoneticData) m() {
	switch {
	case p.matchesAny(-1, "umb") && (p.matchesAny(2, " ") || p.matchesAny(2, "er")):
		p.add("m", "mp")
		p.skip(1)
	case p.matchesAny(0, "mn "): //*
		p.skip(1)
	case !p.matchesAny(-1, " mn"): //*
		p.add("m")
		if p.matchesAny(1, "m") {
			p.skip(1)
		}
	}
}

func (p *phoneticData) n() {
	p.add("n")
	if p.matchesAny(1, "n") {
		p.skip(1)
	}
}

func (p *phoneticData) ñ() {
	p.add("n")
}

func (p *phoneticData) p() {
	switch {
	case p.matchesAny(0, "ph"):
		if p.matchesAny(-p.cur, " uph", " sheph", " haph") {
			// "shepherd" "uphill" "haphazard"
			p.add("p")
		} else {
			p.add("f")
			p.skip(1)
		}
	case p.matchesAny(1, "b", "p"):
		p.add("p")
		p.skip(1)
	case !p.matchesAny(-1, " pn", " ps"):
		p.add("p")
	}
}

func (p *phoneticData) q() {
	if p.matchesAny(1, "q") {
		p.skip(1)
	}
	p.add("k")
}

func (p *phoneticData) r() {
	if p.matchesAny(0, "r ") && !p.isSlavoGermanic &&
		p.matchesAny(-2, "ie") && !p.matchesAny(-4, "me", "ma") {
		p.add("", "r")
	} else {
		p.add("r")
	}

	if p.matchesAny(1, "r") {
		p.skip(1)
	}
}

func (p *phoneticData) s() {
	switch {
	case p.matchesAny(-1, "isl", "ysl"):
		// special cases: "island" "carlysle"
		// skip it
	case p.matchesAny(-1, " sugar"):
		p.add("x", "s")
	case p.matchesAny(0, "sh"):
		if p.matchesAny(1, "holm", "holz", "heim", "hoek") {
			p.add("s")
		} else {
			p.add("x")
		}
		p.skip(1)
	case p.matchesAny(0, "sio", "sia"):
		if !p.isSlavoGermanic {
			p.add("s", "x")
		} else {
			p.add("s")
		}
		p.skip(2)
	case p.matchesAny(-1, " sm", " sn", " sl", " sw") || p.matchesAny(1, "z"):
		p.add("s", "x")
		if p.matchesAny(1, "z") {
			p.skip(1)
		}
	case p.matchesAny(0, "sc"):
		// Schlesinger's rule
		switch {
		case p.matchesAny(2, "h"):
			if p.matchesAny(3, "oo", "er", "en", "uy", "ed", "em") {
				// dutch origin
				if p.matchesAny(3, "er", "en") {
					p.add("x", "sk")
				} else {
					p.add("sk")
				}

			} else if p.matchesAny(-1, " ") && !p.isVowel(3) && !p.matchesAny(3, "w") {
				p.add("x", "s")
			} else {
				p.add("x")
			}
			p.skip(2)
		case p.matchesAny(2, "i", "e", "y"):
			p.add("s")
			p.skip(2)
		default:
			p.add("sk")
			p.skip(1) // *
		}
	case p.matchesAny(-2, "ais ", "ois ", "uis "): // *
		p.add("", "s")
	default:
		p.add("s")
		if p.matchesAny(1, "s", "z") {
			p.skip(1)
		}
	}
}

func (p *phoneticData) t() {
	switch {
	case p.matchesAny(0, "tion"):
		p.add("x")
		p.skip(2)
	case p.matchesAny(0, "tia", "tch"):
		p.add("x")
		p.skip(2)
	case p.matchesAny(0, "th", "tth"):
		if p.matchesAny(2, "om", "am") || p.containsAny(" van ", " von ", " sch") {
			p.add("t")
		} else {
			p.add("0", "t")
		}
		p.skip(1)
	default:
		p.add("t")
		if p.matchesAny(1, "t", "d") {
			p.skip(1)
		}
	}
}

func (p *phoneticData) v() {
	if p.matchesAny(-1, " ") { // *
		p.add("f", "f")
	} else {
		p.add("f")
	}
	if p.matchesAny(1, "v") {
		p.skip(1)
	}
}

func (p *phoneticData) w() {
	switch {
	case p.matchesAny(0, "wr"):
		p.add("r")
		p.skip(1)
	case p.matchesAny(-1, " ") && p.isVowel(1) || p.matchesAny(0, "wh"):
		if p.isVowel(1) {
			p.add("a", "f")
		} else {
			p.add("a")
		}
	case p.matchesAny(1, " ") && p.isVowel(-1) || p.matchesAny(-p.cur, " sch") ||
		p.matchesAny(-1, "ewski", "owski", "ewsky", "owsky"):
		p.add("", "f")
	case p.matchesAny(0, "witz", "wicz"):
		p.add("ts", "fx")
		p.skip(3)
	}
	// otherwise do nothing
}

func (p *phoneticData) x() {
	switch {
	case p.matchesAny(-1, " "):
		p.add("s")
	case !(p.matchesAny(-2, "aux ", "oux ")): // *
		p.add("ks")
	}

	if p.matchesAny(1, "x", "ce", "ci", "cy") {
		p.skip(1)
	}
}

func (p *phoneticData) z() {
	switch {
	case p.matchesAny(1, "h"):
		p.add("j")
		p.skip(1)
	case p.matchesAny(1, "zo", "zi", "za") || (p.isSlavoGermanic && !p.matchesAny(-1, "t")):
		p.add("s", "ts")
	default:
		p.add("s")
	}

	if p.matchesAny(1, "z") {
		p.skip(1)
	}
}

func Metaphone(s string) (string, string) {
	// initialize
	var p *phoneticData
	p = &phoneticData{}

	// pad string
	// normalize
	p.word = []rune(" " + strings.ToLower(s) + " ")

	if p.containsAny("w", "k", "cz", "witz") {
		p.isSlavoGermanic = true
	}

	for i, next := range p.word {
		if p.cur == i {
			//fmt.Println(p.cur, ": ", string(next))
			switch next {
			case 'a', 'e', 'i', 'o', 'u', 'y':
				if p.matchesAny(-1, " ") {
					p.add("a")
				}
			case 'b':
				p.b()
			case 'ç':
				p.ç()
			case 'c':
				p.c()
			case 'd':
				p.d()
			case 'f':
				p.f()
			case 'g':
				p.g()
			case 'h':
				p.h()
			case 'j':
				p.j()
			case 'k':
				p.k()
			case 'l':
				p.l()
			case 'm':
				p.m()
			case 'n':
				p.n()
			case 'ñ':
				p.ñ()
			case 'p':
				p.p()
			case 'q':
				p.q()
			case 'r':
				p.r()
			case 's':
				p.s()
			case 't':
				p.t()
			case 'v':
				p.v()
			case 'w':
				p.w()
			case 'x':
				p.x()
			case 'z':
				p.z()
			}
			p.cur++
		}
	}

	return p.metaphone1, p.metaphone2
}

func printMetaphonePair(s string) {
	// just for testing
	p1, p2 := Metaphone(s)
	// stupid hacks so they all line up nicely
	p2 = "\tSecond: " + p2
	s = "\tOriginal: " + s
	if len(p2) < 16 {
		s = "\t" + s
	}
	if len(p1) < 7 {
		p2 = "\t" + p2
	}

	fmt.Println("First: ", p1, p2, s)
}
