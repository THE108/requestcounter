package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/THE108/requestcounter/common"
	"github.com/THE108/requestcounter/errors"
	"github.com/THE108/requestcounter/log"
	"github.com/THE108/requestcounter/tracedata"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

const (
	// GET - For detected HTTP request GET method
	GET = "GET"

	// POST - For detected HTTP request POST method
	POST = "POST"

	// HEAD
	HEAD = "HEAD"

	// PUT
	PUT = "PUT"

	// PATCH
	PATCH = "PATCH"

	// DELETE
	DELETE = "DELETE"
)

const (
	debugUrlParamName = "debug"
)

// Info described routes
type HandlerInfo struct {
	Name    string
	Method  string
	Route   string
	Handler interface{}
}

// IGetHandler defines handlers that process GET-like requests
// (return some data based on incoming parameters)
type IGetHandler interface {
	Process(ctx context.Context, params common.Params) (interface{}, error)
}

// IPostHandler defines handlers that process POST-like requests
// (process incoming data using additional parameters)
type IPostHandler interface {
	Process(ctx context.Context, data interface{}, params common.Params) (interface{}, error)
	GetBuffer() interface{}
}

type IHttpCodeGetter interface {
	GetHttpCode() int
}

func (this *Application) initRoutes() {
	this.router = mux.NewRouter()

	handlersInfo := this.getHandlers()
	this.handlers = make(map[string]*HandlerInfo, len(handlersInfo))
	for _, info := range handlersInfo {
		key := info.Method + info.Name
		if _, ok := this.handlers[key]; ok {
			panic("cannot add handler for route because it already exists: " + info.Name)
		}

		this.handlers[key] = info

		this.addHandler(info)
	}
}

func (this *Application) addHandler(info *HandlerInfo) {
	var httpHandler http.Handler
	switch h := info.Handler.(type) {
	case IGetHandler:
		httpHandler = this.createGetRequestHandler(info.Name, h)
	case IPostHandler:
		httpHandler = this.createPostRequestHandler(info.Name, h)
	default:
		panic("unknown type")
	}

	this.router.Handle(info.Route, httpHandler).Methods(info.Method)
}

func (this *Application) createGetRequestHandler(handlerName string, h IGetHandler) http.Handler {
	httpHandler := func(rw http.ResponseWriter, req *http.Request) {
		ctx := this.createContext(handlerName, req)
		params := common.NewParams(req)

		output, err := h.Process(ctx, params)
		if err != nil {
			this.writeResponse(ctx, rw, err)
			return
		}

		if req.Method == HEAD {
			// since HEAD request SHOULD NOT return a message-body in the response,
			// so set data empty before request finish
			this.writeResponse(ctx, rw, nil)
			return
		}

		this.writeResponse(ctx, rw, output)
	}

	return http.HandlerFunc(httpHandler)
}

func (this *Application) createPostRequestHandler(handlerName string, h IPostHandler) http.Handler {
	httpHandler := func(rw http.ResponseWriter, req *http.Request) {
		ctx := this.createContext(handlerName, req)
		params := common.NewParams(req)

		var input interface{}
		if input = h.GetBuffer(); input != nil {
			if err := this.getInputFromRequest(req, input); err != nil {
				this.writeResponse(ctx, rw, err)
				return
			}
		}

		output, err := h.Process(ctx, input, params)
		if err != nil {
			this.writeResponse(ctx, rw, err)
			return
		}

		this.writeResponse(ctx, rw, output)
	}

	return http.HandlerFunc(httpHandler)
}

func (this *Application) getInputFromRequest(req *http.Request, input interface{}) error {
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrap(err, http.StatusInternalServerError)
	}

	if err := json.Unmarshal(requestBody, input); err != nil {
		return errors.Wrap(err, http.StatusInternalServerError)
	}

	return nil
}

func (this *Application) makeTraceDataLoggerPrefix(data *tracedata.TraceData) string {
	return data.TraceID + "|" + data.SpanID + "|" + data.ParentSpanID
}

func (this *Application) getLogLevelByHandlerName(handlerName string, req *http.Request) int {
	_, debug := req.URL.Query()[debugUrlParamName]
	if debug {
		return log.DEBUG
	}

	return log.ERROR
}

func (this *Application) createContext(handlerName string, req *http.Request) context.Context {
	traceData := tracedata.GetTraceDataFromRequest(req)
	ctx := tracedata.SetTraceDataToContext(context.Background(), traceData)
	prefix := this.makeTraceDataLoggerPrefix(traceData)
	level := this.getLogLevelByHandlerName(handlerName, req)
	logger := log.New(os.Stderr, prefix, level)

	logger.Debug(req.Method + " " + req.RequestURI)

	return log.SetLoggerToContext(ctx, logger)
}

// SetResult prepares API response for success scenario
func (this *Application) writeResponse(ctx context.Context, rw http.ResponseWriter, data interface{}) {
	httpCode := http.StatusOK
	if r, ok := data.(IHttpCodeGetter); ok {
		httpCode = r.GetHttpCode()
	}

	if data == nil {
		rw.WriteHeader(httpCode)
		return
	}

	var err error
	response, ok := data.([]byte)
	if !ok {
		// If response is not a sequence of bytes, we need to serialize it
		response, err = json.Marshal(data)
	}

	if err != nil {
		log.GetLoggerFromContext(ctx).Error(err.Error())
		httpCode = http.StatusInternalServerError
	}

	rw.WriteHeader(httpCode)
	if _, err := rw.Write(response); err != nil {
		log.GetLoggerFromContext(ctx).Error(err.Error())
	}
}
