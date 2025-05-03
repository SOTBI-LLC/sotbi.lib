package dadata

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	innValidator "github.com/tit/go-inn-validator"
)

type Suggestions struct {
	httpClient *resty.Client
}

func New(apiKey string) *Suggestions {
	return &Suggestions{
		httpClient: resty.New().
			SetAuthScheme("token").
			SetAuthToken(apiKey).
			SetBaseURL("http://suggestions.dadata.ru/suggestions/"),
	}
}

func (c *Suggestions) Get(inn string) ([]Suggestion, error) {
	isPrivate, _ := innValidator.IsPrivatePersonInnValid(inn)
	isLegal, _ := innValidator.IsLegalPersonInnValid(inn)

	if !isPrivate && !isLegal {
		return nil, errors.New("provided INN is not valid")
	}

	resp, err := c.httpClient.R().
		SetHeader("Accept", "application/json").
		SetBody(DadataRequest{
			Query: inn,
		}).
		Post("api/4_1/rs/findById/party")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: %s", resp.Status())
	}

	var dadataResponse DadataResponse
	if err = json.Unmarshal(resp.Body(), &dadataResponse); err != nil {
		return nil, fmt.Errorf("error: %s", err)
	}

	return dadataResponse.Suggestions, nil
}
