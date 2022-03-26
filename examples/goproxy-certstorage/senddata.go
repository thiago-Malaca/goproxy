package main

import (
	"encoding/json"
	"fmt"
	"math/big"
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

func SendData(ctxProposta *CtxProposta, b []byte) {
	if len(b) == 0 {
		return
	}

	if lista[ctxProposta.session] == nil {
		lista[ctxProposta.session] = &proposta{}
	}

	var objmap map[string]interface{}
	if err := json.Unmarshal(b, &objmap); err != nil {
		Error.Printf("Não conseguiu converter para json! Path: %s: %v\n", ctxProposta.reqURI, err)
		return
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
					dataEnvio
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
				"dataEnvio":              time.Now().Format(time.RFC3339),
			},
		},
	}

	body, err := requestGraphql(jsonData)
	if err != nil {
		fmt.Printf("Erro ao ler o body %s\n", err)
	}

	fmt.Println(string(body))
}
