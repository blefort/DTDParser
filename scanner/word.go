package scanner

import (
	"strings"

	"go.uber.org/zap"
)

type word struct {
	sequence   string
	append     bool
	isQuoted   bool
	endChar    string
	done       bool
	inSequence bool
	log        *zap.SugaredLogger
}

func newWord(Log *zap.SugaredLogger) *word {
	var w word
	w.append = false
	w.isQuoted = false
	w.done = false
	w.log = Log
	w.endChar = " "
	w.inSequence = false
	return &w
}

func (w *word) scan(s string) {

	if !w.isQuoted && s == "\n" || len(strings.TrimSpace(w.sequence)) > 0 && s == w.endChar || s == ">" || s == "\t" {
		w.done = true
	}

	if w.done {
		return
	}

	if s != "\"" {
		w.sequence += s
	}

	if s == "\"" {
		w.isQuoted = true
		w.endChar = "\""
	}

}

func (w *word) stopped() bool {
	return w.done
}

func (w *word) Read() string {
	return strings.TrimSpace(w.sequence)
}
