package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2/clientcredentials"
)

type proposta struct {
	reqURI   string
	request  map[string]interface{}
	response map[string]interface{}
}

var lista = make(map[int64]*proposta)

func SendData(ctxProposta *CtxProposta, bodyRequest string) {

	if lista[ctxProposta.session] == nil {
		lista[ctxProposta.session] = &proposta{}
	}

	var objmap map[string]interface{}
	if err := json.Unmarshal([]byte(bodyRequest), &objmap); err != nil {
		log.Fatal(err)
	}

	if ctxProposta.tipo == "request" {
		lista[ctxProposta.session].reqURI = ctxProposta.reqURI
		lista[ctxProposta.session].request = objmap
	} else if ctxProposta.tipo == "response" {
		lista[ctxProposta.session].response = objmap

		//TODO tirar o replace quando for para produção
		reqURI := strings.Replace(ctxProposta.reqURI, "/api/rest", "", 1)
		if reqURI == "/bcpa-vendas/consultaPrevia/iniciarProposta" {
			sendDataService(lista[ctxProposta.session])
		}
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)

	if value != "" {
		panic(fmt.Sprintf("É necessário informar a variável de ambiente: %s", key))
	}

	return value
}

func sendDataService(p *proposta) {

	var err error
	p.request["nTpoPtoVda"], err = strconv.Atoi(fmt.Sprint(p.request["nTpoPtoVda"]))
	if err != nil {
		fmt.Printf("Erro ao converter em int o campo nTpoPtoVda: %s\n", err)
	}

	p.request["parceria"], err = strconv.Atoi(fmt.Sprint(p.request["parceria"]))
	if err != nil {
		fmt.Printf("Erro ao converter em int o campo parceria: %s\n", err)
	}

	p.request["cTpoPtoVda"], err = strconv.Atoi(fmt.Sprint(p.request["cTpoPtoVda"]))
	if err != nil {
		fmt.Printf("Erro ao converter em int o campo cTpoPtoVda: %s\n", err)
	}

	p.response["idConsultaPrevia"] = fmt.Sprint(big.NewInt(int64(p.response["idConsultaPrevia"].(float64))))
	p.response["nrPorpostaCartaoFisico"] = fmt.Sprint(big.NewInt(int64(p.response["nrPorpostaCartaoFisico"].(float64))))

	p.response["status"] = "Pré-Negado"
	if p.response["resultadoAnalise"] == 1 {
		p.response["status"] = "Pré-Aprovado"
	}

	namespace, err := uuid.Parse("edb2bb40-2108-4037-bd28-d5f6b8530e65")
	if err != nil {
		fmt.Printf("Erro ao converter string em uuid: %s\n", err)
	}
	id := uuid.NewSHA1(namespace, []byte(fmt.Sprint(p.response["idConsultaPrevia"])))

	jsonData := map[string]interface{}{
		"query": `
			mutation Update_bcpaProduto($updateBcpaProdutoId: ID!, $input: BCPAPropostaInput) {
				update_bcpaProduto(id: $updateBcpaProdutoId, input: $input) {
					id
					cpf
					numeroProposta
					nrPropostaCartaoFisico
					codigo
					campanha
				}
			}`,
		"variables": map[string]interface{}{
			"updateBcpaProdutoId": id,
			"input": map[string]interface{}{
				"cpf":                    fmt.Sprint(p.request["cpf"]) + fmt.Sprint(p.request["ctrlCpf"]),
				"ddd":                    p.request["ddd"],
				"celular":                p.request["celular"],
				"cdAcessoConsultaPrevia": p.request["cdAcessoConsultaPrevia"],
				"nrPontoVenda":           p.request["nTpoPtoVda"],
				"cdOrigemVenda":          p.request["parceria"],
				"cdTipoPontoVenda":       p.request["cTpoPtoVda"],
				"disp":                   p.request["disp"],
				"nrCtrlConsultaPrevia":   p.request["nrCtrlConsultaPrevia"],
				"campanha":               p.request["campanha"],
				"usuarioConsulta":        p.request["usuario"],
				"numeroProposta":         p.response["idConsultaPrevia"],
				"nrPropostaCartaoFisico": p.response["nrPorpostaCartaoFisico"],
				"numeroCtrlConsPrevio":   p.response["numeroCtrlConsPrevio"],
				"codigoRetorno":          p.response["codigoRetorno"],
				"uuid":                   p.response["uuid"],
				"limite_compra":          p.response["valorLimiteAprovadoCompra"],
				"limite_saque":           p.response["valorLimiteAprovadoSaque"],
				"limite_parcelado":       p.response["valorLimiteAprovadoParcela"],
				"cod_status":             p.response["resultadoAnalise"],
				"status":                 p.response["status"],
			},
		},
	}
	payload, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Printf("Erro ao ler o request: %s\n", err)
	}

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	oauthConfig := &clientcredentials.Config{
		ClientID:     getEnv("CLIENT_ID"),
		ClientSecret: getEnv("CLIENT_SECRET"),
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", getEnv("TENANT_ID")),
		Scopes: []string{
			fmt.Sprintf("api://%s/.default", getEnv("SCOPE_CLIENT_ID_BACK")),
		},
	}

	client := oauthConfig.Client(context.Background())
	req, err := http.NewRequest("POST", getEnv("GRAPHQL_URL"), bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("NewRequest failed with error %s\n", err)
	}
	req.Header.Add("x-functions-key", getEnv("GRAPHQL_CODE"))
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Erro ao ler o body %s\n", err)
	}

	fmt.Println(string(body))
}
