package edit

import (
	"testing"

	"github.com/as/text"
)

type tbl struct {
	in, prog, want string
}

func TestExtractChange(t *testing.T) {
	b, err := text.Open(text.NewBuffer())
	w := b //text.Trace(b)
	if err != nil {
		t.Fatalf("failed: %s\n", err)
	}
	x := []tbl{
		{"", "", ""},
		{"", ",x,apple,d", ""},
		{"aaaaaaaaa", ",d", ""},
		{"a", "#1d", ""},
		{"ab", "#1d", "b"},
		{"abc", "#1d", "bc"},
		{"abcd", "#1i, ,", " abcd"},
		{"the gray fox", "#3a, quick,", "the quick gray fox"},
		{"the gray fox", "#4i, quick,", "the quick gray fox"},
		{"he", "#2,a,y,", "hey"},
		{"he", "#0,i,t,", "the"},
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
		{"teh quark brown f", "0,1a,ox,", "teh quark brown fox"},
		{"nono", ",x,no,c,yes,", "yesyes"},
		{"f", "#1i,e,", "ef"},
		{"x", "#1a,y,", "xy"},
		{"how are you", ",y, ,x,.,c,x,", "xxx xxx xxx"},
		{"the quick the", ",y,quick,d", "quick"},
		{"aaaaaaa", ",x,...,d", "a"},
		{"the\nquick\nbrown\nfox", `^,1d`, "quick\nbrown\nfox"},
		{"the\nquick\nbrown\nfox", `^,2d`, "brown\nfox"},
		{"the\nquick\nbrown\nfox", `2,$d`, "the\n"},
		{"the\nquick\nbrown\nfox", `^,$d`, ""},
		{"the\nquick\nbrown\nfox", `^,#4d`, "quick\nbrown\nfox"},
		//{"12", "0,1|xc", "3132"},
		//{"123", ",1|xc", "313233"},
		//{"123", ",|xc", "313233"},
	}

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
	y := []tbl{
		{"", "", ">"},
		{"", ",x,apple,d", ">"},
		//		{"ab",        "#1d", ">b"},
		{"c", "#1a,b,", ">bc"},
		{"c", "#1i,b,", "b>c"},
		{"abcd", "#1i, ,", " >abcd"},
		{"the brown fox", "#4i, quick,", ">th quicke brown fox"},
		//		{"he", "#2,a,y,", "hey"},
		//		{"he", "#0,i,t,", "the"},
		{"the quick brown peach", ",x,apple,d", ">the quick brown peach"},
		{"the quick brown fox", ",x, ,d", ">thequickbrownfox"},
		{"racecar car carrace", ",x,racecar,x,car,d", ">race car carrace"},
		{"public static void func", ",y,func,d", "func"},
		{"ab aa ab aa", `,x,a.,g,aa,d`, ">ab  ab "},
		{"ab aa ab aa", `,x,a.,v,aa,d`, "> aa  aa"},
		{"visual studio", "/visual/c,crash,", ">crash studio"},
		{"generics debate", ",x,...,c,!@#,", "!@#!@#!@#!@#!@#e"},
		{"programs are processes", "+#12 a, not,", ">programs ar note processes"},
		{"gnu EMACS", ",d", ""},
		{"considered harmful", "a,vim: ,", "vim: >considered harmful"},
	}
	for _, v := range y {
		w.Delete(0, w.Len())
		w.Insert([]byte(v.in), 0)
		w.Insert([]byte{'>'}, 0)
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
