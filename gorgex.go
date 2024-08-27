package gorgex

import (
	"fmt"
	"strconv"
	"strings"
)

type tokenType uint8

// Possible types of token
const (
	group           tokenType = iota
	bracket         tokenType = iota
	or              tokenType = iota
	repeat          tokenType = iota
	literal         tokenType = iota
	groupUncaptured tokenType = iota
)

type token struct {
	tokenType tokenType
	value     interface{}
}

type parseContext struct {
	pos    int     // position of cursor in regex
	tokens []token // tokens already parsed from regex string
}

type repeatPayload struct {
	min   int
	max   int
	token token
}

func parse(regex string) *parseContext {
	// Create the initial parse context
	ctx := &parseContext{
		pos: 0, tokens: []token{},
	}

	// Go through each token, and process them, increasing
	// the position as we go through the string
	for ctx.pos < len(regex) {
		process(regex, ctx)
		ctx.pos++
	}

	return ctx
}

func process(regex string, ctx *parseContext) {
	ch := regex[ctx.pos]
	// Call correct method depending on current token
	if ch == '(' {
		// Start of an expression
		// Create a context for this grouped expression
		groupCtx := &parseContext{
			pos:    ctx.pos,
			tokens: []token{},
		}
		parseGroup(regex, groupCtx)
		ctx.tokens = append(ctx.tokens, token{
			tokenType: group,
			value:     groupCtx.tokens,
		})
	} else if ch == '[' {
		parseBracket(regex, ctx)
	} else if ch == '|' {
		parseOr(regex, ctx)
	} else if ch == '*' || ch == '?' || ch == '+' {
		parseRepeat(regex, ctx)
	} else if ch == '{' {
		parseRepeatSpecified(regex, ctx)
	} else {
		// Must be literal
		t := token{
			tokenType: literal,
			value:     ch,
		}

		ctx.tokens = append(ctx.tokens, t)
	}
}

func parseGroup(regex string, ctx *parseContext) {
	ctx.pos += 1
	// Loop until hitting a right paren (this will
	// work with nested parens, as those will induce a
	// recursive call to this function)
	for regex[ctx.pos] != ')' {
		process(regex, ctx)
		ctx.pos += 1
	}
}

func parseBracket(regex string, ctx *parseContext) {
	ctx.pos++
	var literals []string
	// Loop until hitting correct closing bracket
	for regex[ctx.pos] != ']' {
		ch := regex[ctx.pos]

		if ch == '-' {
			// Need next and previous for ranges
			next := regex[ctx.pos+1]
			prev := regex[ctx.pos-1]
			literals[len(literals)-1] = fmt.Sprintf("%c%c", prev, next)
			ctx.pos++
		} else {
			literals = append(literals, fmt.Sprintf("%c", ch))
		}

		ctx.pos++
	}

	literalsSet := map[uint8]bool{}

	for _, l := range literals {
		// Iterate through all characters between the first and last elements
		// If those are the same, this is just the single element
		for i := l[0]; i <= l[len(l)-1]; i++ {
			literalsSet[i] = true
		}
	}

	ctx.tokens = append(ctx.tokens, token{
		tokenType: bracket,
		value:     literalsSet,
	})
}

func parseOr(regex string, ctx *parseContext) {
	rhsContext := &parseContext{
		pos:    ctx.pos,
		tokens: []token{},
	}
	rhsContext.pos += 1
	for rhsContext.pos < len(regex) && regex[rhsContext.pos] != ')' {
		process(regex, rhsContext)
		rhsContext.pos += 1
	}

	left := token{
		tokenType: groupUncaptured,
		value:     ctx.tokens,
	}

	right := token{
		tokenType: groupUncaptured,
		value:     rhsContext.tokens,
	}

	if regex[rhsContext.pos] == ')' {
		rhsContext.pos--
	}
	ctx.pos = rhsContext.pos

	ctx.tokens = []token{{
		tokenType: or,
		value:     []token{left, right},
	}}
}

const repeatInfinity = -1

func parseRepeat(regex string, ctx *parseContext) {
	ch := regex[ctx.pos]
	var min, max int
	if ch == '*' {
		min = 0
		max = repeatInfinity
	} else if ch == '?' {
		min = 0
		max = 1
	} else {
		min = 1
		max = repeatInfinity
	}

	lastToken := ctx.tokens[len(ctx.tokens)-1]
	ctx.tokens[len(ctx.tokens)-1] = token{
		tokenType: repeat,
		value: repeatPayload{
			min:   min,
			max:   max,
			token: lastToken,
		},
	}
}

func parseRepeatSpecified(regex string, ctx *parseContext) {
	start := ctx.pos + 1
	for regex[ctx.pos] != '}' {
		ctx.pos++
	}
	boundariesStr := regex[start:ctx.pos]
	pieces := strings.Split(boundariesStr, ",")
	var min, max int
	if len(pieces) == 1 {
		if value, err := strconv.Atoi(pieces[0]); err != nil {
			panic(err.Error())
		} else {
			min = value
			max = value
		}
	} else if len(pieces) == 2 {
		if value, err := strconv.Atoi(pieces[0]); err != nil {
			panic(err.Error())
		} else {
			min = value
		}

		if pieces[1] == "" {
			max = repeatInfinity
		} else if value, err := strconv.Atoi(pieces[1]); err != nil {
			panic(err.Error())
		} else {
			max = value
		}
	} else {
		panic(fmt.Sprintf("There must be either 1 or 2 values specified for the quantifier: provided '%s'", boundariesStr))
	}

	lastToken := ctx.tokens[len(ctx.tokens)-1]
	ctx.tokens[len(ctx.tokens)-1] = token{
		tokenType: repeat,
		value: repeatPayload{
			min:   min,
			max:   max,
			token: lastToken,
		},
	}
}

type state struct {
	start       bool
	terminal    bool
	transitions map[uint8][]*state
}

const epsilonChar uint8 = 0

func toNfa(ctx *parseContext) *state {
	startState, endState := tokenToNfa(&ctx.tokens[0])

	for i := 1; i < len(ctx.tokens); i++ {
		startNext, endNext := tokenToNfa(&ctx.tokens[i])
		endState.transitions[epsilonChar] = append(
			endState.transitions[epsilonChar],
			startNext,
		)
		endState = endNext
	}

	start := &state{
		transitions: map[uint8][]*state{
			epsilonChar: {startState},
		},
		start: true,
	}
	end := &state{
		transitions: map[uint8][]*state{},
		terminal:    true,
	}

	endState.transitions[epsilonChar] = append(
		endState.transitions[epsilonChar],
		end,
	)
	return start
}

func tokenToNfa(t *token) (*state, *state) {
	start := &state{
		transitions: map[uint8][]*state{},
	}
	end := &state{
		transitions: map[uint8][]*state{},
	}

	switch t.tokenType {
	case literal:
		ch := t.value.(uint8)
		start.transitions[ch] = []*state{end}
	case or:
		values := t.value.([]token)
		left := values[0]
		right := values[1]

		s1, e1 := tokenToNfa(&left)
		s2, e2 := tokenToNfa(&right)

		start.transitions[epsilonChar] = []*state{s1, s2}
		e1.transitions[epsilonChar] = []*state{end}
		e2.transitions[epsilonChar] = []*state{end}
	case bracket:
		literals := t.value.(map[uint8]bool)

		for l := range literals {
			start.transitions[l] = []*state{end}
		}
	case group, groupUncaptured:
		tokens := t.value.([]token)
		start, end = tokenToNfa(&tokens[0])
		for i := 1; i < len(tokens); i++ {
			ts, te := tokenToNfa(&tokens[i])
			end.transitions[epsilonChar] = append(
				end.transitions[epsilonChar],
				ts,
			)
			end = te
		}
	case repeat:
		p := t.value.(repeatPayload)

		if p.min == 0 {
			start.transitions[epsilonChar] = []*state{end}
		}

		var copyCount int

		if p.max == repeatInfinity {
			if p.min == 0 {
				copyCount = 1
			} else {
				copyCount = p.min
			}
		} else {
			copyCount = p.max
		}

		from, to := tokenToNfa(&p.token)
		start.transitions[epsilonChar] = append(
			start.transitions[epsilonChar],
			from,
		)

		for i := 2; i <= copyCount; i++ {
			s, e := tokenToNfa(&p.token)

			to.transitions[epsilonChar] = append(
				to.transitions[epsilonChar],
				s,
			)

			from = s
			to = e

			if i > p.min {
				s.transitions[epsilonChar] = append(
					s.transitions[epsilonChar],
					end,
				)
			}
		}

		to.transitions[epsilonChar] = append(
			to.transitions[epsilonChar],
			end,
		)

		if p.max == repeatInfinity {
			end.transitions[epsilonChar] = append(
				end.transitions[epsilonChar],
				from,
			)
		}
	default:
		panic("Unknown type of token")
	}

	return start, end
}

const (
	startOfText uint8 = 1
	endOfText   uint8 = 2
)

func getChar(input string, pos int) uint8 {
	if pos >= len(input) {
		return endOfText
	}

	if pos < 0 {
		return startOfText
	}
	return input[pos]
}

func (s *state) check(input string, pos int) bool {
	ch := getChar(input, pos)

	if ch == endOfText && s.terminal {
		return true
	}

	if states := s.transitions[ch]; len(states) > 0 {
		nextState := states[0]
		if nextState.check(input, pos+1) {
			return true
		}
	}

	for _, state := range s.transitions[epsilonChar] {
		if state.check(input, pos) {
			return true
		}

		if ch == startOfText && state.check(input, pos+1) {
			return true
		}
	}

	return false
}
