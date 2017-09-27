package edit

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/as/event"
	"github.com/as/text"
	"github.com/as/worm"
	"io"
)

var (
	ErrNilFunc = errors.New("empty program")
	ErrNilEditor = errors.New("nil editor")
)

var (
	noop = func(ed text.Editor) {}
)

type Command struct {
	cor  *text.COR
	fn   func(text.Editor)
	s    string
	args string
	next *Command
}

func MustCompile(s string) (cmd *Command) {
	cmd, err := Compile(s)
	if err != nil {
		panic(fmt.Sprintf("MustCompile: %s\n", err))
	}
	return cmd
}

// Compile runs the build steps on the input string and returns
// a runnable command.
func Compile(s string) (cmd *Command, err error) {
	return cmdparse(s)
}

// Func returns a function entry point that operates on a text.Editor
func (c *Command) Func() func(text.Editor) {
	return c.fn
}
func interp(buf text.Buffer, rec event.Record, ins, del int) {
	switch t := rec.(type) {
	case *event.Write:
		fmt.Printf("%#v\n", t)
		switch buf := buf.(type) {
		case io.WriterAt:
			buf.WriteAt(t.P, t.Q0)
			return
		}
		buf.Delete(t.Q0, t.Q0+int64(len(t.P)))
		buf.Insert(t.P, t.Q0)
	case *event.Insert:
		fmt.Printf("%#v\n", t)
		buf.Insert(t.P, t.Q0)
	case *event.Delete:
		fmt.Printf("%#v\n", t)
		buf.Delete(t.Q0, t.Q1)
	}
}

// Run runs the compiled program on ed
func (c *Command) Run(ed text.Editor) (err error) {
	if ed == nil{
		return ErrNilEditor
	}
	if c.fn == nil {
		return ErrNilFunc
	}
	hist := worm.NewLogger()
	cor := text.NewCOR(ed, hist)
	defer cor.Close()
	q0,q1:=ed.Dot()
	cor.Select(q0,q1)
	c.fn(cor)
	cor.Flush()
	q0, q1 = cor.Dot()
	ed.Select(q0,q1)
	ins, del := int64(0), int64(0)
	buf := ed
	for i := int64(0); i < hist.Len(); i++ {
		t, _ := hist.ReadAt(int64(i))
		switch t := t.(type) {
		case *event.Insert:
			ins += t.Q1 - t.Q0
		case *event.Delete:
			del += t.Q1 - t.Q0
		}
		//log.Printf("%d: %#v\n", i, t)
	}
	ep := buf.Len()
	sp := ep
	isp := sp
	iep := buf.Len()
	if ins > del {
		buf.Insert(bytes.Repeat([]byte{0}, int(ins-del)), buf.Len())
	} else if del > ins {
		defer buf.Delete(buf.Len()-(del-ins), buf.Len())
	}
	
	for i := int64(hist.Len())- 1; i >= 0; i-- {
		e, err := hist.ReadAt(i)
		switch t := e.(type) {
		case *event.Write:
			q0 := t.Q0 + ins
			buf.(io.WriterAt).WriteAt(t.P, q0)
			iep -= t.Q1 - t.Q0
		case *event.Insert:
			isp = t.Q0
			if i == hist.Len()-1 {
				ep = t.Q1
			}
			if iep > isp {
				buf.(io.WriterAt).WriteAt(buf.Bytes()[isp:iep], isp+(ins)-del)
			}
			ins -= t.Q1 - t.Q0
			q0 := t.Q0 + ins - del
			iep = isp
			buf.(io.WriterAt).WriteAt(t.P, q0)
		case *event.Delete:
			sp = t.Q1
			if i == hist.Len()-1 {
				ep = buf.Len()
			}
			delta := ins - (t.Q1 - t.Q0)
			if ep > sp {
				buf.(io.WriterAt).WriteAt(buf.Bytes()[sp:ep], sp+delta)
			}
			del -= t.Q1 - t.Q0
			if del == 0 {
				ep = sp + delta
			}
		}
		if err != nil {
			break
		}
	}
	return err
}

// Next returns the next instruction for the compiled program. This
// effectively steps through x,..., and y,...,
func (c *Command) Next() *Command {
	return c.next
}

func (c *Command) nextFn() func(f text.Editor) {
	if c.next == nil {
		return nil
	}
	return c.next.fn
}

func compileAddr(a Address) func(f text.Editor) {
	if a == nil {
		return noop
	}
	return a.Set
}

func compile(p *parser) (cmd *Command) {
	for i := range p.cmd {
		if i+1 == len(p.cmd) {
			break
		}
		p.cmd[i].next = p.cmd[i+1]
	}
	fn := func(f text.Editor) {
		addr := compileAddr(p.addr)
		if addr != nil {
			addr(f)
		}
		if p.cmd != nil && p.cmd[0] != nil && p.cmd[0].fn != nil {
			p.cmd[0].fn(f)
		}
	}
	return &Command{fn: fn}
}

func cmdparse(s string) (cmd *Command, err error) {
	_, itemc := lex("cmd", s)
	p := parse(itemc)
	err = <-p.stop
	return compile(p), err
}
