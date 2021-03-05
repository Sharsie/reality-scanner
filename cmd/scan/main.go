package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
	"github.com/Sharsie/reality-scanner/cmd/scan/slack"
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
		l.Debug("%s", err)
		log.Fatal("Failed initialization")
	}

	l.Debug("Initialized with %d realities", len(realities))
	slack.Send(&l, fmt.Sprintf("<@U01C5RDRAKZ> <@U01CH3DQ9EH> Restartoval se scanner \"%s\", začínáme s %d realitama", config.ScannerID, len(realities)))

	// gracefulStop is a channel of os.Signals that we will watch for -SIGTERM
	var gracefulStop = make(chan os.Signal)

	// watch for SIGTERM and SIGINT from the operating system, and notify the app on
	// the gracefulStop channel
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	// launch a worker whose job it is to always watch for gracefulStop signals
	go func() {
		// wait for our os signal to stop the app
		// on the graceful stop channel
		// this goroutine will block until we get an OS signal
		<-gracefulStop
		slack.Send(&l, fmt.Sprintf("<@U01C5RDRAKZ> <@U01CH3DQ9EH> Nějak se stalo, že jsem se vypnul, omlouvám se, měl jsem v zásobě %d realit a byl jsem \"%s\"", len(realities), config.ScannerID))
		os.Exit(0)
	}()

	for {
		select {
		case <-statusCheck.C:
			l.Debug("Polling status...")
			newRealities, deletedRealities, err := sreality.CheckForNew(&l)

			if err == nil {
				for _, item := range newRealities {
					slack.SendNewReality(&l, item)

					l.Debug("Reality %d is new", item.Id)
				}

				for _, item := range deletedRealities {
					slack.Send(&l, fmt.Sprintf("Smazali nám věc: %s - %s, Cena %d,-, id inzerátu: %s", item.Title, item.Place, item.Price, item.Id))
					l.Debug("Reality %d was deleted", item.Id)
				}

				if len(newRealities) == 0 && len(deletedRealities) == 0 {
					l.Debug("No new realities were found")
				}
			}

		}
	}

}
