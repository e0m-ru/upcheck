package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/e0m-ru/upcheck/config"
	"github.com/e0m-ru/upcheck/model"
)

func main() {
	C, err := config.LoadConfig("./config.yaml")
	if err != nil {
		log.Print(err)
	}
	chekr := model.NewUpChecker(time.Duration(C.Interval)*time.Second, C.Alerts)

	for _, site := range *C.Sites {
		chekr.Sites[site.URL] = &model.SiteStats{URL: site.URL}
	}

	go chekr.Start()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down monitor...")
	chekr.Stop()
}
