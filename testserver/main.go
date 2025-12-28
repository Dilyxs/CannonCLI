package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type Data struct {
	N       int    `json:"n"`
	Content string `json:"content"`
}

func ReturnData(w http.ResponseWriter, r *http.Request) {
	time.Sleep(500 * time.Millisecond)
	// make 20 percent of Response incorrect
	w.Header().Set("Content-type", "application/json")
	n := rand.Int31n(5)
	if n == 0 {
		data := Data{int(n), "I don't love you!"}
		w.WriteHeader(400)
		stuff, _ := json.Marshal(data)
		w.Write(stuff)
	} else {
		data := Data{int(n), "I  love you!"}
		stuff, _ := json.Marshal(data)
		w.Write(stuff)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { ReturnData(w, r) })
	fmt.Println("running server")
	http.ListenAndServe(":8080", nil)
}
