package sreality

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
)

type responseData struct {
	_embedded struct {
		estates []struct {
			IsNew bool `json:"new"`
			Id int `json:"hash_id"`
			Locality string `json:"locality"`
		}
	}
}

type knownReality struct {
	Id    int
	IsNew bool
}

type knownRealities map[int]knownReality

var realities knownRealities

func Initialize(l *logger.Log) (knownRealities, error) {
	l.Debug("Initializing sreality")

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(config.SrealityEndpoint)

	if err != nil {
		return nil, errors.New("Could not initialize sreality");
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, errors.New("Could not decode sreality response body")
	}

	var data responseData

	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, errors.New("Could not unmarshal sreality ersponse")
	}

	realities = make(knownRealities)

	for _, estate := range data._embedded.estates {
		reality := knownReality{
			Id: estate.Id,
			IsNew: estate.IsNew,
		}

		realities[estate.Id] = reality
	}

	return realities, nil
}

func CheckForNew(l *logger.Log) (knownRealities, knownRealities, error) {
	l.Debug("Checking sreality")

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(config.SrealityEndpoint)

	if err != nil {
		return nil, nil, errors.New("Could not initialize sreality");
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, nil, errors.New("Could not decode sreality response body")
	}

	var data responseData

	err = json.Unmarshal(body, &data)

	if err != nil {
		return nil, nil, errors.New("Could not unmarshal sreality ersponse")
	}

	currentRealities := make(knownRealities)
	deletedRealities := make(knownRealities)
	newRealities := make(knownRealities)

	for _, estate := range data._embedded.estates {
		reality := knownReality{
			Id: estate.Id,
			IsNew: estate.IsNew,
		}

		currentRealities[estate.Id] = reality

		_, exists := realities[estate.Id]
		if exists == false {
			newRealities[estate.Id] = reality
		}
	}

	for id, reality := range realities {
		_, exists := currentRealities[id]

		if exists == false {
			deletedRealities[id] = reality
			delete(realities, id)
			continue
		}

		// We could check if it changed somehow
		delete(currentRealities, id)
	}

	return newRealities, deletedRealities, nil
}
