package stub

import (
	"time"
)

type Request struct {
	Query string `json:"query"`
}

type Response struct {
	Type     string          `json:"type"`
	Model    Model           `json:"model"`
	Metadata Metadata        `json:"metadata"`
	Dataset  string          `json:"dataset"`
	Data     [][]interface{} `json:"data"`
}

type Model struct {
	Name     string  `json:"name"`
	Fields   []Field `json:"fields"`
	JSONPath string  `json:"$jsonPath"`
	Model    string  `json:"$model"`
}

type Hints struct {
	Kind  string `json:"kind"`
	Field string `json:"field"`
	Type  string `json:"type"`
}

type Field struct {
	Alias string `json:"alias"`
	Type  string `json:"type"`
	Hints Hints  `json:"hints"`
	Form  string `json:"form,omitempty"`
	Model Model  `json:"model,omitempty"`
}

type Metadata struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}
