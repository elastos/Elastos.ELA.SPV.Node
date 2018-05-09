package rpc

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/elastos/Elastos.ELA.SPV.Node/config"

	"github.com/elastos/Elastos.ELA.SPV/log"
)

var methods MethodMap

func initMethods() {
	methods = make(MethodMap)
	methods["registeraddresses"] = RegisterAddresses
	methods["registeraddress"] = RegisterAddress
	methods["getblockcount"] = GetBlockCount
	methods["getbestblockhash"] = GetBestBlockHash
	methods["getblockhash"] = GetBlockHash
	methods["getblock"] = GetBlock
	methods["getblockbyheight"] = GetBlockByHeight
	methods["getrawtransaction"] = GetRawTransaction
	methods["sendrawtransaction"] = SendRawTransaction
}

func StartServer() {
	log.Debug("Start RPC server at port:", config.Values().RPCPort)
	initMethods()
	http.HandleFunc("/", Handle)
	err := http.ListenAndServe(":"+strconv.Itoa(config.Values().RPCPort), nil)
	if err != nil {
		log.Error("ListenAndServe: ", err.Error())
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	//JSON RPC commands should be POSTs
	if r.Method != "POST" {
		log.Warn("HTTP JSON RPC Handle - Method!=\"POST\"")
		http.Error(w, "JSON RPC procotol only allows POST method", http.StatusMethodNotAllowed)
		return
	}

	if r.Header["Content-Type"][0] != "application/json" {
		log.Warn("HTTP JSON RPC Handle - Content-Type: ", r.Header["Content-Type"][0], " not supported")
		http.Error(w, "need content type to be application/json", http.StatusUnsupportedMediaType)
		return
	}

	//read the body of the request
	body, _ := ioutil.ReadAll(r.Body)
	var request Request
	var response Response
	error := json.Unmarshal(body, &request)
	if error != nil {
		log.Warn("HTTP JSON RPC Handle - json.Unmarshal: ", error)
		response.WriteError(w, http.StatusBadRequest, ParseError, "rpc json parse error:"+error.Error())
		return
	}

	response.Id = request.Id
	response.Version = request.Version

	if len(request.Method) == 0 {
		response.WriteError(w, http.StatusBadRequest, InvalidRequest, "need a method!")
		return
	}
	method, ok := methods[request.Method]
	if !ok {
		response.WriteError(w, http.StatusNotFound, MethodNotFound, "method "+request.Method+" not found")
		return
	}

	// Json rpc 1.0 support positional parameters while json rpc 2.0 support named parameters.
	// positional parameters: { "params":[1, 2, 3....] }
	// named parameters: { "params":{ "a":1, "b":2, "c":3 } }
	// Here we support both of them, just like bitcion does.
	var params Params
	switch requestParams := request.Params.(type) {
	case nil:
	case []interface{}:
		params = formatParams(request.Method, requestParams)
	case map[string]interface{}:
		params = Params(requestParams)
	default:
		response.WriteError(w, http.StatusBadRequest, InvalidRequest, "params format error, must be an array or a map")
		return
	}

	log.Debug("RPC params:", params)

	result, err := method(params)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, InternalError, "internal error: "+err.Error())
		return
	}

	response.Result = result
	response.Write(w)
}

func (r *Response) WriteError(w http.ResponseWriter, httpStatus, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(httpStatus)
	r.Error = new(Error)
	r.Error.Code = code
	r.Error.Message = message
	data, _ := json.Marshal(r)
	w.Write(data)
}

func (r *Response) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Content-Type", "charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(r)
	w.Write(data)
}

func formatParams(method string, params []interface{}) Params {
	switch method {
	case "registeraddresses":
		return FromArray(params, "addresses")
	case "registeraddress":
		return FromArray(params, "address")
	case "getblockhash":
		return FromArray(params, "index")
	case "getblock":
		return FromArray(params, "hash", "format")
	case "getblockbyheight":
		return FromArray(params, "height", "format")
	case "getrawtransaction":
		return FromArray(params, "hash", "format")
	case "sendrawtransaction":
		return FromArray(params, "data", "format")
	default:
		return Params{}
	}
}
