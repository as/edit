
package edit

import (
	"testing"
	"github.com/as/text"
)

type tbl struct{
	in, prog, want string
}

func TestExtractChange(t *testing.T){
	w, err := text.Open(text.NewBuffer())
	if err != nil{
		t.Fatalf("failed: %s\n", err)
	}
	x := []tbl{
	{"the quick brown fox", ",x, ,d", "thequickbrownfox"},
	{"racecar car carrace", ",x,racecar,x,car,d", "race car carrace"},
	{"public static void func", ",y,func,d", "func"},
	{"ab aa ab aa", `,x,a.,g,aa,d`, "ab  ab "},
	{"ab aa ab aa", `,x,a.,v,aa,d`, " aa  aa"},
	{"visual studio",  "/visual/c,crash,", "crash studio"},
	{"generics debate", ",x,...,c,!@#,", "!@#!@#!@#!@#!@#"},
	{"programs are processes", "+#12 a, not,", "programs are not processes"},
	{"gnu EMACS", ",d", ""},
	{"considered harmful", "a,vim: ,", "vim: considered harmful"},
	{"abcdefabcdefabcdefabcdef", ",x,....,x,..,x,.,i,@,", "@a@b@c@d@e@f@a@b@c@d@e@f@a@b@c@d@e@f@a@b@c@d@e@"},
	{"Ralpha Rcmd Rdigit", ",x,R(alpha|digit),x,R,c,r,", "ralpha Rcmd rdigit"},
	}
	for _, v := range x{
		w.Delete(0, 999999)
		w.Insert([]byte(v.in), 0)
		w.Select(0,0)
		cmd, err := Compile(v.prog)
		if err != nil{
			t.Fatalf("failed: %s\n", err)
		}
		cmd.Run(w)
		if s := string(w.Bytes()); s != v.want{
			t.Fatalf("have: %q\nwant: %q\n", s, v.want)
		}
	}
}

