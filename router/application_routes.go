package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/doug-benn/go-json-api/database"
)

func HandleHelloWorld(log *slog.Logger, dbInterface *database.PostgresInterface) http.HandlerFunc {

	// type responseBody struct {
	// 	Message string `json:"Message"`
	// 	Uptime  string `json:"Uptime"`
	// }

	//res := responseBody{Message: "Hello World"}

	// ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	// defer cancel()

	//up := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		log.Info("Getting data from DB")

		data, _ := dbInterface.GetAll()
		fmt.Printf("Data has value %+v\n", data)
		// data2, _ := dbInterface.Get("1")
		// fmt.Printf("Data2 has value %+v\n", data2)
		data2, _ := dbInterface.GetAccountByID(1)
		fmt.Printf("Data3 has value %+v\n", data2)
		// data4, _ := dbInterface.GetAccounts()
		// fmt.Printf("Data4 has value %+v\n", data4)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		//res.Uptime = time.Since(up).String()
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
