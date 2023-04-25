package stub

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

type Options struct {
	Port     string
	BindAddr string
	StubPath string
}

const DEFAULT_PORT = "4771"

func RunStubServer(opt Options) {
	if opt.Port == "" {
		opt.Port = DEFAULT_PORT
	}
	addr := opt.BindAddr + ":" + opt.Port
	r := chi.NewRouter()
	r.Post("/add", addStub)
	r.Post("/uql/add", updateUqlResponse)
	r.Get("/", listStub)
	r.Post("/find", handleFindStub)
	r.Get("/uql/", listUqlStub)
	r.Post("/uql/monitoring/v1dev/query/execute", handleUqlStub)
	r.Get("/clear", handleClearStub)

	if opt.StubPath != "" {
		readStubFromFile(opt.StubPath)
	}
	initUqlResponse()
	fmt.Println("Serving stub admin on http://" + addr)
	go func() {
		err := http.ListenAndServe(addr, r)
		log.Fatal(err)
	}()
}

func responseError(err error, w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

type Stub struct {
	Service string `json:"service"`
	Method  string `json:"method"`
	Input   Input  `json:"input"`
	Output  Output `json:"output"`
}

type Input struct {
	Equals   map[string]interface{} `json:"equals"`
	Contains map[string]interface{} `json:"contains"`
	Matches  map[string]interface{} `json:"matches"`
}

type Output struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
}

func addStub(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(err, w)
		return
	}

	stub := new(Stub)
	err = json.Unmarshal(body, stub)
	if err != nil {
		responseError(err, w)
		return
	}

	err = validateStub(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	err = storeStub(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	w.Write([]byte("Success add stub"))
}

func updateUqlResponse(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseError(err, w)
		return
	}
	fmt.Printf("Body : %s", body)
	fmt.Println()
	var req uqlRequest
	if err = json.Unmarshal(body, &req); err != nil {
		responseError(err, w)
	}
	fmt.Printf("Request : %s", req.Query)
	uqlStorage[req.Query] = req.Response
	w.Write([]byte("Updated uql response"))
}

func initUqlResponse() {
	req := uqlRequest{
		Query: "FETCH attributes('agent.id') FROM entities",
		Response: []Response{
			{
				Type:     "model",
				Model:    Model{},
				Metadata: Metadata{},
				Dataset:  "d:main",
				Data:     [][]interface{}{},
			},
			{
				Type:     "data",
				Model:    Model{},
				Metadata: Metadata{},
				Dataset:  "d:main",
				Data: [][]interface{}{
					{
						"01GWM6D42R6CT30S1P64T3EB9H",
					},
					{
						"01GWMH72CRC9GKECB560V32B9P",
					},
				},
			},
		},
	}

	fmt.Printf("Request : %s", req.Query)
	uqlStorage[req.Query] = req.Response

	req = uqlRequest{
		Query: "FETCH attributes FROM entities",
		Response: []Response{
			{
				Type:     "model",
				Model:    Model{},
				Metadata: Metadata{},
				Dataset:  "d:main",
				Data:     [][]interface{}{},
			},
			{
				Type:     "data",
				Model:    Model{},
				Metadata: Metadata{},
				Dataset:  "d:attributes-1",
				Data: [][]interface{}{
					{
						"k8s.namespace.name",
						"cosmos_unit",
					},
					{
						"platform",
						"k8s_2",
					},
					{
						"agent.id",
						"01GWM6D42R6CT30S1P64T3EB9H",
					},
					{
						"agent.version",
						"2.0",
					},
				},
			},
			{
				Type:     "data",
				Model:    Model{},
				Metadata: Metadata{},
				Dataset:  "d:attributes-2",
				Data: [][]interface{}{
					{
						"k8s.namespace.name",
						"cosmos_3",
					},
					{
						"platform",
						"k8s_3",
					},
					{
						"agent.id",
						"01GWMH72CRC9GKECB560V32B9P",
					},
					{
						"agent.version",
						"2.0",
					},
				},
			},
		},
	}
	fmt.Printf("Request : %s", req.Query)
	uqlStorage[req.Query] = req.Response
}

func listStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allStub())
}

func listUqlStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allUqlStub())
}

func validateStub(stub *Stub) error {
	if stub.Service == "" {
		return fmt.Errorf("Service name can't be empty")
	}

	if stub.Method == "" {
		return fmt.Errorf("Method name can't be emtpy")
	}

	// due to golang implementation
	// method name must capital
	stub.Method = strings.Title(stub.Method)

	switch {
	case stub.Input.Contains != nil:
		break
	case stub.Input.Equals != nil:
		break
	case stub.Input.Matches != nil:
		break
	default:
		return fmt.Errorf("Input cannot be empty")
	}

	// TODO: validate all input case

	if stub.Output.Error == "" && stub.Output.Data == nil {
		return fmt.Errorf("Output can't be empty")
	}
	return nil
}

type findStubPayload struct {
	Service string                 `json:"service"`
	Method  string                 `json:"method"`
	Data    map[string]interface{} `json:"data"`
}

func handleFindStub(w http.ResponseWriter, r *http.Request) {
	stub := new(findStubPayload)
	err := json.NewDecoder(r.Body).Decode(stub)
	if err != nil {
		responseError(err, w)
		return
	}

	// due to golang implementation
	// method name must capital
	stub.Method = strings.Title(stub.Method)

	output, err := findStub(stub)
	if err != nil {
		log.Println(err)
		responseError(err, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func handleUqlStub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := io.ReadAll(r.Body)
	for uqlKey, uqlResponse := range uqlStorage {
		if strings.Contains(string(body), uqlKey) {
			json.NewEncoder(w).Encode(uqlResponse)
		}
	}
}

func handleClearStub(w http.ResponseWriter, r *http.Request) {
	clearStorage()
	w.Write([]byte("OK"))
}
