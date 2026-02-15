package repl

import "io"

type ttyLineWriter struct {
	out io.Writer
}

func newTTYLineWriter(out io.Writer) io.Writer {
	return &ttyLineWriter{out: out}
}

func (w *ttyLineWriter) Write(p []byte) (int, error) {
	buf := make([]byte, 0, len(p)+8)
	for i := 0; i < len(p); i++ {
		b := p[i]
		if b == '\n' {
			prevCR := i > 0 && p[i-1] == '\r'
			if !prevCR {
				buf = append(buf, '\r')
			}
		}
		buf = append(buf, b)
	}
	_, err := w.out.Write(buf)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
