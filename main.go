package main

import (
	"log"

	"github.com/clockworksoul/cog2/context"
	"github.com/clockworksoul/cog2/relay"
)

func main() {
	log.Printf("Starting Cog2 version %s\n", context.CogVersion)

	relay.StartListening()

	<-make(chan struct{})
}

// func main() {
// 	imageName := "bfirsh/reticulate-splines"
// 	timeout := 5 * time.Second

// 	worker, err := worker.NewWorker(imageName, timeout)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	logs, err := worker.Start()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	for s := range logs {
// 		fmt.Println(s)
// 	}
// }
