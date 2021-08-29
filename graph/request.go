package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	url    = "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-alt"
	method = "POST"
)


type graphErr struct {
	Message string
}

func (e graphErr) Error() string {
	return "graphql: " + e.Message
}

type graphResponse struct {
	Data   interface{} `json:"data"`
	Errors []graphErr
}

func RunQuery(ctx context.Context, graphQuery string, dataStructure interface{}) error {
	q := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     graphQuery,
		Variables: make(map[string]interface{}),
	}

	var requestBody bytes.Buffer
	if err := json.NewEncoder(&requestBody).Encode(q); err != nil {
		panic(err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, res.Body); err != nil {
		return fmt.Errorf("reading body: %v", err)
	}

	gr := graphResponse{
		Data: dataStructure,
	}
	if err := json.NewDecoder(&buf).Decode(&gr); err != nil {
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("graphql: server returned a non-200 status code: %v", res.StatusCode)
		}
		return fmt.Errorf("decoding response: %v", err)
	}

	if len(gr.Errors) > 0 {
		return gr.Errors[0]
	}

	return nil
}

