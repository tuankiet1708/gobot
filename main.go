package main

import (
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/gorilla/mux"
)

func main() {
	// registerGreetingnMenu()
	// return

	r := mux.NewRouter()
	r.HandleFunc("/", chatbotHandler)

	// info
	log.Info("Started GoBot")

	// start server
	// if err := http.ListenAndServe(":8080", r); err != nil {
	// 	log.Fatal(err.Error())
	// }
	// start server
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), r); err != nil {
		log.Fatal(err.Error())
	}
}

func chatbotHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		verifyWebhook(w, r)
	case "POST":
		processWebhook(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Errorf("Không hỗ trợ phương thức HTTP %v", r.Method)
	}
}
