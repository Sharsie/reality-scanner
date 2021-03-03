package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
	"github.com/Sharsie/reality-scanner/cmd/scan/reality"
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

	if config.Debug {
		images := make([]string, 0)
		images = append(images, "https://d18-a.sdn.cz/d_18/c_img_gY_K/0lsdFo.jpeg?fl=res,400,300,3%7Cshr,,20%7Cjpg,90")
		images = append(images, "https://d18-a.sdn.cz/d_18/c_img_gT_K/xMPdF2.jpeg?fl=res,400,300,3%7Cshr,,20%7Cjpg,90")
		slack.SendNewReality(&l, reality.KnownReality{
			Deletable: false,
			Id:        "x",
			Images:    images,
			IsNew:     false,
			Link:      "https://seznam.cz",
			Place:     "Some nice place",
			Price:     50000,
			Title:     "Je to fakt hezkly no",
		})

		log.Fatal("")
	}

	statusCheck := time.NewTicker(config.StatusCheckPeriod)

	realities, err := sreality.Initialize(&l)

	if err != nil {
		l.Debug("%s", err)
		log.Fatal("Failed initialization")
	}

	l.Debug("Initialized with %d realities", len(realities))
	slack.Send(&l, fmt.Sprintf("<@U01C5RDRAKZ> <@U01CH3DQ9EH> Restartoval se scanner, začínáme s %d realitama", len(realities)))

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
