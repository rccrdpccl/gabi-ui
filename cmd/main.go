package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	wire "github.com/jeroenrinzema/psql-wire"
	"github.com/jeroenrinzema/psql-wire/codes"
	psqlerr "github.com/jeroenrinzema/psql-wire/errors"
	"github.com/lib/pq/oid"
)

const (
	ListenAddress = "0.0.0.0:5432"
)

func main() {
	url := os.Getenv("GABI_DOMAIN") + "/query"
	gabi := NewGabiClient(url)

	handler := NewGabiPgProxy(gabi)
	auth := wire.ClearTextPassword(func(username, password string) (bool, error) {
		gabi.SetToken(password)
		return true, nil
	})

	srv, err := wire.NewServer(wire.SimpleQuery(handler.Handle))
	if err != nil {
		panic(err)
	}
	srv.Auth = auth
	log.Println("PostgreSQL server is up and running at " + ListenAddress)
	srv.ListenAndServe(ListenAddress)
}

type Body struct {
	Query string `json:"query"`
}

type Response struct {
	Result Results     `json:"result"`
	Error  interface{} `json:"error"`
}

type Results []Result

type Result []string

type GabiClient struct {
	client *http.Client
	token  string
	url    string
}

func (g *GabiClient) SetToken(token string) {
	g.token = token
}

func (g *GabiClient) Query(query string) (*Response, error) {
	body := Body{
		Query: query,
	}
	payload, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}
	if g.token == "" {
		err := psqlerr.WithCode(errors.New("No token has been set"), codes.InvalidPassword)
		return nil, err
	}
	req, err := http.NewRequest("POST", g.url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		err := psqlerr.WithCode(errors.New("Gabi API returned status unauthorized code"), codes.InvalidPassword)
		return nil, err
	}
	response := &Response{}
	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(bodyBytes, response)
		if err != nil {
			return nil, err
		}
		if response.Result == nil {
			response.Result = Results{}
		}
		return response, nil
	}
	log.Println(fmt.Sprintf("StatusCode: %d, Query: %s", res.StatusCode, query))
	response.Result = Results{}
	return response, nil
}

func NewGabiClient(url string) *GabiClient {
	client := &http.Client{}
	return &GabiClient{
		client: client,
		url:    url,
	}
}

type GabiPgProxy struct {
	gabi *GabiClient
}

func NewGabiPgProxy(gabiClient *GabiClient) *GabiPgProxy {
	return &GabiPgProxy{
		gabi: gabiClient,
	}
}

func (h *GabiPgProxy) Handle(ctx context.Context, query string, writer wire.DataWriter, parameters []string) error {
	response, err := h.gabi.Query(query)
	if err != nil {
		log.Println("Gabi query returned an error", err)
		return err
	}
	for i, res := range response.Result {
		if i == 0 {
			table := createTable(res)
			writer.Define(table)
			continue
		}
		row := toRow(res)
		writer.Row(row)
	}

	return writer.Complete("OK")
}

func toRow(values []string) []any {
	row := make([]any, len(values))
	for i, v := range values {
		row[i] = v
	}
	return row
}

func createTable(headers []string) wire.Columns {
	columns := wire.Columns{}
	for _, header := range headers {
		columns = append(columns, wire.Column{
			Table:  0,
			Name:   header,
			Oid:    oid.T_text,
			Width:  256,
			Format: wire.TextFormat,
		})
	}
	return columns
}
