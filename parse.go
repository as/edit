package edit

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

import (
	"github.com/as/text"
)

type Print string

var eprint = n

type Emitted struct {
	Name string
	Dot  []Dot
}

type parser struct {
	in        chan item
	out       chan func()
	last, tok item
	err       error
	stop      chan error
	cmd       []*Command
	addr      Address
	q         int64
	
	recache map[string]*regexp.Regexp

	Emit    *Emitted
	Options *Options
}

func (p *parser) compileRegexp(s string) (re *regexp.Regexp, err error){
	if p.recache == nil{
		p.recache = make(map[string]*regexp.Regexp)
	}
	re, ok := p.recache[s]
	if !ok{
		re, err = regexp.Compile(s)
		if err != nil{
			return nil, err
		}
		p.recache[s] = re
	}
	return re, nil
}

func parse(i chan item, opts ...*Options) *parser {
	var o *Options
	if len(opts) != 0 {
		o = opts[0]
	}
	p := &parser{
		in:      i,
		stop:    make(chan error),
		Emit: &Emitted{},
		Options: o,
	}
	go p.run()
	return p
}

// Put
func parseAddr(p *parser) (a Address) {
	a0 := parseSimpleAddr(p)
	p.Next()
	op, a1 := parseOp(p)
	if op == '\x00' {
		return a0
	}
	p.Next()
	return &Compound{a0: a0, a1: a1, op: op}
}

func parseOp(p *parser) (op byte, a Address) {
	//	fmt.Printf("parseOp:1 %s\n", p.tok)
	if p.tok.kind != kindOp {
		return
	}
	v := p.tok.value
	if v == "" {
		eprint("no value" + v)
		return
	}
	if strings.IndexAny(v, "+-;,") == -1 {
		//		eprint(fmt.Sprintf("bad op: %q", v))
	}
	p.Next()
	return v[0], parseSimpleAddr(p)
}

func tryRelative(p *parser) int {
	v := p.tok.value
	k := p.tok
	if k.kind == kindRel {
		defer p.Next()
		if v == "+" {
			return 1
		}
		return -1
	}
	return 0
}

// Put
func parseSimpleAddr(p *parser) (a Address) {
	//fmt.Printf("parseSimpleAddr:1 %s\n", p.tok)
	back := false
	rel := tryRelative(p)
	v := p.tok.value
	k := p.tok
	//fmt.Printf("%s\n", k)
	switch k.kind {
	case kindRegexpBack:
		back = true
		fallthrough
	case kindRegexp:
		re, err := regexp.Compile(v)
		if err != nil {
			p.fatal(err)
			return
		}
		if rel != -1 {
			rel = 1
		}
		return &Regexp{re, back, 1}
	case kindLineOffset, kindByteOffset:
		i := p.mustatoi(v)
		if rel < 0 {
			i = -i
		}
		if k.kind == kindLineOffset {
			return &Line{i, rel}
		}
		return &Byte{i, rel}
	case kindDot:
		return &Dot{}
	}
	p.err = fmt.Errorf("bad address: %q", v)
	return
}

func parseArg(p *parser) (arg string) {
	p.Next()
	if p.tok.kind != kindArg {
		p.fatal(fmt.Errorf("want arg, have %q", p.tok.value))
	}
	return p.tok.value
}

func (p *parser) Dot(f text.Editor) (q0, q1 int64) {
	q0, q1 = f.Dot()
	//q0+= p.q
	//q1+= p.q
	return
}

type Sender interface {
	Send(e interface{})
	SendFirst(e interface{})
}

// Put
func parseCmd(p *parser) (c *Command) {
	v := p.tok.value
	//	fmt.Printf("parseCmd: %s\n", v)
	c = &Command{}
	c.s = v
	switch v {
	case "h":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			p.Emit.Dot = append(p.Emit.Dot, Dot{q0, q1})
		}
		return
	case "=":
		argv := parseArg(p)
		c.args = argv
			
			if p.Options == nil || p.Options.Sender == nil {
				return
			}
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			str := fmt.Sprintf("%s:#%d,#%d", p.Options.Origin, q0+1, q1)
			p.Options.Sender.Send(Print(str))
		}
		return
	case "p":
			if p.Options == nil || p.Options.Sender == nil {
				return
			}
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			str := fmt.Sprintf("%s", f.Bytes()[q0:q1])
			p.Options.Sender.Send(Print(str))
		}
		return
	case "a", "i":
		argv := parseArg(p)
		c.args = argv
		b := []byte(argv)
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			if v == "i" {
				f.Insert(b, q0)
			} else {
				f.Insert(b, q1)
			}
			p.q += int64(len(argv))
		}
		return
	case "c":
		argv := parseArg(p)
		c.args = argv
		b := []byte(argv)
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			f.Delete(q0, q1)
			f.Insert(b, q0)
			p.q += int64(len(argv))
			p.q -= q1 - q0
		}
		return
	case "d":
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			f.Delete(q0, q1)
			p.q -= q1 - q0
		}
		return
	case "e":
	case "k":
	case "r":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			data, err := ioutil.ReadFile(c.args)
			if err != nil {
				eprint(err)
				return
			}
			q0, q1 := f.Dot()
			if q0 != q1 {
				f.Delete(q0, q1)
			}
			f.Insert(data, q0)
		}
		return
	case "s":
		matchn := int64(1)
		sre := ""
		parseArg(p)
		a1 := p.tok
		if a1.kind == kindCount {
			matchn = p.mustatoi(a1.value)
			parseArg(p)
			a2 := p.tok
			sre = a2.value
		} else {
			sre = a1.value
		}
		parseArg(p)
		a3 := p.tok
		replProg := compileReplaceAmp(a3.value)

		// And at this point I realized that instead
		// of a one token look-ahead parser, I have
		// a one token look-behind parser. How unfortunate.
		//
		// TODO(as): check for 'g' here after fixing the parser
		// try parsing the last part of the construction anyway
		// and look for 'g'
		parseArg(p)
		if p.tok.kind == kindGlobal {
			if p.tok.value != "g" {
				p.fatal("s: suffix not supported: " + p.tok.value)
				return
			}
			matchn = -1
		}
		if sre == "" {
			eprint("s: no regexp to find")
			return
		}

		re, err := regexp.Compile(sre)
		if err != nil {
			p.fatal(err)
			return
		}
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			x0, x1 := int64(0), int64(0)

			buf := bytes.NewReader(f.Bytes()[q0:q1])
			for i := int64(1); ; i++ {
				loc := re.FindReaderIndex(buf)
				if loc == nil {
					buf.Seek(x1, 0)
					eprint("not found")
					break
				}
				x0, x1 = int64(loc[0])+x1, int64(loc[1])+x1
				f.Select(q0+x0, q0+x1)

				if i == matchn || matchn == -1 {
					q0, q1 := p.Dot(f)
					buf := replProg.Gen(f.Bytes()[q0:q1])
					f.Delete(q0, q1)
					f.Insert(buf, q0)
					p.q -= q1 - q0
					p.q += int64(len(buf))
				}

				buf.Seek(x1, 0)
			}
			f.Select(q0, q1)
		}
		return

	case "w":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			fd, err := os.Create(argv)
			if err != nil {
				eprint(err)
				return
			}
			defer fd.Close()
			q0, q1 := f.Dot()
			_, err = io.Copy(fd, bytes.NewReader(f.Bytes()[q0:q1]))
			if err != nil {
				eprint(err)
			}
		}
		return
	case "m":
		a1 := parseSimpleAddr(p)
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			p := append([]byte{}, f.Bytes()[q0:q1]...)
			a1.Set(f)
			_, a1 := f.Dot()
			f.Delete(q0, q1)
			f.Insert(p, a1)
		}
		return
	case "t":
		a1 := parseSimpleAddr(p)
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			p := f.Bytes()[q0:q1]
			a1.Set(f)
			_, a1 := f.Dot()
			f.Insert(p, a1)
		}
		return
	case "g":
		argv := parseArg(p)
		c.args = argv
		re, err :=  regexp.Compile(argv) //p.compileRegexp(argv) 
		if err != nil {
			p.fatal(err)
			return
		}
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			if re.Match(f.Bytes()[q0:q1]) {
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
			}
		}
		return
	case "v":
		argv := parseArg(p)
		c.args = argv
		re, err :=  regexp.Compile(argv) //p.compileRegexp(argv) 
		if err != nil {
			p.fatal(err)
			return
		}
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			if !re.Match(f.Bytes()[q0:q1]) {
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
			}
		}
		return
	case "|":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			x := strings.Fields(argv)
			if len(x) == 0 {
				eprint("|: nothing on rhs")
			}
			n := x[0]
			var a []string
			if len(x) > 1 {
				a = x[1:]
			}
			q0, q1 := f.Dot()
			cmd := exec.Command(n, a...)
			cmd.Stdin = bytes.NewReader(append([]byte{}, f.Bytes()[q0:q1]...))
			buf := new(bytes.Buffer)
			cmd.Stdout = buf
			err := cmd.Run()
			if err != nil {
				eprint(err)
			}
			f.Delete(q0, q1)
			f.Insert(buf.Bytes(), q0)

		}
		return
	case ">":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			fd, err := os.Create(argv)
			if err != nil {
				eprint(err)
				return
			}
			defer fd.Close()
			q0, q1 := f.Dot()
			_, err = io.Copy(fd, bytes.NewReader(f.Bytes()[q0:q1]))
			if err != nil {
				eprint(err)
			}
		}
		return
	case "x":
		argv := parseArg(p)
		c.args = argv
		re, err :=  regexp.Compile(argv) //p.compileRegexp(argv) 
		if err != nil {
			p.fatal(err)
			return
		}
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			x0, x1 := int64(0), int64(0)

			buf := bytes.NewReader(f.Bytes()[q0:q1])
			for {
				loc := re.FindReaderIndex(buf)
				if loc == nil {
					buf.Seek(x1, 0)
					eprint("not found")
					break
				}
				x0, x1 = int64(loc[0])+x1, int64(loc[1])+x1

				f.Select(q0+x0, q0+x1)
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
				buf.Seek(x1, 0)
			}
			f.Select(q0, q1)
		}
		return
	case "y":
		argv := parseArg(p)
		c.args = argv
		re, err := regexp.Compile(argv) //p.compileRegexp(argv)
		if err != nil {
			p.fatal(err)
			return
		}
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			x0, x1 := int64(0), int64(0)
			y0, y1 := int64(0), q1
			buf := bytes.NewReader(f.Bytes()[q0:q1])
			for {
				loc := re.FindReaderIndex(buf)
				if loc == nil {
					buf.Seek(x1, 0)
					eprint("not found")
					break
				}
				y0 = x1
				x0, x1 = int64(loc[0])+x1, int64(loc[1])+x1
				y1 = x0
				f.Select(q0+y0, q0+y1)
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
				buf.Seek(x1, 0)
			}
			if x1 != q1 {
				f.Select(q0+x1, q1)
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
			}
		}
		return
	}
	return nil
}

type ReplaceAmp []func([]byte) string

func (r ReplaceAmp) Run(ed text.Editor, q1 int64, sel []byte) (n int) {
	for _, fn := range r {
		b := []byte(fn(sel))
		n += len(b)
		ed.Insert(b, q1)
	}
	return n
}
func (r ReplaceAmp) Gen(replace []byte) (b []byte) {
	for _, fn := range r {
		b = append(b, fn(replace)...)
	}
	return b
}
func compileReplaceAmp(in string) (s ReplaceAmp) {
	// we can use strings.Map but then invalid runes
	// are replaced. we'll do it the old fashioned
	// way for now

	// strings are immutable, but this is faster than
	// the string 'builders' on platforms tested
	for {
		i := strings.Index(in, `&`)
		if i == -1 {
			s = append(s, func([]byte) string { return in })
			return
		}
		if i > 0 && in[i] == '\\' {
			s = append(s, func(b []byte) string { return in[:i-2] })
			s = append(s, func(b []byte) string { return "&" })
			i += 2
		} else {
			s = append(s, func(b []byte) string { return in[:i-1] })
			s = append(s, func(b []byte) string { return string(b) })
			i++
		}
		if i == len(in) {
			break
		}
		in = in[i:]
	}
	if s == nil {
		s = append(s, func(b []byte) string { return "" })
	}
	return
}
func (p *parser) Next() *item {
	p.last = p.tok
	p.tok = <-p.in
	return &p.tok
}

func (p *parser) run() {
	tok := p.Next()
	if tok.kind == kindEof || p.err != nil {
		if tok.kind == kindEof {
			p.fatal(fmt.Errorf("run: unexpected eof"))
			return
		}
		p.fatal(fmt.Errorf("run: %s", p.err))
		return
	}
	p.addr = parseAddr(p)
	for {
		c := parseCmd(p)
		if c == nil {
			break
		}
		p.cmd = append(p.cmd, c)
		eprint(fmt.Sprintf("(%s) %#v and cmd is %#v\n", tok, p.addr, c))
		p.Next()
	}
	p.stop <- p.err
	close(p.stop)
}

func (p *parser) mustatoi(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		p.fatal(err)
	}
	return int64(i)
}
func (p *parser) fatal(why interface{}) {
	switch why := why.(type) {
	default:
		//fmt.Println(why)
		_ = why
	}
}

func n(i ...interface{}) (n int, err error) {
	return
}
