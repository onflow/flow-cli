package flix

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetInteractionTemplateByName(templateName string) (string, error) {
	baseUrl := "https://flix.flow.com/v1/templates"
	url := fmt.Sprintf("%s?name=%s", baseUrl, templateName)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: error while closing the response body: %v", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	sb := string(body)

	return sb, nil
}
