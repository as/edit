package edit

import (
	"fmt"
	"testing"
	"time"

	"github.com/as/text"
)

type tbl struct {
	in, prog, want string
}

var tabstop = []string{"A more common example\nis indenting a block of text\nby a tab stop.",
	"A more common example\n\tis indenting a block of text\n\tby a tab stop.",
}

func TestDot(t *testing.T) {
	ed, err := text.Open(text.NewBuffer())
	if err != nil {
		t.Fatal(err)
	}
	ed.Insert([]byte("one 1one1 2one2"), 0)
	ed.Select(4, 10)
	cmd, err := Compile(`x,n,c,z,`)
	if err != nil {
		t.Fatal(err)
	}
	cmd.Run(ed)
	want := "one 1oze1 2one2"
	have := string(ed.Bytes())
	if have != want {
		t.Fatalf("have: %q\nwant: %q\n", have, want)
	}

}

func TestInsertAppend(t *testing.T) {
	ed, err := text.Open(text.NewBuffer())
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range [...]*tbl{
		100: {"", `#1i,x,`, "x"},
		101: {"abc", `#1,#2i,Q,`, "aQbc"},
		102: {"abc", `#1,#2a,Q,`, "abQc"},
		103: {"abc", `,x,b,a,Q,`, "abQc"},
		200: {"abcdefg", `#1i,x,`, "axbcdefg"},
		201: {"abcdefg", `#1a,x,`, "axbcdefg"},
		202: {"abcdefg", `#1,#2a,x,`, "abxcdefg"},
		203: {"abcdefg", `#1,#2i,x,`, "axbcdefg"},
	} {
		if v == nil {
			continue
		}
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ed.Delete(0, ed.Len())
			ed.Insert([]byte(v.in), 0)
			ed.Select(0, 0)
			cmd, err := Compile(v.prog)
			if err != nil {
				t.Fatalf("failed: %s\n", err)
			}
			cmd.Run(ed)
			if s := string(ed.Bytes()); s != v.want {
				t.Fatalf("have: %q\nwant: %q\n", s, v.want)
			}

		})
	}
}

func TestExtractChange(t *testing.T) {
	b, err := text.Open(text.NewBuffer())
	w := b //text.Trace(b)
	if err != nil {
		t.Fatalf("failed: %s\n", err)
	}

	excerpt := `Here is an example. 
Say we wanted to double space the document.
That is.
Turn every newline
Into two newlines.
`

	want := `Here is an example. 

Say we wanted to double space the document.

That is.

Turn every newline

Into two newlines.

`
	want = want
	x := []tbl{
		{"Ralpha Rcmd Rdigit", ",x,R(alpha|digit),x,R,c,r,", "ralpha Rcmd rdigit"},
		{"visual studio", "/visual/c,crash,", "crash studio"},
		{"abc", `,x,b,a,Q,`, "abQc"},
		{"aca", `,x,a,a,b,`, "abcab"},
		{"abbcbbd", `,x,bb,i,Q,`, "aQbbcQbbd"},
		{"abbcbbd", `,x,bb,i,QQ,`, "aQQbbcQQbbd"}, // break
		{"abcbd", `,x,b,i,RR,`, "aRRbcRRbd"},
		{"abcbd", `,x,b,i,SSS,`, "aSSSbcSSSbd"},
		{"abcbd", `,x,b,i,T,`, "aTbcTbd"}, // break
		{"abcbd", `,x,b,a,U,`, "abUcbUd"}, // break
		{excerpt, `,x,\n,a,\n,`, want},
		{excerpt, `,x/\n/ i/\n/`, want},
		{excerpt, `,x/\n/ a/\n/`, want},
		{"", "", ""},
		{"", ",x,apple,d", ""},
		{"aaaaaaaaa", ",d", ""},
		{"a", "#0,#1d", ""},
		{"ab", "#0,#1d", "b"},
		{"abc", "#0,#1d", "bc"},
		{"abcd", "#0,#1i, ,", " abcd"},
		{"the gray fox", "#3a, quick,", "the quick gray fox"},
		{"the gray fax", "#4i, quick,", "the  quickgray fax"},
		{"he", "#2,a,y,", "hey"},
		{"he", "#0,i,t,", "the"},
		{"the quick brown peach", ",x,apple,d", "the quick brown peach"},
		{"the quick brown fox", ",x, ,d", "thequickbrownfox"},
		{"racecar car carrace", ",x,racecar,x,car,d", "race car carrace"},
		{"public static void func", ",y,func,d", "func"},
		{"ab aa ab aa", `,x,a.,g,aa,d`, "ab  ab "},
		{"ab aa ab aa", `,x,a.,v,aa,d`, " aa  aa"},
		{"generics debate", ",x,...,c,!@#,", "!@#!@#!@#!@#!@#"},
		{"programs are processes", "+#12 a, not,", "programs are not processes"},
		{"gnu EMACS", ",d", ""},
		{"considered harmful", "a,vim: ,", "vim: considered harmful"},
		{"................", ",x,....,x,..,x,.,i,@,", "@.@.@.@.@.@.@.@.@.@.@.@.@.@.@.@."},
		{"................", ",x,....,x,..,x,.,a,@,", ".@.@.@.@.@.@.@.@.@.@.@.@.@.@.@.@"},
		{"teh quark brown f", "0,1a,ox,", "teh quark brown fox"},
		{"nono", ",x,no,c,yes,", "yesyes"},
		{"f", "#1i,e,", "fe"},
		{"x", "#1a,y,", "xy"},
		{"how are you", ",y, ,x,.,c,x,", "xxx xxx xxx"},
		{"the quick the", ",y,quick,d", "quick"},
		{"aaaaaaa", ",x,...,d", "a"},
		{"the\nquick\nbrown\nfox", `^,1d`, "quick\nbrown\nfox"},
		{"the\nquick\nbrown\nfox", `^,2d`, "brown\nfox"},
		{"the\nquick\nbrown\nfox", `2,$d`, "the\n"},
		{"the\nquick\nbrown\nfox", `^,$d`, ""},
		{"the\nquick\nbrown\nfox", `^,#4d`, "quick\nbrown\nfox"},
		{"adefg", `^,+#1a@bc@,`, `abcdefg`},
		{"qrstuv", `$a,wxyz,`, `qrstuvwxyz`},
		{"qrstuv", `$a,wxyz,`, `qrstuvwxyz`},
		{"abbc", `,s/b/x/`, `axbc`},
		{"teh teh teh", `,s/teh/the/g`, `the the the`},
		{"Oh peter", `,s/peter/& & & & &/g`, `Oh peter peter peter peter peter`},
		{"They", `,s/ey/&&&&&/g`, `Theyeyeyeyey`},
		{excerpt, `,x/\n/ a/\n/`, want},
		{excerpt, `,x/\n/ c/\n\n/`, want},
		/*
			{excerpt, `,x/$/ a/\n/`, want},
			{excerpt, `,x/^/ i/\n/`, want},
		*/
		//		{tabstop[0], `,x/^/a/ /`,tabstop[1]},
		//		{tabstop[0], `,x/^/c/ /`,tabstop[1]},
		//		{tabstop[0], `,x/.*\n/i/ /`,tabstop[1]},
	}
	excerpt = excerpt
	done := make(chan bool)
	go func() {
		time.Sleep(time.Second * 5)
		select {
		case <-done:
		default:
			t.Fatal("timed out")
		}
	}()
	for _, v := range x {
		w.Delete(0, w.Len())
		w.Insert([]byte(v.in), 0)
		w.Select(0, 0)
		cmd, err := Compile(v.prog)
		if err != nil {
			t.Fatalf("failed: %s\n", err)
		}
		cmd.Run(w)
		if s := string(w.Bytes()); s != v.want {
			t.Fatalf("have: %q\nwant: %q\n", s, v.want)
		}
	}
	close(done)
}
