
package main

import (
	"log"
	"net/http"

	"time"

	"os"

	"encoding/json"
	"fmt"

	"math/rand"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	fabclient "github.com/securekey/marbles-perf/fabric-client"
	"github.com/securekey/marbles-perf/utils"
	"github.com/spf13/viper"
)

const (
	ConsortiumChannelID = "consortium"
	MarblesCC           = "marblescc"
)

var fc fabclient.Client
var logger = logging.MustGetLogger("marbles-service")

func main() {

	if len(os.Args) < 2 {
		log.Fatal("expecting configuration file as first argument")
	}
	cfgFile := os.Args[1]

	err := SetupViper(cfgFile)
	if err != nil {
		log.Fatalf("error setting up viper using config file and environmental variables: %v ", err)
	}

	utils.InitLogger()

	fc, err = fabclient.NewClient()
	if err != nil {
		log.Fatalf("failed to initialize fabric client: %s", err)
	}

	r := mux.NewRouter()
	// ping
	r.HandleFunc("/api/hello", handleHello)
	// CRUD

	r.HandleFunc("/api/transfer", transfer).Methods(http.MethodPost)
	r.HandleFunc("/api/clear_marbles", clearMarbles).Methods(http.MethodPost)
	r.HandleFunc("/api/change",change).Methods(http.MethodPost)
	r.HandleFunc("/api/delete/{id}",delete).Methods(http.MethodGet)
	r.HandleFunc("/api/expert/{id}", getExpert).Methods(http.MethodGet)
	r.HandleFunc("/api/institution/{id}", getInstitution).Methods(http.MethodGet)
	r.HandleFunc("/api/city/{id}", getCity).Methods(http.MethodGet)
	r.HandleFunc("/api/demand/{id}", getDemand).Methods(http.MethodGet)
	r.HandleFunc("/api/scheme/{id}", getScheme).Methods(http.MethodGet)
	r.HandleFunc("/api/patent/{id}", getPatent).Methods(http.MethodGet)
	r.HandleFunc("/api/paper/{id}", getPaper).Methods(http.MethodGet)
	r.HandleFunc("/api/transfer/{id}", getTransfer).Methods(http.MethodGet)
	r.HandleFunc("/api/gethistory/{id}", get_history).Methods(http.MethodGet)
	r.HandleFunc("/api/read_everything",read_everything).Methods(http.MethodGet)
	r.HandleFunc("/api/read_allmarble/{id}",read_allmarble).Methods(http.MethodGet)
	r.HandleFunc("/api/sign_in", sign_in).Methods(http.MethodPost)
	r.HandleFunc("/api/sign_up", sign_up).Methods(http.MethodPost)
	r.HandleFunc("/api/change_pwd", change_pwd).Methods(http.MethodPost)
	r.HandleFunc("/api/able/{id}",able).Methods(http.MethodGet)

	// batch (random) transfers
	//r.HandleFunc("/api/batch_run", initBatchTransfers).Methods(http.MethodPost)
	//r.HandleFunc("/api/batch_run/{id}", fetchBatchResults).Methods(http.MethodGet)

	// Seed the random generator so we get different values each time
	rand.Seed(time.Now().UTC().UnixNano())

	srv := &http.Server{
		Handler:      r,
		Addr:         viper.GetString("http.server.address"),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World\n"))
}

func writeErrorResponse(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args)
	}
	w.Write([]byte(fmt.Sprintf(`{error: "%s"}`, msg)))
	logger.Infof("error: %s", msg)
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	jsonStr, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "failed to JSON marshal response: %s", err)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonStr)
}
