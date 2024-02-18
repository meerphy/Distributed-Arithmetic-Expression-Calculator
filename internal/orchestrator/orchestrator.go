package orchestrator

import (
	"microservice/pkg/database"
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
	Id     int    `json:"id"`
	Result string `json:"result"`
}

func CheckPings() {
	for {
		servers := *database.GetServers()
		for _, server := range servers {
			targetTime, _ := time.Parse("2006-01-02 15:04:05", server.Next_ping)
			currentTime, _ := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
			if targetTime.Before(currentTime) {
				database.SetServerPing(server.Id, "dead", 0, server.Next_ping)
			}
		}
		t := *database.GetTimes()
		time.Sleep(time.Duration(t[4].Seconds) * time.Second)
	}
}

func Distribution() (string, int, int, string, bool) {
	data := *database.GetSubExpressionsForWorks()
	var flag int
	for i := 0; i < len(data); i++ {
		flag = 0
		n1, err1 := strconv.Atoi(data[i].First)
		if err1 != nil {
			n := *database.GetSubExpression(data[i].First)
			if n.Result == "?" {
				flag = 1
			} else {
				n1, _ = strconv.Atoi(n.Result)
			}
		}
		n2, err2 := strconv.Atoi(data[i].Second)
		if err2 != nil {
			n := *database.GetSubExpression(data[i].Second)
			if n.Result == "?" {
				flag = 1
			} else {
				n2, _ = strconv.Atoi(n.Result)
			}
		}
		if flag != 1 {
			database.SetSubStatus(data[i].Id_sub, 1)
			go WaitSubResult(data[i].Id_sub, data[i].Action)
			return data[i].Id_sub, n1, n2, data[i].Action, true
		}
	}
	return "", 0, 0, "", false
}

func WaitSubResult(id_sub, action string) {
	for {
		time.Sleep(time.Duration(database.GetTime(action)+3) * time.Second)
		if ans := *database.GetSubExpression(id_sub); ans.Result == "'?'" {
			database.SetSubStatus(id_sub, 0)
		} else {
			database.SetSubStatus(id_sub, -1)
			id := *database.GetSubExpression(id_sub)
			if id.Main != 0 {
				database.SetExpressionResult(id.Main, ans.Result)
			}
			break
		}
	}
}

func GarbageCollector() {
	for {
		time.Sleep(time.Second)
		database.DeleteServers()
		time.Sleep(2000 * time.Second)
	}
}
