package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

var client_id = getEnv("CLIENT_ID")
var client_secret = getEnv("CLIENT_SECRET")
var tenant_id = getEnv("TENANT_ID")
var scope_client_id_back = getEnv("SCOPE_CLIENT_ID_BACK")
var graphql_url = getEnv("GRAPHQL_URL")
var graphql_code = getEnv("GRAPHQL_CODE")

func requestGraphql(jsonData map[string]interface{}) ([]byte, error) {
	payload, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Printf("Erro ao ler o request: %s\n", err)
	}

	oauthConfig := &clientcredentials.Config{
		ClientID:     client_id,
		ClientSecret: client_secret,
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant_id),
		Scopes: []string{
			fmt.Sprintf("api://%s/.default", scope_client_id_back),
		},
	}

	client := oauthConfig.Client(context.Background())
	req, err := http.NewRequest("POST", graphql_url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("NewRequest failed with error %s\n", err)
	}
	req.Header.Add("x-functions-key", graphql_code)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}
