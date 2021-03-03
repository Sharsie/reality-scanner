package sreality

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Sharsie/reality-scanner/cmd/scan/config"
	"github.com/Sharsie/reality-scanner/cmd/scan/logger"
	"github.com/Sharsie/reality-scanner/cmd/scan/reality"
)

type Estate struct {
	IsNew    bool   `json:"new"`
	Id       int64  `json:"hash_id"`
	Locality string `json:"locality"`
	Links    struct {
		Images []struct {
			Href string `json:"href"`
		} `json:"images"`
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Name  string `json:"name"`
	Paid  int    `json:"paid_logo"`
	Price int    `json:"price"`
	Seo   struct {
		Locality string `json:"locality"`
	} `json:"seo"`
}

type responseData struct {
	Embedded struct {
		Estates []Estate `json:"estates"`
	} `json:"_embedded"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
	Size    int `json:"result_size"`
}

var realities reality.KnownRealities

func Initialize(l *logger.Log) (reality.KnownRealities, error) {
	l.Debug("Initializing sreality")

	client := http.Client{Timeout: 5 * time.Second}

	var err error
	realities, err = fetchRealities(l, client)

	return realities, err
}

func CheckForNew(l *logger.Log) (reality.KnownRealities, reality.KnownRealities, error) {
	l.Debug("Checking sreality")

	client := http.Client{Timeout: 5 * time.Second}
	currentRealities, err := fetchRealities(l, client)

	if err != nil {
		return nil, nil, err
	}

	deletedRealities := make(reality.KnownRealities)

	for id, item := range realities {
		_, exists := currentRealities[id]

		if exists == false && item.Deletable {
			deletedRealities[id] = item
			delete(realities, id)
			continue
		}

		// The reality already exists
		delete(currentRealities, id)
	}

	for id, item := range currentRealities {
		realities[id] = item
	}

	return currentRealities, deletedRealities, nil
}

func fetchRealities(l *logger.Log, client http.Client) (reality.KnownRealities, error) {
	estates := make([]Estate, 0)

	i := 0
	for {
		i++
		data, err := recursiveFetch(l, client, i)

		if err != nil {
			return nil, err
		}

		for _, estate := range data.Embedded.Estates {
			estates = append(estates, estate)
		}

		if data.PerPage*i > data.Size {
			break
		}
	}

	return createRealities(estates), nil
}

func recursiveFetch(l *logger.Log, client http.Client, page int) (*responseData, error) {
	pagedUrl := fmt.Sprintf("%s&page=%d", config.SrealityEndpoint, page)

	response, err := client.Get(pagedUrl)

	if err != nil {
		l.Debug("%s", err)
		return nil, errors.New("Could not initialize sreality")
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		l.Debug("%s", err)
		return nil, errors.New("Could not decode sreality response body")
	}

	var data responseData

	err = json.Unmarshal(body, &data)

	if err != nil {
		l.Debug("Error> %v", err)
		return nil, errors.New("Could not unmarshal sreality response")
	}

	return &data, nil
}

func createRealities(estates []Estate) reality.KnownRealities {
	currentRealities := make(reality.KnownRealities)

	for _, estate := range estates {
		images := make([]string, 0)

		for _, image := range estate.Links.Images {
			images = append(images, image.Href)
		}

		item := reality.KnownReality{
			Deletable: estate.Paid == 0,
			Id:        strconv.FormatInt(estate.Id, 10),
			Images:    images,
			IsNew:     estate.IsNew,
			Link:      fmt.Sprintf("https://www.sreality.cz/detail/prodej/pozemek/bydleni/%s/%s", estate.Seo.Locality, strconv.FormatInt(estate.Id, 10)),
			Place:     estate.Locality,
			Price:     estate.Price,
			Title:     estate.Name,
		}

		currentRealities[strconv.FormatInt(estate.Id, 10)] = item
	}

	return currentRealities
}
