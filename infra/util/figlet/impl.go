package figlet

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type FlfFont struct {
	Height    int
	Baseline  int
	MaxWidth  int
	Hardblank byte
	SmushMode int
	Chars     map[rune][]string
}

const (
	SM_SMUSH     = 1
	SM_KERN      = 2
	SM_HARDBLANK = 16
	SM_EQUAL     = 32
	SM_LOWLINE   = 64
	SM_HIERARCHY = 128
)

var hierarchy = map[byte]int{
	'_': 1,
	'|': 2, '/': 2, '\\': 2,
	'[': 3, ']': 3, '{': 3, '}': 3,
	'(': 3, ')': 3,
	'<': 3, '>': 3,
	'+':  4,
	'-':  4,
	'*':  5,
	'.':  6,
	'=':  7,
	'^':  8,
	'"':  9,
	'\'': 9,
	'~':  10,
	'!':  11,
	'?':  11,
	'@':  12,
	'#':  13,
	'%':  14,
	'$':  15,
	'&':  16,
	':':  17,
	';':  17,
}

//go:embed ansi_shadow.flf
var defaultFontData []byte
var defaultFont *FlfFont
var defaultFontInitOnce sync.Once

func Render(text string) (string, error) {
	defaultFontInitOnce.Do(func() {
		font, err := ParseFlfFromBytes(defaultFontData)
		if err != nil {
			panic(err)
		}
		defaultFont = font
	})
	return defaultFont.Render(text)
}

func ParseFlfFromBytes(data []byte) (*FlfFont, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !scanner.Scan() {
		return nil, errors.New("empty font file")
	}

	headerLine := scanner.Text()
	fields := strings.Fields(headerLine)
	if len(fields) < 6 {
		return nil, fmt.Errorf("unsupported or invalid header: %s", headerLine)
	}

	sig := fields[0]
	if !strings.HasPrefix(sig, "flf2a") {
		return nil, fmt.Errorf("unsupported or invalid header: %s", headerLine)
	}
	if len(sig) < 6 {
		return nil, fmt.Errorf("invalid hardblank in header: %s", headerLine)
	}
	hardblank := sig[len("flf2a")]

	height, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid height: %w", err)
	}
	baseline, err := strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("invalid baseline: %w", err)
	}
	maxWidth, err := strconv.Atoi(fields[3])
	if err != nil {
		return nil, fmt.Errorf("invalid max width: %w", err)
	}
	oldLayout, err := strconv.Atoi(fields[4])
	if err != nil {
		return nil, fmt.Errorf("invalid old layout: %w", err)
	}
	commentLines, err := strconv.Atoi(fields[5])
	if err != nil {
		return nil, fmt.Errorf("invalid comment lines: %w", err)
	}

	smushMode := oldLayout
	if len(fields) >= 8 {
		fullLayout, err := strconv.Atoi(fields[7])
		if err == nil {
			smushMode = fullLayout
		}
	}

	for i := 0; i < commentLines && scanner.Scan(); i++ {
	}

	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	startASCII := 32
	endASCII := 126
	numChars := endASCII - startASCII + 1
	expectedTotalLines := numChars * height
	for len(allLines) < expectedTotalLines {
		allLines = append(allLines, "")
	}

	chars := make(map[rune][]string)
	for i := 0; i < numChars; i++ {
		r := rune(startASCII + i)
		var glyph []string
		for j := 0; j < height; j++ {
			idx := i*height + j
			line := allLines[idx]
			line = strings.TrimRight(line, "@")
			glyph = append(glyph, line)
		}
		chars[r] = glyph
	}

	return &FlfFont{
		Height:    height,
		Baseline:  baseline,
		MaxWidth:  maxWidth,
		Hardblank: hardblank,
		SmushMode: smushMode,
		Chars:     chars,
	}, nil
}

func (f *FlfFont) replaceHardblank(s string) string {
	if f.Hardblank == ' ' {
		return s
	}
	return strings.ReplaceAll(s, string(f.Hardblank), " ")
}

func (f *FlfFont) smushTwoLines(left, right string) string {
	if f.SmushMode == 0 {
		return left + right
	}

	canSmush := (f.SmushMode&SM_SMUSH) != 0 || (f.SmushMode&SM_KERN) != 0

	widthL := len(left)
	widthR := len(right)
	if widthL == 0 {
		return right
	}
	if widthR == 0 {
		return left
	}

	overlap := 0
	if canSmush {
		leftChar := left[widthL-1]
		rightChar := right[0]

		leftIsHB := leftChar == f.Hardblank
		rightIsHB := rightChar == f.Hardblank

		if leftIsHB || rightIsHB {
			if (f.SmushMode & SM_HARDBLANK) == 0 {
				return left + right
			}
			if leftIsHB {
				leftChar = ' '
			}
			if rightIsHB {
				rightChar = ' '
			}
		}

		if leftChar == ' ' && rightChar == ' ' {
			if (f.SmushMode & SM_KERN) != 0 {
				overlap = 1
			}
		} else if (f.SmushMode & SM_SMUSH) != 0 {
			smushed := byte(' ')
			applied := false

			if (f.SmushMode&SM_EQUAL) != 0 && leftChar == rightChar && leftChar != ' ' {
				smushed = leftChar
				applied = true
			}

			if !applied && (f.SmushMode&SM_LOWLINE) != 0 {
				if leftChar == '_' && rightChar == ' ' {
					smushed = '_'
					applied = true
				} else if leftChar == ' ' && rightChar == '_' {
					smushed = '_'
					applied = true
				}
			}

			if !applied && (f.SmushMode&SM_HIERARCHY) != 0 {
				leftPrec := hierarchy[leftChar]
				rightPrec := hierarchy[rightChar]
				if leftPrec > 0 || rightPrec > 0 {
					if leftPrec > rightPrec {
						smushed = leftChar
						applied = true
					} else if rightPrec > leftPrec {
						smushed = rightChar
						applied = true
					} else if leftPrec == rightPrec && leftPrec > 0 {
						smushed = rightChar
						applied = true
					}
				}
			}

			if applied {
				overlap = 1
				newCol := smushed
				left = left[:widthL-1] + string(newCol)
				right = right[1:]
			}
		}
	}

	if overlap == 0 {
		return left + right
	}
	return left + right
}

func (f *FlfFont) Render(text string) (string, error) {
	if f.Height == 0 {
		return "", errors.New("font not initialized")
	}

	lines := make([]string, f.Height)

	for _, ch := range text {
		glyph, ok := f.Chars[ch]
		if !ok {
			glyph = f.Chars[' ']
			if glyph == nil {
				glyph = make([]string, f.Height)
				for i := range glyph {
					glyph[i] = strings.Repeat(" ", f.MaxWidth)
				}
			}
		}

		cleanGlyph := make([]string, f.Height)
		for i, line := range glyph {
			cleanGlyph[i] = f.replaceHardblank(line)
		}

		if lines[0] == "" {
			copy(lines, cleanGlyph)
			continue
		}

		newLines := make([]string, f.Height)
		for i := 0; i < f.Height; i++ {
			newLines[i] = f.smushTwoLines(lines[i], cleanGlyph[i])
		}
		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}
