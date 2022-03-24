package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
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

	url := "http://localhost:7071/api/graphql/"
	method := "POST"

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

	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	req.Header.Add("x-functions-key", "7LBPOKyfOpFZtPFawzkDPr1bzaQWnFax8C4D/ZzfMKyzXwMs7F1AlQ==")
	req.Header.Add("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6ImpTMVhvMU9XRGpfNTJ2YndHTmd2UU8yVnpNYyIsImtpZCI6ImpTMVhvMU9XRGpfNTJ2YndHTmd2UU8yVnpNYyJ9.eyJhdWQiOiJhcGk6Ly9jZmJiYmRiMS1kYzY1LTQxY2QtOWI2Ny0zMDQ5YzhiNTYwNmUiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC83N2Y5MWY5Ny1lZWFjLTRlYzAtYmZmNS1hMzM4MjJiMGEwMjkvIiwiaWF0IjoxNjQ4MTE0NzY1LCJuYmYiOjE2NDgxMTQ3NjUsImV4cCI6MTY0ODExODY2NSwiYWlvIjoiRTJaZ1lGaHk4dXZTVW0ySjBQV2hDMk5mZkU3UkFRQT0iLCJhcHBpZCI6IjMxN2FiZTVmLTdmOGMtNDhkZi04YmRmLTBmN2M4M2FkNGMxZiIsImFwcGlkYWNyIjoiMSIsImlkcCI6Imh0dHBzOi8vc3RzLndpbmRvd3MubmV0Lzc3ZjkxZjk3LWVlYWMtNGVjMC1iZmY1LWEzMzgyMmIwYTAyOS8iLCJvaWQiOiIxNmNlOWQ1My0zM2JmLTQ0MmMtOGNkNy1hMTI4NWU0YjVhM2UiLCJyaCI6IjAuQVhVQWx4XzVkNnp1d0U2XzlhTTRJckNnS2JHOXU4OWwzTTFCbTJjd1NjaTFZRzUxQUFBLiIsInJvbGVzIjpbIk1TRl9MZWl0dXJhIiwiTURNX0xlaXR1cmEiXSwic3ViIjoiMTZjZTlkNTMtMzNiZi00NDJjLThjZDctYTEyODVlNGI1YTNlIiwidGlkIjoiNzdmOTFmOTctZWVhYy00ZWMwLWJmZjUtYTMzODIyYjBhMDI5IiwidXRpIjoiVWZjWFBZZUZ6RS1DVktOWWlINXJBUSIsInZlciI6IjEuMCJ9.imcRzKjboK0don_Dqbz8mHelQsl4u4FYwUHQuH1uJ_hgyOXzWmVlKHpLyYNWJeiqfaoXY6u3IGpKpJWCkScpELPeH3oryS58cnwdet7fvB9-v-WJc1Mi5X3ei4s8d2vyuUH28Pkp7RtQz6Jkwa0McNSxQsRj4ZHjE08L2ri_ikmszmX8jxm_I9w-k6MgPbhHs9m4iQam_oQ6a--xWUUogq35w6l7AROs8XRVjgwvDoTDHZPE8Ua-Hf1Gu7at_CQGroJxI_nKgWp3YI0Q5jwSVMd_aNz_ucPdgPWdi2PGBJygg1kSE3RN3udLw-BDPlA564e92gkONytsWk2p4_QfdQ")
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
