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

var eprint = n

type parser struct {
	in        chan item
	out       chan func()
	last, tok item
	err       error
	stop      chan error
	cmd       []*Command
	addr      Address
	q         int64
}

func parse(i chan item) *parser {
	p := &parser{
		in:   i,
		stop: make(chan error),
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
		if rel != -1{
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
	//fmt.Printf("parseArg: %s\n", p.tok.value)
	p.Next()
	//fmt.Printf("parseArg: %s\n", p.tok.value)
	if p.tok.kind != kindArg {
		//		p.fatal(fmt.Errorf("not arg"))
	}
	return p.tok.value
}

func (p *parser) Dot(f text.Editor) (q0, q1 int64) {
	q0, q1 = f.Dot()
	//q0+= p.q
	//q1+= p.q
	return
}

// Put
func parseCmd(p *parser) (c *Command) {
	c = new(Command)
	v := p.tok.value
	//	fmt.Printf("parseCmd: %s\n", v)
	c.s = v
	switch v {
	case "a", "i":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			b := []byte(argv)
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
		c.fn = func(f text.Editor) {
			q0, q1 := p.Dot(f)
			f.Delete(q0, q1)
			f.Insert([]byte(argv), q0)
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
	case "s":
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
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			ok, err := regexp.Match(argv, f.Bytes()[q0:q1])
			if err != nil {
				panic(err)
			}
			if ok {
				if nextfn := c.nextFn(); nextfn != nil {
					nextfn(f)
				}
			}
		}
		return
	case "v":
		argv := parseArg(p)
		c.args = argv
		c.fn = func(f text.Editor) {
			q0, q1 := f.Dot()
			ok, err := regexp.Match(argv, f.Bytes()[q0:q1])
			if err != nil {
				panic(err)
			}
			if !ok {
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
		re, err := regexp.Compile(argv)
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
		re, err := regexp.Compile(argv)
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
