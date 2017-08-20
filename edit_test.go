package edit

import (
	"github.com/as/text"
	"testing"
)

type tbl struct {
	in, prog, want string
}

func TestExtractChange(t *testing.T) {
	w, err := text.Open(text.NewBuffer())
	if err != nil {
		t.Fatalf("failed: %s\n", err)
	}
	x := []tbl{
		{"", "", ""},
		{"",          ",x,apple,d", ""},
//		{"ab",        "#1d", "b"},
		{"c",         "#1a,b,", "cb"},
		{"c",         "#1i,b,", "cb"},
		{"abcd",      "#1i, ,", "a bcd"},
		{"the brown fox", "#3i, quick,", "the quick brown fox"},
//		{"he", "#2,a,y,", "hey"},
//		{"he", "#0,i,t,", "the"},
		{"the quick brown peach", ",x,apple,d", "the quick brown peach"},
		{"the quick brown fox", ",x, ,d", "thequickbrownfox"},
		{"racecar car carrace", ",x,racecar,x,car,d", "race car carrace"},
		{"public static void func", ",y,func,d", "func"},
		{"ab aa ab aa", `,x,a.,g,aa,d`, "ab  ab "},
		{"ab aa ab aa", `,x,a.,v,aa,d`, " aa  aa"},
		{"visual studio", "/visual/c,crash,", "crash studio"},
		{"generics debate", ",x,...,c,!@#,", "!@#!@#!@#!@#!@#"},
		{"programs are processes", "+#12 a, not,", "programs are not processes"},
		{"gnu EMACS", ",d", ""},
		{"considered harmful", "a,vim: ,", "vim: considered harmful"},
		{"................", ",x,....,x,..,x,.,i,@,", "@.@.@.@.@.@.@.@.@.@.@.@.@.@.@.@."},
		{"................", ",x,....,x,..,x,.,a,@,", ".@.@.@.@.@.@.@.@.@.@.@.@.@.@.@.@"},
		{"Ralpha Rcmd Rdigit", ",x,R(alpha|digit),x,R,c,r,", "ralpha Rcmd rdigit"},
	}
	for _, v := range x {
		w.Delete(0, 999999)
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
}
