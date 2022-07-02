package main

import (
	"log"
	"net/http"

	"github.com/8run0/kllla2/backend/pkg/api"
	"github.com/8run0/kllla2/backend/pkg/api/tasks"
	tasksdb "github.com/8run0/kllla2/backend/pkg/db/tasks"
)

func main() {

	indexResource := api.NewResource()
	taskResource := tasks.NewResource(tasksdb.NewDB("./tasks.db"))

	indexResource.Mux.Mount(tasks.Path, taskResource.Routes())
	svr := api.NewServer(indexResource)

	port := "8080"
	log.Printf("Starting up on http://localhost:%s", port)

	http.ListenAndServe(":3333", svr)
}
