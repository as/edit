package edit

import (
	"errors"

	"github.com/as/text"
)

var (
	ErrNilFunc = errors.New("empty program")
)

type Command struct {
	fn   func(text.Editor)
	s    string
	args string
	next *Command
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

// Run runs the compiled program on ed
func (c *Command) Run(ed text.Editor) (err error) {
	if c.fn == nil {
		return ErrNilFunc
	}
	// TODO: return an error
	// TODO: type switch on Invertable and commit the transaction
	//       if that interface is implemented so Undo doesn't step
	//       through x,foo,x,., ...
	c.fn(ed)
	return err
}

// Next returns the next instruction for the compiled program. This
// effectively steps through x,..., and y,...,
func (c *Command) Next() *Command {
	return c.next
}

func (c *Command) nextFn() func(f text.Editor) {
	return c.next.fn
}

func compileAddr(a Address) func(f text.Editor) {
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
