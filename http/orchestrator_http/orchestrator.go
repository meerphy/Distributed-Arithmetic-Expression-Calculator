package orchestrator_http

import (
	"encoding/json"
	"html/template"
	authorization "microservice/internal/autorization"
	"microservice/internal/orchestrator"
	"microservice/pkg/database"
	"strconv"
	"time"

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

func loginPage(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err == nil {
		http.Redirect(w, r, "/expressions", http.StatusFound)
	} else if r.Method == "POST" {
		login := r.FormValue("name")
		password := r.FormValue("password")
		id := database.Login(login, password)
		if id != 0 {
			token := authorization.MakeToken(id, login)
			http.SetCookie(w, &http.Cookie{
				Name:     "user",
				Value:    token,
				HttpOnly: true,
				Expires:  time.Now().Add(5 * time.Minute),
			})
			http.Redirect(w, r, "/expressions", http.StatusFound)
		}
	}
	tmp_login.Execute(w, nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func viewExpressions(w http.ResponseWriter, r *http.Request) {
	cookieData, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		id := r.URL.Query().Get("id")
		token := cookieData.Value
		user_id := authorization.GetTokenValue(token).ID
		if id_, err := strconv.Atoi(id); err == nil {
			expressions := ViewExpressions{Title: "All expressions:", Expressions: *database.GetExpression(id_, user_id)}
			tmp_view_expressions.Execute(w, expressions)
		} else {
			expressions := ViewExpressions{Title: "All expressions:", Expressions: *database.GetExpressions(user_id)}
			tmp_view_expressions.Execute(w, expressions)
		}
	}
}

func sendExpression(w http.ResponseWriter, r *http.Request) {
	cookieData, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else if r.Method == "POST" {
		expression := r.FormValue("input")
		token := cookieData.Value
		user_id := authorization.GetTokenValue(token).ID
		data := ViewId{Title: "Your unique id:", Id: database.SendExpression(expression, user_id)}
		tmp_send_expression.Execute(w, data)
	} else {
		tmp_send_expression.Execute(w, nil)
	}
}

func manageTime(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else if r.Method == "POST" {
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
	_, err := r.Cookie("user")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	} else {
		servers := ViewServers{Title: "All servers:", Servers: *database.GetServers()}
		tmp_view_servers.Execute(w, servers)
	}
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

var tmp_view_expressions, tmp_send_expression, tmp_manage_time, tmp_view_servers, tmp_login *template.Template

func runServerHTTP() {
	tmp_view_expressions = template.Must(template.ParseFiles("front/templates/view_expressions.html"))
	http.HandleFunc("/expressions", viewExpressions)
	tmp_send_expression = template.Must(template.ParseFiles("front/templates/send_expression.html"))
	http.HandleFunc("/send", sendExpression)
	tmp_manage_time = template.Must(template.ParseFiles("front/templates/manage_time.html"))
	http.HandleFunc("/manage", manageTime)
	tmp_view_servers = template.Must(template.ParseFiles("front/templates/view_servers.html"))
	http.HandleFunc("/servers", viewServers)
	tmp_login = template.Must(template.ParseFiles("front/templates/login.html"))
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/request_expression", requesting)

	http.ListenAndServe(":8080", nil)
}

func RunOrchestrator() {
	database.Create()
	database.SubStarts()
	go orchestrator.CheckPings()
	go orchestrator.GarbageCollector()
	go runServerHTTP()
}
