package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	example "temporal102/samples/age-estimation"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, example.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(example.EstimateAge)
	w.RegisterActivity(example.RetrieveEstimate)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
