package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
	"github.com/Sharsie/reality-scanner/cmd/scan/sreality"
	"github.com/Sharsie/reality-scanner/cmd/scan/version"
)

func main() {
	l := logger.Log{}
	if version.Tag != "" {
		fmt.Printf("Running version %s\n", version.Tag)

		l.Debug("Built at %s from commit hash '%s'", version.BuildTime, version.Commit)
	}

	statusCheck := time.NewTicker(config.StatusCheckPeriod)

	realities, err := sreality.Initialize(&l)

	if err != nil {
		log.Fatal("Failed initialization")
	}

	l.Debug("Initialized with %d realities", len(realities))

	for {
		select {
			case <-statusCheck.C:
				l.Debug("Polling status...")
				newRealities, deletedRealities, err := sreality.CheckForNew(&l)

				if err == nil {
					for _, reality := range newRealities {
						l.Debug("Reality %d is new", reality.Id)
					}

					for _, reality := range deletedRealities {
						l.Debug("Reality %d was deleted", reality.Id)
					}
				}
		}
	}

}
