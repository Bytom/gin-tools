package handler

import (
	"encoding/json"
	"reflect"

	"github.com/bytom/bytom/errors"
	"github.com/gin-gonic/gin"
)

// FrontFilter represent a filter function that is called before processing request
type FrontFilter func(ctx *gin.Context) error

// RequestFilter represent a filter function that is filter request
type RequestFilter func(ctx *gin.Context, req interface{}) error

// Handler is a framework for processing each API request, which contains parsing request parameters, error handling and so on
type Handler struct {
	frontFilters   []FrontFilter
	requestFilters []RequestFilter
	errorCodes     map[error]int
	respAdaptor    ResponseAdaptor
}

type handlerFun interface{}

// NewHandler return a handler instance
func NewHandler(errorCodes map[error]int, frontFilters []FrontFilter, requestFilters []RequestFilter) *Handler {
	return &Handler{
		frontFilters:   frontFilters,
		requestFilters: requestFilters,
		errorCodes:     errorCodes,
		respAdaptor:    &StandardResponse{},
	}
}

func (h *Handler) SetResponseAdaptor(respAdaptor ResponseAdaptor) *Handler {
	h.respAdaptor = respAdaptor
	return h
}

func callHandleFunc(fun handlerFun, args ...interface{}) []interface{} {
	fv := reflect.ValueOf(fun)

	params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}

	rs := fv.Call(params)
	result := make([]interface{}, len(rs))
	for i, r := range rs {
		result[i] = r.Interface()
	}
	return result
}

func createHandleReqArg(fun handlerFun, context *gin.Context) (interface{}, error) {
	ft := reflect.TypeOf(fun)
	if ft.NumIn() == 1 {
		return nil, nil
	}

	if ft.In(1) == paginationQueryType {
		return nil, nil
	}

	argType := ft.In(1).Elem()

	reqArg := reflect.New(argType).Interface()
	if err := context.ShouldBindJSON(reqArg); err != nil {
		return nil, errors.Wrap(err, "bind reqArg")
	}

	b, err := json.Marshal(reqArg)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal")
	}

	context.Set(ReqBodyLabel, string(b))

	return reqArg, nil
}

// HandleMiddleware wrap a handler function, and return a gin-compatible processing functions
func (h *Handler) HandleMiddleware(handleFunc interface{}) func(*gin.Context) {
	if err := ValidateFuncType(handleFunc); err != nil {
		panic(err)
	}

	return func(context *gin.Context) {
		for _, filter := range h.frontFilters {
			if err := filter(context); err != nil {
				h.respAdaptor.RespondErrorResp(context, err, h.handlerErrCode(err))
				return
			}
		}
		h.handleRequest(context, handleFunc)
	}
}

func (h *Handler) handleRequest(context *gin.Context, fun handlerFun) {
	args, err := h.buildHandleFuncArgs(fun, context)
	if err != nil {
		h.respAdaptor.RespondErrorResp(context, err, h.handlerErrCode(err))
		return
	}

	result := callHandleFunc(fun, args...)
	if err := result[len(result)-1]; err != nil {
		h.respAdaptor.RespondErrorResp(context, err.(error), h.handlerErrCode(err.(error)))
		return
	}

	if exist := h.processPaginationIfPresent(args, result, context); exist {
		return
	}

	if len(result) == 1 {
		h.respAdaptor.RespondSuccessResp(context, struct{}{})
		return
	}

	h.respAdaptor.RespondSuccessResp(context, result[0])
}

func (h *Handler) processPaginationIfPresent(args []interface{}, result []interface{}, context *gin.Context) bool {
	// default the last param is pagination query param
	query, ok := args[len(args)-1].(*PaginationQuery)
	if !ok {
		return false
	}

	paginationResult := result[0].(*PaginationResult)
	list := paginationResult.data
	paginationProcessor := NewPaginationProcessor(query, paginationResult.total)
	h.respAdaptor.RespondSuccessPaginationResp(context, list, paginationProcessor)
	return true
}

func (h *Handler) buildHandleFuncArgs(fun handlerFun, context *gin.Context) ([]interface{}, error) {
	args := []interface{}{context}

	req, err := createHandleReqArg(fun, context)
	if err != nil {
		return nil, errors.Wrap(err, "createHandleReqArg")
	}

	for _, filter := range h.requestFilters {
		if err := filter(context, req); err != nil {
			return nil, err
		}
	}

	if req != nil {
		args = append(args, req)
	}

	ft := reflect.TypeOf(fun)

	// not exist pagination
	if ft.In(ft.NumIn()-1) != paginationQueryType {
		return args, nil
	}

	query, err := ParsePagination(context)
	if err != nil {
		return nil, errors.Wrap(err, "ParsePagination")
	}

	args = append(args, query)
	return args, nil
}

var (
	errorType            = reflect.TypeOf((*error)(nil)).Elem()
	contextType          = reflect.TypeOf((*gin.Context)(nil))
	paginationQueryType  = reflect.TypeOf((*PaginationQuery)(nil))
	paginationResultType = reflect.TypeOf((*PaginationResult)(nil))
)

// ValidateFuncType used to validate the handler function's argumetns and return value
func ValidateFuncType(fun handlerFun) error {
	ft := reflect.TypeOf(fun)
	if ft.Kind() != reflect.Func || ft.IsVariadic() {
		return errors.New("need nonvariadic func in " + ft.String())
	}

	if ft.NumIn() < 1 || ft.NumIn() > 3 {
		return errors.New("need one or two or three parameters in " + ft.String())
	}

	if ft.In(0) != contextType {
		return errors.New("the first parameter must point of context in " + ft.String())
	}

	if ft.NumIn() == 2 && ft.In(1).Kind() != reflect.Ptr {
		return errors.New("the second parameter must point in " + ft.String())
	}

	if ft.NumOut() < 1 || ft.NumOut() > 2 {
		return errors.New("the size of return value must one or two in " + ft.String())
	}

	hasPagination := ft.In(ft.NumIn()-1) == paginationQueryType
	// if has pagination, the first return value must slice or array
	if hasPagination && ft.Out(0) != paginationResultType {
		return errors.New("the first return value of pagination must paginationResultType in " + ft.String())
	}

	if !ft.Out(ft.NumOut() - 1).Implements(errorType) {
		return errors.New("the last return value must error in " + ft.String())
	}
	return nil
}

func (h *Handler) handlerErrCode(err error) int {
	root := errors.Root(err)
	if errCode, ok := h.errorCodes[root]; ok {
		return errCode
	}

	return 0
}
