package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Expression struct {
	Id         int
	Expression string
	Result     string
	Status     int
}

type SubExpression struct {
	Id          int
	Id_sub      string
	First       string
	Second      string
	Action      string
	Result      string
	In_progress int
	Main        int
	Id_main     int
}

type Time struct {
	Id        int
	Operation string
	Seconds   int
}

type Server struct {
	Id            int
	Server_status string
	Goroutines    int
	Next_ping     string
}

func Create() {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Query("CREATE TABLE IF NOT EXISTS public.expressions(expression_id integer NOT NULL GENERATED BY DEFAULT AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),expression_string character varying(50) NOT NULL,status integer NOT NULL,result character varying(100) NOT NULL DEFAULT '?'::character varying,PRIMARY KEY (expression_id));;")
	if err != nil {
		panic(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS public.servers(server_id integer NOT NULL GENERATED BY DEFAULT AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),server_status character varying(50) NOT NULL DEFAULT 'idle'::character varying,goroutines integer NOT NULL DEFAULT 0,next_ping character varying NOT NULL,PRIMARY KEY (server_id));")
	if err != nil {
		panic(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS public.sub_exps(id integer NOT NULL GENERATED BY DEFAULT AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),id_sub character varying NOT NULL,first character varying NOT NULL,second character varying NOT NULL,action character varying NOT NULL,result character varying NOT NULL DEFAULT '?'::character varying,in_progress integer NOT NULL DEFAULT 0,main integer NOT NULL DEFAULT 0,id_main integer NOT NULL,PRIMARY KEY (id_sub));")
	if err != nil {
		panic(err)
	}

	_, err = db.Query("CREATE TABLE IF NOT EXISTS public.times(times_id integer NOT NULL GENERATED BY DEFAULT AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),operation character varying(12) NOT NULL,seconds integer NOT NULL,PRIMARY KEY (times_id));")
	if err != nil {
		panic(err)
	}

	db.Exec("INSERT INTO public.times (times_id, operation, seconds) VALUES (1, '+', 10) ON CONFLICT (times_id) DO NOTHING")
	db.Exec("INSERT INTO public.times (times_id, operation, seconds) VALUES (2, '-', 10) ON CONFLICT (times_id) DO NOTHING")
	db.Exec("INSERT INTO public.times (times_id, operation, seconds) VALUES (3, '*', 10) ON CONFLICT (times_id) DO NOTHING")
	db.Exec("INSERT INTO public.times (times_id, operation, seconds) VALUES (4, '/', 10) ON CONFLICT (times_id) DO NOTHING")
	db.Exec("INSERT INTO public.times (times_id, operation, seconds) VALUES (5, 'server crash', 10) ON CONFLICT (times_id) DO NOTHING")
}

func GetExpression(id int) *[]Expression {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from expressions where expression_id = $1", id)
	if err != nil {
		panic(err)
	}
	expressions := []Expression{}

	for rows.Next() {
		exp := Expression{}
		err := rows.Scan(&exp.Id, &exp.Expression, &exp.Status, &exp.Result)
		if err != nil {
			fmt.Println(err)
			continue
		}
		expressions = append(expressions, exp)
	}
	return &expressions
}

func GetExpressions() *[]Expression {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from expressions")
	if err != nil {
		panic(err)
	}
	expressions := []Expression{}

	for rows.Next() {
		exp := Expression{}
		err := rows.Scan(&exp.Id, &exp.Expression, &exp.Status, &exp.Result)
		if err != nil {
			fmt.Println(err)
			continue
		}
		expressions = append(expressions, exp)
	}
	return &expressions
}

func SendExpression(expression string) int {
	var id, status int

	if ok := CheckExp(expression); ok {
		status = 200
	} else {
		status = 400
	}
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.QueryRow("insert into expressions (expression_string, status) values ($1, $2) returning expression_id", expression, status).Scan(&id)

	if status == 200 {
		go SplitExp(id)
	}

	return id
}

func CheckExp(exp string) bool {
	exp = strings.ReplaceAll(exp, " ", "")
	re := regexp.MustCompile(`^[0-9\+\-\*\/]+$`)
	if !re.MatchString(exp) || !strings.ContainsAny(exp, "+-*/") || !strings.ContainsAny(exp, "0123456789") {
		return false
	}
	runes := []rune(exp)
	if strings.Contains("+-*/", string(runes[0])) || strings.Contains("+-*/", string(runes[len(runes)-1])) {
		return false
	}
	for i := 1; i < len(runes)-2; i++ {
		if strings.Contains("+-*/", string(runes[i])) {
			if !strings.Contains("0123456789", string(runes[i-1])) || !strings.Contains("0123456789", string(runes[i+1])) {
				return false
			}
		} else if !strings.Contains("0123456789", string(runes[i])) {
			return false
		}
	}
	return true
}

func SplitExp(id int) {
	expression := []string{}
	exp := *GetExpression(id)
	runes := []rune(exp[0].Expression)
	s := ""
	for i := 0; i < len(runes); i++ {
		if strings.Contains("0123456789", string(runes[i])) {
			s += string(runes[i])
		} else {
			expression = append(expression, s)
			s = ""
			expression = append(expression, string(runes[i]))
		}
	}
	expression = append(expression, s)
	result := [][]string{}
	e1 := []string{}
	pred := ""
	if slices.Contains(expression, "*") || slices.Contains(expression, "/") {
		for i := 1; i < len(expression); i += 2 {
			if expression[i] == "*" || expression[i] == "/" {
				if pred != "" {
					pred1 := pred
					pred = fmt.Sprintf("%dx.%d", i, id)
					result = append(result, []string{pred, pred1, expression[i], expression[i+1]})
				} else {
					pred = fmt.Sprintf("%dx.%d", i, id)
					result = append(result, []string{pred, expression[i-1], expression[i], expression[i+1]})
				}
			} else {
				if pred != "" {
					e1 = append(e1, pred)
				} else {
					e1 = append(e1, expression[i-1])
				}
				pred = ""
				e1 = append(e1, expression[i])
			}
		}
		if pred != "" {
			e1 = append(e1, pred)
		} else {
			e1 = append(e1, expression[len(expression)-1])
		}
	} else {
		e1 = expression
	}
	pred = ""
	if slices.Contains(expression, "+") || slices.Contains(expression, "-") {
		for i := 1; i < len(e1); i += 2 {
			if pred != "" {
				pred1 := pred
				pred = fmt.Sprintf("%dy.%d", i, id)
				result = append(result, []string{pred, pred1, e1[i], e1[i+1]})
			} else {
				pred = fmt.Sprintf("%dy.%d", i, id)
				result = append(result, []string{pred, e1[i-1], e1[i], e1[i+1]})
			}
			pred = fmt.Sprintf("%dy.%d", i, id)
		}
	}
	AddSubExpression(result, id)
}

func SetExpressionStatus(id int, status int) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update expressions set status = $2 where expression_id = $1", id, status)
}

func SetExpressionResult(id int, result string) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update expressions set result = $2 where expression_id = $1", id, result)
}

func AddSubExpression(exps [][]string, id int) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	for _, exp := range exps[:len(exps)-1] {
		db.Exec("insert into sub_exps (id_sub, first, second, action, result, id_main) values ($1, $2, $3, $4, '?', $5)", exp[0], exp[1], exp[3], exp[2], id)
	}
	db.Exec("insert into sub_exps (id_sub, first, second, action, result, main, id_main) values ($1, $2, $3, $4, '?', $5, $6)",
		exps[len(exps)-1][0], exps[len(exps)-1][1], exps[len(exps)-1][3], exps[len(exps)-1][2], id, id)
}

func SubStarts() {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("update sub_exps set in_progress = 0 where in_progress = 1 and result = '?'")

	rows, err := db.Query("select * from sub_exps where in_progress = -1 and main != 0")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		exp := SubExpression{}
		rows.Scan(&exp.Id, &exp.Id_sub, &exp.First, &exp.Second, &exp.Action, &exp.Result, &exp.In_progress, &exp.Main, &exp.Id_main)
		SetExpressionResult(exp.Id_main, exp.Result)
		db.Exec("delete from sub_exps where id_main = $1", exp.Id_main)
	}
}

func GetSubExpressionsForWorks() *[]SubExpression {
	sub_exps := []SubExpression{}

	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from sub_exps where in_progress = 0")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		exp := SubExpression{}
		rows.Scan(&exp.Id, &exp.Id_sub, &exp.First, &exp.Second, &exp.Action, &exp.Result, &exp.In_progress, &exp.Main, &exp.Id_main)
		sub_exps = append(sub_exps, exp)
	}
	return &sub_exps
}

func GetSubExpression(id_sub string) *SubExpression {
	sub_exps := []SubExpression{}

	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from sub_exps where id_sub = $1", id_sub)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		exp := SubExpression{}
		err := rows.Scan(&exp.Id, &exp.Id_sub, &exp.First, &exp.Second, &exp.Action, &exp.Result, &exp.In_progress, &exp.Main, &exp.Id_main)
		if err != nil {
			continue
		}
		sub_exps = append(sub_exps, exp)
	}
	return &sub_exps[0]
}

func CancelExp(id_main int) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("update sub_exps set in_progress = -1 where id_sub = $1", id_main)
}

func SetSubStatus(id_sub string, status int) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("update sub_exps set in_progress = $2 where id_sub = $1", id_sub, status)
}

func SetSubResult(id_sub string, result string) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update sub_exps set result = $2 where id_sub = $1", id_sub, result)
}

func GetTimes() *[]Time {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from times")
	if err != nil {
		panic(err)
	}
	times := []Time{}
	defer rows.Close()

	for rows.Next() {
		time := Time{}
		err := rows.Scan(&time.Id, &time.Operation, &time.Seconds)
		if err != nil {
			fmt.Println(err)
			continue
		}
		times = append(times, time)
	}
	return &times
}

func GetTime(action string) int {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from times where operation = $1", action)
	if err != nil {
		panic(err)
	}
	times := []Time{}
	defer rows.Close()

	for rows.Next() {
		time := Time{}
		err := rows.Scan(&time.Id, &time.Operation, &time.Seconds)
		if err != nil {
			fmt.Println(err)
			continue
		}
		times = append(times, time)
	}
	return times[0].Seconds
}

func ManageTimes(NewTimes []string) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	for i := 1; i <= 5; i++ {
		if n, _ := strconv.Atoi(NewTimes[i-1]); n < 1 {
			NewTimes[i-1] = "1"
		}
		db.Exec("update times set seconds = $1 where times_id = $2", NewTimes[i-1], i)
	}
}

func GetServers() *[]Server {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("select * from servers")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	servers := []Server{}

	for rows.Next() {
		server := Server{}
		err := rows.Scan(&server.Id, &server.Server_status, &server.Goroutines, &server.Next_ping)
		if err != nil {
			fmt.Println(err)
			continue
		}
		servers = append(servers, server)
	}
	return &servers
}

func SetServerPing(id int, status string, goroutines int, next_ping string) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update servers set server_status = $2, goroutines = $3, next_ping = $4 where server_id = $1", id, status, goroutines, next_ping)
}

func SetServerStatus(id int, status string) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update servers set server_status = $2 where server_id = $1", id, status)
}

func AddServer() int {
	var id int
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	t := *GetTimes()
	db.QueryRow("insert into servers (server_status, next_ping) values ('idle', $1) returning server_id",
		time.Now().Add(time.Duration(t[4].Seconds)).Format("2006-01-02 15:04:05")).Scan(&id)

	return id
}

func DeleteServers() {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("delete from servers where server_status = 'dead'")
}

func SetGoroutines(id_server, goroutines int) {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("update servers set goroutines = $2 where server_id = $1", id_server, goroutines)
}

func GetGoroutines(id_server int) int {
	connStr := "user=postgres password=qwerty1234 dbname=microservice sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, _ := db.Query("select * from servers where server_id = $1", id_server)
	defer rows.Close()
	servers := []Server{}

	for rows.Next() {
		server := Server{}
		rows.Scan(&server.Id, &server.Server_status, &server.Goroutines, &server.Next_ping)
		servers = append(servers, server)
	}
	return servers[0].Goroutines
}
