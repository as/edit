package edit

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/as/event"
	"github.com/as/text"
	"github.com/as/worm"
)

var (
	ErrNilFunc   = errors.New("empty program")
	ErrNilEditor = errors.New("nil editor")
)

var (
	noop = func(ed Editor) {}
)

type Options struct {
	Sender Sender
	Origin string
}

type Command struct {
	fn       func(Editor)
	s        string
	args     string
	next     *Command
	Emit     *Emitted
	modified bool
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
func Compile(s string, opts ...*Options) (cmd *Command, err error) {
	_, itemc := lex("cmd", s)
	p := parse(itemc, opts...)
	err = <-p.stop
	return compile(p), err
}

// Modified returns true if the last call to c.Run() modified the contents
// of the editor
func (c *Command) Modified() bool {
	return c.modified
}

// Func returns a function entry point that operates on a Editor
func (c *Command) Func() func(Editor) {
	return c.fn
}
func interp(buf text.Buffer, rec event.Record, ins, del int) {
	switch t := rec.(type) {
	case *event.Write:
		switch buf := buf.(type) {
		case io.WriterAt:
			buf.WriteAt(t.P, t.Q0)
			return
		}
		buf.Delete(t.Q0, t.Q0+int64(len(t.P)))
		buf.Insert(t.P, t.Q0)
	case *event.Insert:
		buf.Insert(t.P, t.Q0)
	case *event.Delete:
		buf.Delete(t.Q0, t.Q1)
	}
}

func net(hist worm.Logger) (ins, del int64) {
	for i := int64(0); i < hist.Len(); i++ {
		t, _ := hist.ReadAt(int64(i))
		switch t := t.(type) {
		case *event.Insert:
			ins += t.Q1 - t.Q0
		case *event.Delete:
			del += t.Q1 - t.Q0
		}
	}
	return
}

type editWriter interface {
	io.WriterAt
	Editor
}

func newEditWriter(ed Editor) editWriter {
	switch ed := ed.(type) {
	case editWriter:
		return ed
	}
	return &ew{Editor: ed}
}

type ew struct {
	Editor
}

func (e *ew) WriteAt(p []byte, off int64) (n int, err error) {
	panic("not implemented")
	//	q0 := off
	//	q1 := off + int64(len(p))
	//	if q1 > e.Len() {
	//		q1 = e.Len()
	//	}
	//	e.Delete(q0, q1)
	//	return e.Insert(p, q0), nil
}

// Commit plays back the history onto ed, starting from
// the last event to the first in reverse order. This is
// useful only when hist contains a set of independent events
// applied as a transaction where shifts in the address offsets
// are not observed.
//
// If the command is
//		a,abc,
//		x,.,a,Q,
// The result is:
//		abc -> aQbQcQ
// The log should contain
// 		i 1 Q
// 		i 2 Q	(not i 3 Q)
// 		i 3 Q (not i 5 Q)
//
// Commit will only reallocate ed's size once. If ed implements
// io.WriterAt, a write-through fast path is used to commit the
// transaction.
func Commit(ed Editor, hist worm.Logger) (err error) {
	buf := newEditWriter(ed)
	ins, del := net(hist)
	ep := buf.Len()
	sp, isp, iep := ep, ep, ep

	if last := hist.Len() - 1; last >= 0 {
		e, _ := hist.ReadAt(last)
		q0, q1 := int64(0), int64(0)
		switch t := e.(type) {
		case *event.Write:
			q0, q1 = t.Q0, t.Q1
		case *event.Insert:
			q0, q1 = t.Q0, t.Q1
		case *event.Delete:
			q0, q1 = t.Q0, t.Q0
		}
		delta := ins - del
		defer buf.Select(q0+delta, q1+delta)
	}

	if ins > del {
		buf.Insert(bytes.Repeat([]byte{0}, int(ins-del)), buf.Len())
	} else if del > ins {
		defer buf.Delete(buf.Len()-(del-ins), buf.Len())
	}

	for i := int64(hist.Len()) - 1; i >= 0; i-- {
		e, err := hist.ReadAt(i)
		switch t := e.(type) {
		case *event.Write:
			q0 := t.Q0 + ins
			buf.WriteAt(t.P, q0)
			iep -= t.Q1 - t.Q0
		case *event.Insert:
			isp = t.Q0
			if i == hist.Len()-1 {
				ep = t.Q1
			}
			if iep > isp {
				buf.WriteAt(buf.Bytes()[isp:iep], isp+(ins)-del)
			}

			ins -= t.Q1 - t.Q0
			q0 := t.Q0 + ins - del
			iep = isp

			buf.WriteAt(t.P, q0)

		case *event.Delete:
			sp = t.Q1
			if i == hist.Len()-1 {
				ep = buf.Len()
			}
			delta := ins - (t.Q1 - t.Q0)
			if ep > sp {
				buf.WriteAt(buf.Bytes()[sp:ep], sp+delta)
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

func (c *Command) ck(ed Editor) error {
	c.modified = false
	if ed == nil {
		return ErrNilEditor
	}
	if c.fn == nil {
		return ErrNilFunc
	}
	return nil
}

// Transcribe runs the compiled program on ed
func (c *Command) Transcribe(ed Editor) (log worm.Logger, err error) {
	if err = c.ck(ed); err != nil {
		return nil, err
	}
	log = worm.NewLogger()
	cor := text.NewCOR(ed, log)
	defer cor.Close()

	q0, q1 := ed.Dot()
	cor.Select(q0, q1)

	c.Emit.Dot = c.Emit.Dot[:0]
	c.fn(cor)
	cor.Flush()

	q0, q1 = cor.Dot()
	ed.Select(q0, q1)
	return log, nil
}

// Run runs the compiled program on ed
func (c *Command) RunTransaction(ed Editor) (err error) {
	hist, err := c.Transcribe(ed)
	if err != nil {
		return err
	}
	c.modified = hist.Len() > 0
	return Commit(ed, hist)
}

// Run runs the compiled program on ed
func (c *Command) Run(ed Editor) (err error) {
	return c.RunTransaction(ed)
}

func (c *Command) oldRun(ed Editor) (err error) {
	if err = c.ck(ed); err != nil {
		return err
	}
	c.Emit.Dot = c.Emit.Dot[:0]
	c.fn(ed)
	return nil
}

// Next returns the next instruction for the compiled program. This
// effectively steps through x,..., and y,...,
func (c *Command) Next() *Command {
	return c.next
}

func (c *Command) nextFn() func(f Editor) {
	if c.next == nil {
		return nil
	}
	return c.next.fn
}

func compileAddr(a Address) func(f Editor) {
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
	fn := func(f Editor) {
		addr := compileAddr(p.addr)
		if addr != nil {
			addr(f)
		}
		if p.cmd != nil && p.cmd[0] != nil && p.cmd[0].fn != nil {
			p.cmd[0].fn(f)
		}
	}
	return &Command{fn: fn, Emit: p.Emit}
}
