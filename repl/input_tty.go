package repl

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"karl/interpreter"
)

type ttyByteEvent struct {
	b   byte
	err error
}

type ttyInput struct {
	in         *os.File
	out        io.Writer
	state      *term.State
	events     chan ttyByteEvent
	history    []string
	maxHistory int
}

func newTTYInput(in io.Reader, out io.Writer) (*ttyInput, bool) {
	inFile, ok := in.(*os.File)
	if !ok {
		return nil, false
	}
	outFile, ok := out.(*os.File)
	if !ok {
		return nil, false
	}
	if !term.IsTerminal(int(inFile.Fd())) || !term.IsTerminal(int(outFile.Fd())) {
		return nil, false
	}

	state, err := term.MakeRaw(int(inFile.Fd()))
	if err != nil {
		return nil, false
	}

	ti := &ttyInput{
		in:         inFile,
		out:        out,
		state:      state,
		events:     make(chan ttyByteEvent, 128),
		history:    make([]string, 0, 256),
		maxHistory: 1000,
	}
	go ti.readBytes()
	return ti, true
}

func (t *ttyInput) Close() {
	if t == nil || t.state == nil {
		return
	}
	_ = term.Restore(int(t.in.Fd()), t.state)
}

func (t *ttyInput) readBytes() {
	defer close(t.events)
	buf := make([]byte, 1)
	for {
		n, err := t.in.Read(buf)
		if n > 0 {
			t.events <- ttyByteEvent{b: buf[0]}
		}
		if err != nil {
			t.events <- ttyByteEvent{err: err}
			return
		}
	}
}

func (t *ttyInput) readLine(prompt string, eval *interpreter.Evaluator) (string, bool) {
	if t == nil {
		return "", false
	}
	line := make([]byte, 0, 64)
	cursor := 0
	historyIndex := len(t.history)
	inHistoryNav := false
	draftLine := make([]byte, 0, 64)
	fmt.Fprint(t.out, prompt)

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-t.events:
			if !ok {
				return "", false
			}
			if ev.err != nil {
				return "", false
			}

			switch ev.b {
			case '\r', '\n':
				fmt.Fprint(t.out, "\r\n")
				entered := string(line)
				t.appendHistory(entered)
				return entered, true
			case 0x03: // Ctrl+C
				fmt.Fprint(t.out, "^C\r\n")
				return "", false
			case 0x04: // Ctrl+D
				if len(line) == 0 {
					fmt.Fprint(t.out, "\r\n")
					return "", false
				}
			case 0x0c: // Ctrl+L
				clearScreen(t.out)
				redrawLine(t.out, prompt, line, cursor)
			case 0x7f, 0x08: // Backspace
				if cursor > 0 {
					if inHistoryNav {
						inHistoryNav = false
						historyIndex = len(t.history)
					}
					line = append(line[:cursor-1], line[cursor:]...)
					cursor--
					redrawLine(t.out, prompt, line, cursor)
				}
			case 0x1b: // Escape sequence (arrows/home/end/delete)
				next, ok := t.readByteWithTimeout(10 * time.Millisecond)
				if !ok {
					continue
				}
				if next != '[' && next != 'O' {
					continue
				}
				code, ok := t.readByteWithTimeout(10 * time.Millisecond)
				if !ok {
					continue
				}
				switch code {
				case 'A': // Up arrow
					if len(t.history) == 0 {
						continue
					}
					if !inHistoryNav {
						draftLine = append(draftLine[:0], line...)
						inHistoryNav = true
						historyIndex = len(t.history) - 1
					} else if historyIndex > 0 {
						historyIndex--
					}
					line = []byte(t.history[historyIndex])
					cursor = len(line)
					redrawLine(t.out, prompt, line, cursor)
				case 'B': // Down arrow
					if !inHistoryNav {
						continue
					}
					if historyIndex < len(t.history)-1 {
						historyIndex++
						line = []byte(t.history[historyIndex])
					} else {
						inHistoryNav = false
						historyIndex = len(t.history)
						line = append([]byte(nil), draftLine...)
					}
					cursor = len(line)
					redrawLine(t.out, prompt, line, cursor)
				case 'D': // Left arrow
					if cursor > 0 {
						cursor--
						redrawLine(t.out, prompt, line, cursor)
					}
				case 'C': // Right arrow
					if cursor < len(line) {
						cursor++
						redrawLine(t.out, prompt, line, cursor)
					}
				case 'H': // Home
					cursor = 0
					redrawLine(t.out, prompt, line, cursor)
				case 'F': // End
					cursor = len(line)
					redrawLine(t.out, prompt, line, cursor)
				case '3': // Delete sequence ESC [ 3 ~
					termByte, ok := t.readByteWithTimeout(10 * time.Millisecond)
					if ok && termByte == '~' && cursor < len(line) {
						if inHistoryNav {
							inHistoryNav = false
							historyIndex = len(t.history)
						}
						line = append(line[:cursor], line[cursor+1:]...)
						redrawLine(t.out, prompt, line, cursor)
					}
				}
			default:
				// Printable ASCII and tab; insert at cursor position.
				if ev.b >= 0x20 || ev.b == '\t' {
					if inHistoryNav {
						inHistoryNav = false
						historyIndex = len(t.history)
					}
					line = append(line, 0)
					copy(line[cursor+1:], line[cursor:])
					line[cursor] = ev.b
					cursor++
					redrawLine(t.out, prompt, line, cursor)
				}
			}
		case <-ticker.C:
			err := eval.CheckUnhandledTaskFailures()
			if err == nil {
				continue
			}
			// Move to a clean line for the error, then redraw the input line.
			fmt.Fprint(t.out, "\r\x1b[K")
			fmt.Fprintf(t.out, "Error: %s\r\n", err)
			if isFatalREPLError(err) {
				return "", false
			}
			redrawLine(t.out, prompt, line, cursor)
		}
	}
}

func (t *ttyInput) appendHistory(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	if n := len(t.history); n > 0 && t.history[n-1] == line {
		return
	}
	t.history = append(t.history, line)
	if t.maxHistory > 0 && len(t.history) > t.maxHistory {
		t.history = t.history[len(t.history)-t.maxHistory:]
	}
}

func (t *ttyInput) readByteWithTimeout(timeout time.Duration) (byte, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case ev, ok := <-t.events:
		if !ok || ev.err != nil {
			return 0, false
		}
		return ev.b, true
	case <-timer.C:
		return 0, false
	}
}

func redrawLine(out io.Writer, prompt string, line []byte, cursor int) {
	fmt.Fprintf(out, "\r%s%s\x1b[K", prompt, string(line))
	moveLeft := len(line) - cursor
	if moveLeft > 0 {
		fmt.Fprintf(out, "\x1b[%dD", moveLeft)
	}
}

func clearScreen(out io.Writer) {
	fmt.Fprint(out, "\x1b[H\x1b[2J")
}
