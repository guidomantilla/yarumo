package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	serviceworkflow "temporal101/demos/service-workflow"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "greeting-tasks", worker.Options{})

	w.RegisterWorkflow(serviceworkflow.GreetSomeone)
	w.RegisterActivity(serviceworkflow.GreetInSpanish)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
