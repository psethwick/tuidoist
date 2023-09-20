%{
package filter 

import (
    "time"
    "strings"
)
%}
%union{
    token Token
    expr Expression
}

%type<expr> filter
%type<expr> expr
%type<expr> s_datetime
%type<expr> s_date
%type<expr> s_date_year
%type<expr> s_overdue s_nodate s_project_key s_project_all_key s_label_key s_no_labels
%type<expr> s_time
%token<token> STRING NUMBER
%token<token> MONTH_IDENT TWELVE_CLOCK_IDENT
%token<token> TODAY_IDENT TOMORROW_IDENT YESTERDAY_IDENT DAYS
%token<token> DUE BEFORE AFTER OVER OVERDUE NO DATE LABELS '#' '@'
%left '&' '|'

%%

filter
    :
    {
        $$ = VoidExpr{}
    }
    | expr
    {
        $$ = $1
        yylex.(*Lexer).result = $$
    }

expr
    : expr '|' expr
    {
        $$ = BoolInfixOpExpr{left: $1, operator: '|', right: $3}
    }
    | expr '&' expr
    {
        $$ = BoolInfixOpExpr{left: $1, operator: '&', right: $3}
    }
    | STRING
    {
        $$ = StringExpr{literal: $1.literal}
    }
    | s_project_key STRING
    {
        $$ = ProjectExpr{isAll: false, name: $2.literal}
    }
    | s_project_all_key STRING
    {
        $$ = ProjectExpr{isAll: true, name: $2.literal}
    }
    | s_label_key STRING
    {
        $$ = LabelExpr{name: $2.literal}
    }
    | s_no_labels
    {
        $$ = LabelExpr{name: ""}
    }
    | '(' expr ')'
    {
        $$ = $2
    }
    | '!' expr
    {
        $$ = NotOpExpr{expr: $2}
    }
    | s_overdue
    {
        $$ = DateExpr{allDay: false, datetime: now(), operation: DUE_BEFORE}
    }
    | s_nodate
    {
        $$ = DateExpr{operation: NO_DUE_DATE}
    }
    | NUMBER DAYS
    {
        date := today().AddDate(0, 0, atoi($1.literal))
        $$ = DateExpr{allDay: true, datetime: date, operation: DUE_BEFORE}
    }
    | DUE BEFORE ':' s_datetime
    {
        e := $4.(DateExpr)
        e.operation = DUE_BEFORE
        $$ = e
    }
    | DUE AFTER ':' s_datetime
    {
        e := $4.(DateExpr)
        e.operation = DUE_AFTER
        $$ = e
    }
    | s_datetime

s_project_all_key
    : '#' '#'
    {
        $$ = $1
    }

s_project_key
    : '#'
    {
        $$ = $1
    }

s_label_key
    : '@'
    {
        $$ = $1
    }

s_no_labels
    : NO LABELS
    {
        $$ = $1
    }

s_nodate
    : NO DATE
    {
        $$ = $1
    }
    | NO DUE DATE
    {
        $$ = $1
    }

s_overdue
    : OVER DUE
    {
        $$ = $1
    }
    | OVERDUE
    {
        $$ = $1
    }


s_datetime
    : s_date_year s_time
    {
        date := $1.(time.Time)
        time := $2.(time.Duration)
        $$ = DateExpr{allDay: false, datetime: date.Add(time)}
    }
    | s_date_year
    {
        $$ = DateExpr{allDay: true, datetime: $1.(time.Time)}
    }
    | s_time
    {
        nd := now().Sub(today())
        d := $1.(time.Duration)
        if (d <= nd) {
          d = d + time.Duration(int64(time.Hour) * 24)
        }
        $$ = DateExpr{allDay: false, datetime: today().Add(d)}
    }

s_date_year
    : NUMBER '/' NUMBER '/' NUMBER
    {
        $$ = time.Date(atoi($5.literal), time.Month(atoi($1.literal)), atoi($3.literal), 0, 0, 0, 0, timezone())
    }
    | MONTH_IDENT NUMBER NUMBER
    {
        $$ = time.Date(atoi($3.literal), MonthIdentHash[strings.ToLower($1.literal)], atoi($2.literal), 0, 0, 0, 0, timezone())
    }
    | NUMBER MONTH_IDENT NUMBER
    {
        $$ = time.Date(atoi($3.literal), MonthIdentHash[strings.ToLower($2.literal)], atoi($1.literal), 0, 0, 0, 0, timezone())
    }
    | s_date
    {
        tod := today()
        date := $1.(time.Time)
        if date.Before(tod) {
            date = date.AddDate(1, 0, 0)
        }
        $$ = date
    }
    | TODAY_IDENT
    {
        $$ = today()
    }
    | TOMORROW_IDENT
    {
        $$ = today().AddDate(0, 0, 1)
    }
    | YESTERDAY_IDENT
    {
        $$ = today().AddDate(0, 0, -1)
    }

s_date
    : MONTH_IDENT NUMBER
    {
        $$ = time.Date(today().Year(), MonthIdentHash[strings.ToLower($1.literal)], atoi($2.literal), 0, 0, 0, 0, timezone())
    }
    | NUMBER MONTH_IDENT
    {
        $$ = time.Date(today().Year(), MonthIdentHash[strings.ToLower($2.literal)], atoi($1.literal), 0, 0, 0, 0, timezone())
    }
    | NUMBER '/' NUMBER
    {
        $$ = time.Date(now().Year(), time.Month(atoi($3.literal)), atoi($1.literal), 0, 0, 0, 0, timezone())
    }

s_time
    : NUMBER ':' NUMBER
    {
        $$ = time.Duration(int64(time.Hour) * int64(atoi($1.literal)) + int64(time.Minute) * int64(atoi($3.literal)))
    }
    | NUMBER ':' NUMBER ':' NUMBER
    {
        $$ = time.Duration(int64(time.Hour) * int64(atoi($1.literal)) + int64(time.Minute) * int64(atoi($3.literal)) + int64(time.Second) * int64(atoi($5.literal)))
    }
    | NUMBER TWELVE_CLOCK_IDENT
    {
        hour := atoi($1.literal)
        if TwelveClockIdentHash[$2.literal] {
            hour = hour + 12
        }
        $$ = time.Duration(int64(time.Hour) * int64(hour))
    }

%%
