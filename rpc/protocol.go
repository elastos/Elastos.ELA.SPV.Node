package rpc

const (
	// JSON-RPC protocol error codes.
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
	//-32000 to -32099	Server error, waiting for defining
)

type Method func(Params) (Result, error)

type Result interface{}

type MethodMap map[string]Method

type Request struct {
	Id      uint32      `json:"id,omitempty"`
	Version string      `json:"jsonrpc,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Response struct {
	Id      uint32 `json:"id,omitempty"`
	Version string `json:"jsonrpc,omitempty"`
	Result  Result `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Id      uint32 `json:"id,omitempty"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
