package orchestrator_http

import (
	"encoding/json"
	"html/template"
	"microservice/internal/orchestrator"
	"microservice/pkg/database"
	"strconv"

	"net/http"
)

type ViewId struct {
	Title string
	Id    int
}

type ViewExpressions struct {
	Title       string
	Expressions []database.Expression
}

type ViewTimes struct {
	Title       string
	Expressions []database.Time
}

type ViewServers struct {
	Title   string
	Servers []database.Server
}

type Response_sub struct {
	Id_sub string
	First  int
	Second int
	Action string
	Exists bool
}

func viewExpressions(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id_, err := strconv.Atoi(id); err == nil {
		expressions := ViewExpressions{Title: "All expressions:", Expressions: *database.GetExpression(id_)}
		tmp_view_expressions.Execute(w, expressions)
	} else {
		expressions := ViewExpressions{Title: "All expressions:", Expressions: *database.GetExpressions()}
		tmp_view_expressions.Execute(w, expressions)
	}
}

func sendExpression(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		expression := r.FormValue("input")
		data := ViewId{Title: "Your unique id:", Id: database.SendExpression(expression)}
		tmp_send_expression.Execute(w, data)
	} else {
		tmp_send_expression.Execute(w, nil)
	}
}

func manageTime(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var FieldValues []string
		r.ParseForm()
		for _, values := range r.Form {
			FieldValues = append(FieldValues, values...)
		}
		database.ManageTimes(FieldValues)
	}
	times := ViewTimes{Title: "All times:", Expressions: *database.GetTimes()}
	tmp_manage_time.Execute(w, times)
}

func viewServers(w http.ResponseWriter, r *http.Request) {
	servers := ViewServers{Title: "All servers:", Servers: *database.GetServers()}
	tmp_view_servers.Execute(w, servers)
}

func requesting(w http.ResponseWriter, r *http.Request) {
	id_sub, first, second, action, exists := orchestrator.Distribution()
	response := Response_sub{
		Id_sub: id_sub,
		First:  first,
		Second: second,
		Action: action,
		Exists: exists,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var tmp_view_expressions, tmp_send_expression, tmp_manage_time, tmp_view_servers *template.Template

func runServerHTTP() {
	tmp_view_expressions = template.Must(template.ParseFiles("front/templates/view_expressions.html"))
	http.HandleFunc("/expressions", viewExpressions)
	tmp_send_expression = template.Must(template.ParseFiles("front/templates/send_expression.html"))
	http.HandleFunc("/send", sendExpression)
	tmp_manage_time = template.Must(template.ParseFiles("front/templates/manage_time.html"))
	http.HandleFunc("/manage", manageTime)
	tmp_view_servers = template.Must(template.ParseFiles("front/templates/view_servers.html"))
	http.HandleFunc("/servers", viewServers)
	http.HandleFunc("/request_expression", requesting)

	http.ListenAndServe(":8080", nil)
}

func RunOrchestrator() {
	database.SubStarts()
	go orchestrator.CheckPings()
	go orchestrator.GarbageCollector()
	go runServerHTTP()
}
