package scanner

import (
	"strings"

	"github.com/blefort/DTDParser/DTD"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type sentence struct {
	sequence         string
	openedStartChar  int
	DTDType          int
	append           bool
	appendTosentence bool
	start            string
	end              string
	value            string
	log              *zap.SugaredLogger
	words            []*word
}

func newsentence(start string, end string, Log *zap.SugaredLogger) *sentence {
	var se sentence
	se.start = start
	se.end = end
	se.append = false
	se.appendTosentence = false
	se.log = Log
	se.DTDType = DTD.UNIDENTIFIED
	se.openedStartChar = 0

	word := newWord(Log)
	se.words = append(se.words, word)
	return &se
}

func (se *sentence) MarshalLogObject(enc zapcore.ObjectEncoder) error {

	for i, w := range se.words {
		enc.AddString("Word "+string(i), w.read())
	}
	return nil
}

func (se *sentence) getWords() []*word {
	var words []*word
	for _, w := range se.words {
		if w.stopped() && w.read() != "" {
			words = append(words, w)
		}
	}
	return words
}

func (se *sentence) scan(s string) bool {

	wordIdx := len(se.words) - 1

	if !se.append && s == se.start[0:1] {
		se.append = true
	}

	if !se.append {
		return se.stopped()
	}

	if s == se.start {
		se.openedStartChar++
	}
	if s == se.end {
		se.openedStartChar--
	}

	// start appending
	if se.sequence == se.start {
		se.appendTosentence = true
		se.append = true
		se.sequence = se.start
	}

	se.log.Warnf("char:" + s)
	se.sequence += s

	if se.appendTosentence {

		se.value += s

		if s == se.end && se.openedStartChar == 0 {
			se.appendTosentence = false
			se.append = false
			return se.stopped()
		}

		if !se.words[wordIdx].stopped() {
			se.words[wordIdx].scan(s)
		} else {
			word := newWord(se.log)
			se.words = append(se.words, word)
			wordIdx = len(se.words) - 1
			se.words[wordIdx].scan(s)
		}

	}

	return se.stopped()

}

func (se *sentence) stopped() bool {
	return len(se.value) > 0 && !se.appendTosentence
}

func (se *sentence) read() string {
	//se.log.Warnf("sentence delimited by '%s' and '%s' is [%s], sequence [%s]", se.start, se.end, se.value, se.sequence)
	//if se.stopped() {
	se.log.Warnf("sequence is [" + se.sequence + "]")
	//}
	if len(se.value) > len(se.end) {
		return strings.TrimSpace(se.value[0 : len(se.value)-len(se.end)])
	}
	return strings.TrimSpace(se.value)
}

func (se *sentence) readSequence() string {
	return strings.Trim(se.sequence[1:len(se.sequence)-1], "!- ")
}

func (se *sentence) readAndClose() string {

	var s = se.value[0 : len(se.value)-len(se.end)]
	//se.log.Warnf("sentence delimited by '%s' and '%s' is [%s], sequence [%s]", se.start, se.end, s, se.sequence)
	se.value = ""
	se.sequence = ""
	return s
}
