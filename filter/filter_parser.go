// Code generated by goyacc -o filter/filter_parser.go filter/filter_parser.y. DO NOT EDIT.

// adapted from upstream https://github.com/sachaos/todoist
//
//line filter/filter_parser.y:2
package filter

import __yyfmt__ "fmt"

//line filter/filter_parser.y:3

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
	"time"
)

type Expression interface{}
type Token struct {
	token   int
	literal string
}

type VoidExpr struct{}

type StringExpr struct {
	literal string
}

type BoolInfixOpExpr struct {
	left     Expression
	operator rune
	right    Expression
}

type ProjectExpr struct {
	isAll bool
	name  string
}

type LabelExpr struct {
	name string
}

type NotOpExpr struct {
	expr Expression
}

const (
	DUE_ON int = iota
	DUE_BEFORE
	DUE_AFTER
	NO_DUE_DATE
)

type DateExpr struct {
	operation int
	datetime  time.Time
	allDay    bool
}

func atoi(a string) (i int) {
	i, _ = strconv.Atoi(a)
	return
}

var now = time.Now
var today = func() time.Time {
	return time.Date(now().Year(), now().Month(), now().Day(), 0, 0, 0, 0, now().Location())
}
var timezone = func() *time.Location {
	return now().Location()
}

//line filter/filter_parser.y:73
type yySymType struct {
	yys   int
	token Token
	expr  Expression
}

const STRING = 57346
const NUMBER = 57347
const MONTH_IDENT = 57348
const TWELVE_CLOCK_IDENT = 57349
const TODAY_IDENT = 57350
const TOMORROW_IDENT = 57351
const YESTERDAY_IDENT = 57352
const DUE = 57353
const BEFORE = 57354
const AFTER = 57355
const OVER = 57356
const OVERDUE = 57357
const NO = 57358
const DATE = 57359
const LABELS = 57360

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"STRING",
	"NUMBER",
	"MONTH_IDENT",
	"TWELVE_CLOCK_IDENT",
	"TODAY_IDENT",
	"TOMORROW_IDENT",
	"YESTERDAY_IDENT",
	"DUE",
	"BEFORE",
	"AFTER",
	"OVER",
	"OVERDUE",
	"NO",
	"DATE",
	"LABELS",
	"'#'",
	"'@'",
	"'&'",
	"'|'",
	"'('",
	"')'",
	"'!'",
	"':'",
	"'/'",
}

var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line filter/filter_parser.y:295

type Lexer struct {
	scanner.Scanner
	result Expression
	error  string
}

var MonthIdentHash = map[string]time.Month{
	"jan":  time.January,
	"feb":  time.February,
	"mar":  time.March,
	"apr":  time.April,
	"may":  time.May,
	"june": time.June,
	"july": time.July,
	"aug":  time.August,
	"sept": time.September,
	"oct":  time.October,
	"nov":  time.November,
	"dec":  time.December,

	"january":   time.January,
	"february":  time.February,
	"march":     time.March,
	"april":     time.April,
	"august":    time.August,
	"september": time.September,
	"october":   time.October,
	"november":  time.November,
	"december":  time.December,
}

var TwelveClockIdentHash = map[string]bool{
	"am": false,
	"pm": true,
}

var TodayIdentHash = map[string]bool{
	"today": true,
	"tod":   true,
}

var TomorrowIdentHash = map[string]bool{
	"tomorrow": true,
	"tom":      true,
}

var OverDueHash = map[string]bool{
	"overdue": true,
	"od":      true,
}

func (l *Lexer) Lex(lval *yySymType) int {
	token := int(l.Scan())
	switch token {
	case scanner.Ident:
		lowerToken := strings.ToLower(l.TokenText())
		if _, ok := MonthIdentHash[lowerToken]; ok {
			token = MONTH_IDENT
		} else if _, ok := TwelveClockIdentHash[lowerToken]; ok {
			token = TWELVE_CLOCK_IDENT
		} else if _, ok := TodayIdentHash[lowerToken]; ok {
			token = TODAY_IDENT
		} else if _, ok := TomorrowIdentHash[lowerToken]; ok {
			token = TOMORROW_IDENT
		} else if lowerToken == "yesterday" {
			token = YESTERDAY_IDENT
		} else if lowerToken == "due" {
			token = DUE
		} else if lowerToken == "before" {
			token = BEFORE
		} else if lowerToken == "after" {
			token = AFTER
		} else if lowerToken == "over" {
			token = OVER
		} else if _, ok := OverDueHash[lowerToken]; ok {
			token = OVERDUE
		} else if lowerToken == "no" {
			token = NO
		} else if lowerToken == "date" {
			token = DATE
		} else if lowerToken == "labels" {
			token = LABELS
		} else {
			token = STRING
		}
	case scanner.Int:
		token = NUMBER
	}
	lval.token = Token{token: token, literal: l.TokenText()}
	return token
}

func (l *Lexer) Error(e string) {
	l.error = fmt.Sprintf("Filter error: %s \nFor proper filter syntax see https://support.todoist.com/hc/en-us/articles/205248842-Filters\n", e)
}

func Filter(f string) (e Expression, err error) {
	l := new(Lexer)
	l.Init(strings.NewReader(f))
	// important to exclude scanner.ScanFloats because afternoon times in am/pm format trigger float parsing
	l.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.SkipComments
	yyParse(l)
	if l.error != "" {
		return nil, errors.New(l.error)
	}
	return l.result, nil
}

//line yacctab:1
var yyExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 72

var yyAct = [...]int8{
	13, 3, 21, 22, 60, 24, 25, 26, 12, 2,
	46, 17, 18, 16, 44, 46, 14, 15, 32, 33,
	8, 61, 9, 20, 52, 51, 36, 28, 27, 45,
	50, 28, 27, 39, 45, 43, 53, 48, 49, 38,
	37, 21, 22, 41, 24, 25, 26, 34, 35, 40,
	63, 62, 58, 59, 57, 56, 55, 54, 47, 42,
	31, 30, 29, 7, 6, 5, 4, 11, 10, 19,
	23, 1,
}

var yyPact = [...]int16{
	-3, -1000, 10, -1000, 58, 57, 56, -1000, -3, -3,
	-1000, -1000, 35, -1000, 7, -1000, 22, 38, -1000, 54,
	-1000, 8, 53, -1000, -1000, -1000, -1000, -3, -3, -1000,
	-1000, -1000, 6, 10, -1, -2, -1000, -1000, -1000, 19,
	-1000, -1000, 3, 52, 51, 50, -1000, 49, -1000, -1000,
	-1000, 36, 36, -1000, -23, -1000, -5, -1000, -1000, -1000,
	46, 45, -1000, -1000,
}

var yyPgo = [...]int8{
	0, 71, 9, 0, 70, 69, 68, 67, 66, 65,
	64, 63, 23,
}

var yyR1 = [...]int8{
	0, 1, 1, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 9, 8, 10,
	11, 7, 7, 6, 6, 3, 3, 3, 5, 5,
	5, 5, 5, 5, 5, 4, 4, 4, 12, 12,
	12,
}

var yyR2 = [...]int8{
	0, 0, 1, 3, 3, 1, 2, 2, 2, 1,
	3, 2, 1, 1, 4, 4, 1, 2, 1, 1,
	2, 2, 3, 2, 1, 2, 1, 1, 5, 3,
	3, 1, 1, 1, 1, 2, 2, 3, 3, 5,
	2,
}

var yyChk = [...]int16{
	-1000, -1, -2, 4, -8, -9, -10, -11, 23, 25,
	-6, -7, 11, -3, 19, 20, 16, 14, 15, -5,
	-12, 5, 6, -4, 8, 9, 10, 22, 21, 4,
	4, 4, -2, -2, 12, 13, 19, 18, 17, 11,
	11, -12, 5, 27, 6, 26, 7, 5, -2, -2,
	24, 26, 26, 17, 5, 5, 5, 5, -3, -3,
	27, 26, 5, 5,
}

var yyDef = [...]int8{
	1, -2, 2, 5, 0, 0, 0, 9, 0, 0,
	12, 13, 0, 16, 18, 19, 0, 0, 24, 26,
	27, 0, 0, 31, 32, 33, 34, 0, 0, 6,
	7, 8, 0, 11, 0, 0, 17, 20, 21, 0,
	23, 25, 0, 0, 36, 0, 40, 35, 3, 4,
	10, 0, 0, 22, 37, 30, 38, 29, 14, 15,
	0, 0, 28, 39,
}

var yyTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 25, 3, 19, 3, 3, 21, 3,
	23, 24, 3, 3, 3, 3, 3, 27, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 26, 3,
	3, 3, 3, 3, 20, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 22,
}

var yyTok2 = [...]int8{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18,
}

var yyTok3 = [...]int8{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(yyPact[state])
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && int(yyChk[int(yyAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || int(yyExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := int(yyExca[i])
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(yyTok1[0])
		goto out
	}
	if char < len(yyTok1) {
		token = int(yyTok1[char])
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = int(yyTok2[char-yyPrivate])
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = int(yyTok3[i+0])
		if token == char {
			token = int(yyTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(yyTok2[1]) /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = int(yyPact[yystate])
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = int(yyAct[yyn])
	if int(yyChk[yyn]) == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = int(yyDef[yystate])
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && int(yyExca[xi+1]) == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = int(yyExca[xi+0])
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = int(yyExca[xi+1])
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = int(yyPact[yyS[yyp].yys]) + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = int(yyAct[yyn]) /* simulate a shift of "error" */
					if int(yyChk[yystate]) == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= int(yyR2[yyn])
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = int(yyR1[yyn])
	yyg := int(yyPgo[yyn])
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = int(yyAct[yyg])
	} else {
		yystate = int(yyAct[yyj])
		if int(yyChk[yystate]) != -yyn {
			yystate = int(yyAct[yyg])
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-0 : yypt+1]
//line filter/filter_parser.y:95
		{
			yyVAL.expr = VoidExpr{}
		}
	case 2:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:99
		{
			yyVAL.expr = yyDollar[1].expr
			yylex.(*Lexer).result = yyVAL.expr
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:106
		{
			yyVAL.expr = BoolInfixOpExpr{left: yyDollar[1].expr, operator: '|', right: yyDollar[3].expr}
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:110
		{
			yyVAL.expr = BoolInfixOpExpr{left: yyDollar[1].expr, operator: '&', right: yyDollar[3].expr}
		}
	case 5:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:114
		{
			yyVAL.expr = StringExpr{literal: yyDollar[1].token.literal}
		}
	case 6:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:118
		{
			yyVAL.expr = ProjectExpr{isAll: false, name: yyDollar[2].token.literal}
		}
	case 7:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:122
		{
			yyVAL.expr = ProjectExpr{isAll: true, name: yyDollar[2].token.literal}
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:126
		{
			yyVAL.expr = LabelExpr{name: yyDollar[2].token.literal}
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:130
		{
			yyVAL.expr = LabelExpr{name: ""}
		}
	case 10:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:134
		{
			yyVAL.expr = yyDollar[2].expr
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:138
		{
			yyVAL.expr = NotOpExpr{expr: yyDollar[2].expr}
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:142
		{
			yyVAL.expr = DateExpr{allDay: false, datetime: now(), operation: DUE_BEFORE}
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:146
		{
			yyVAL.expr = DateExpr{operation: NO_DUE_DATE}
		}
	case 14:
		yyDollar = yyS[yypt-4 : yypt+1]
//line filter/filter_parser.y:150
		{
			e := yyDollar[4].expr.(DateExpr)
			e.operation = DUE_BEFORE
			yyVAL.expr = e
		}
	case 15:
		yyDollar = yyS[yypt-4 : yypt+1]
//line filter/filter_parser.y:156
		{
			e := yyDollar[4].expr.(DateExpr)
			e.operation = DUE_AFTER
			yyVAL.expr = e
		}
	case 17:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:165
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 18:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:171
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 19:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:177
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 20:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:183
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 21:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:189
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:193
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:199
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 24:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:203
		{
			yyVAL.expr = yyDollar[1].token
		}
	case 25:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:209
		{
			date := yyDollar[1].expr.(time.Time)
			time := yyDollar[2].expr.(time.Duration)
			yyVAL.expr = DateExpr{allDay: false, datetime: date.Add(time)}
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:215
		{
			yyVAL.expr = DateExpr{allDay: true, datetime: yyDollar[1].expr.(time.Time)}
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:219
		{
			nd := now().Sub(today())
			d := yyDollar[1].expr.(time.Duration)
			if d <= nd {
				d = d + time.Duration(int64(time.Hour)*24)
			}
			yyVAL.expr = DateExpr{allDay: false, datetime: today().Add(d)}
		}
	case 28:
		yyDollar = yyS[yypt-5 : yypt+1]
//line filter/filter_parser.y:230
		{
			yyVAL.expr = time.Date(atoi(yyDollar[5].token.literal), time.Month(atoi(yyDollar[1].token.literal)), atoi(yyDollar[3].token.literal), 0, 0, 0, 0, timezone())
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:234
		{
			yyVAL.expr = time.Date(atoi(yyDollar[3].token.literal), MonthIdentHash[strings.ToLower(yyDollar[1].token.literal)], atoi(yyDollar[2].token.literal), 0, 0, 0, 0, timezone())
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:238
		{
			yyVAL.expr = time.Date(atoi(yyDollar[3].token.literal), MonthIdentHash[strings.ToLower(yyDollar[2].token.literal)], atoi(yyDollar[1].token.literal), 0, 0, 0, 0, timezone())
		}
	case 31:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:242
		{
			tod := today()
			date := yyDollar[1].expr.(time.Time)
			if date.Before(tod) {
				date = date.AddDate(1, 0, 0)
			}
			yyVAL.expr = date
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:251
		{
			yyVAL.expr = today()
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:255
		{
			yyVAL.expr = today().AddDate(0, 0, 1)
		}
	case 34:
		yyDollar = yyS[yypt-1 : yypt+1]
//line filter/filter_parser.y:259
		{
			yyVAL.expr = today().AddDate(0, 0, -1)
		}
	case 35:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:265
		{
			yyVAL.expr = time.Date(today().Year(), MonthIdentHash[strings.ToLower(yyDollar[1].token.literal)], atoi(yyDollar[2].token.literal), 0, 0, 0, 0, timezone())
		}
	case 36:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:269
		{
			yyVAL.expr = time.Date(today().Year(), MonthIdentHash[strings.ToLower(yyDollar[2].token.literal)], atoi(yyDollar[1].token.literal), 0, 0, 0, 0, timezone())
		}
	case 37:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:273
		{
			yyVAL.expr = time.Date(now().Year(), time.Month(atoi(yyDollar[3].token.literal)), atoi(yyDollar[1].token.literal), 0, 0, 0, 0, timezone())
		}
	case 38:
		yyDollar = yyS[yypt-3 : yypt+1]
//line filter/filter_parser.y:279
		{
			yyVAL.expr = time.Duration(int64(time.Hour)*int64(atoi(yyDollar[1].token.literal)) + int64(time.Minute)*int64(atoi(yyDollar[3].token.literal)))
		}
	case 39:
		yyDollar = yyS[yypt-5 : yypt+1]
//line filter/filter_parser.y:283
		{
			yyVAL.expr = time.Duration(int64(time.Hour)*int64(atoi(yyDollar[1].token.literal)) + int64(time.Minute)*int64(atoi(yyDollar[3].token.literal)) + int64(time.Second)*int64(atoi(yyDollar[5].token.literal)))
		}
	case 40:
		yyDollar = yyS[yypt-2 : yypt+1]
//line filter/filter_parser.y:287
		{
			hour := atoi(yyDollar[1].token.literal)
			if TwelveClockIdentHash[yyDollar[2].token.literal] {
				hour = hour + 12
			}
			yyVAL.expr = time.Duration(int64(time.Hour) * int64(hour))
		}
	}
	goto yystack /* stack new state and value */
}
