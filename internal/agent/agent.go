package agent

import (
	"encoding/json"
	"microservice/pkg/database"
	"net/http"
	"strconv"
	"time"
)

type Response_agent struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Goroutines int    `json:"goroutines"`
	Next_ping  string `json:"last_ping"`
}

type Response_result struct {
	Id_sub string `json:"id_sub"`
	Result string `json:"result"`
}

type Response_sub struct {
	Id_sub string
	First  int
	Second int
	Action string
	Exists bool
}

func calculate(data Response_sub, id_server int) {
	database.SetServerStatus(id_server, "working")
	database.SetGoroutines(id_server, database.GetGoroutines(id_server)+1)
	times := *database.GetTimes()
	result := 0
	if data.Action == "+" {
		time.Sleep(time.Duration(times[0].Seconds) * time.Second)
		result = data.First + data.Second
	} else if data.Action == "-" {
		time.Sleep(time.Duration(times[1].Seconds) * time.Second)
		result = data.First - data.Second
	} else if data.Action == "*" {
		time.Sleep(time.Duration(times[2].Seconds) * time.Second)
		result = data.First * data.Second
	} else if data.Action == "/" {
		time.Sleep(time.Duration(times[3].Seconds) * time.Second)
		if data.Second != 0 {
			result = data.First / data.Second
		} else {
			sub := *database.GetSubExpression(data.Id_sub)
			database.SetExpressionStatus(sub.Id_main, 400)
			database.CancelExp(sub.Id_main)
		}
	}
	database.SetGoroutines(id_server, database.GetGoroutines(id_server)-1)
	if database.GetGoroutines(id_server) == 0 {
		database.SetServerStatus(id_server, "idle")
	}
	database.SetSubResult(data.Id_sub, strconv.Itoa(result))
}

func requestExpression() Response_sub {
	resp, _ := http.Get("http://localhost:8080/request_expression")
	var data Response_sub
	json.NewDecoder(resp.Body).Decode(&data)
	return data
}

func pinging(id_server int) {
	for {
		t := *database.GetTimes()
		next_ping := time.Now().Add(time.Duration(t[4].Seconds) * time.Second).Format("2006-01-02 15:04:05")
		if g := database.GetGoroutines(id_server); g > 0 {
			database.SetServerPing(id_server, "working", g, next_ping)
		} else {
			database.SetServerPing(id_server, "idle", 0, next_ping)
		}
		time.Sleep(time.Duration(int(t[4].Seconds/4*3)) * time.Second)
	}
}

func agent(goroutines int) {
	var id_server int = database.AddServer()
	go pinging(id_server)
	for {
		time.Sleep(time.Second)
		if database.GetGoroutines(id_server) < goroutines {
			data := requestExpression()
			if data.Exists {
				go calculate(data, id_server)
			}
		}
	}
}

func RunAgent(n int) {
	go agent(n)
}
