package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

type data struct {
	Repo               string `json:"repo"`
	ServiceFileChanged bool   `json:"service_file_changed"`
}

func registerRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var data data

		err := decoder.Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// if the service file has been modified, we need to replace it and reload the systemctl daemon
		var cmd *exec.Cmd
		if data.ServiceFileChanged {
			fmt.Println("service file changed, reloading systemctl daemon")
			cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo git pull && sudo cp %s.service /etc/systemd/system/%s.service && sudo systemctl daemon-reload && sudo systemctl restart %s.service", data.Repo, data.Repo, data.Repo))
		} else {
			cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo git pull && sudo systemctl restart %s.service", data.Repo))
		}

		cmd.Dir = fmt.Sprintf("/home/cloud-user/%s", data.Repo)
		go cmd.Run()
		w.WriteHeader(http.StatusOK)
		// if err != nil {
		// 	http.Error(w, "output: "+string(out)+", error: "+err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	})

	return mux
}

func main() {
	server := http.Server{
		Addr:         ":9000",
		Handler:      registerRoutes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	server.ListenAndServe()
}
