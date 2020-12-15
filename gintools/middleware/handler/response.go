package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bytom/bytom/errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ResponseAdaptor response interface
type ResponseAdaptor interface {
	RespondErrorResp(c *gin.Context, err error, errCode int)
	RespondSuccessResp(c *gin.Context, data interface{})
	RespondSuccessPaginationResp(c *gin.Context, data interface{}, paginationProcessor *PaginationProcessor)
}

// StandardResponse standard response
type StandardResponse struct {
}

// Response describes the response standard. Code & Msg are always present.
// Data is present for a success response only.
type Response struct {
	Code       int             `json:"code"`
	Msg        string          `json:"msg"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationResp `json:"pagination,omitempty"`
}

// RespondErrorResp return error response
func (h *StandardResponse) RespondErrorResp(c *gin.Context, err error, errCode int) {
	log.WithFields(log.Fields{
		"url":     c.Request.URL,
		"request": c.Value(ReqBodyLabel),
		"err":     err,
	})
	c.AbortWithStatusJSON(http.StatusOK, h.formatErrResp(err, errCode))
}

// formatErrResp will find error code by specified error, then build the err response
func (h *StandardResponse) formatErrResp(err error, code int) Response {
	// default error response
	response := Response{
		Code: 300,
		Msg:  "request error",
	}

	root := errors.Root(err)
	if code != 0 {
		response.Code = code
		response.Msg = root.Error()
	}
	return response
}

// RespondSuccessResp return success response
func (h *StandardResponse) RespondSuccessResp(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, Response{Code: 200, Data: data})
}

// RespondSuccessPaginationResp return success response context of the pagination request
func (h *StandardResponse) RespondSuccessPaginationResp(c *gin.Context, data interface{}, paginationProcessor *PaginationProcessor) {
	url := fmt.Sprintf("%v", c.Request.URL)
	base := strings.Split(url, "?")[0]
	links := paginationProcessor.getLinks(base)
	c.AbortWithStatusJSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Data: data,
		Pagination: &PaginationResp{
			Pagination: paginationProcessor.Pagination,
			Links:      links,
		},
	})
}

// SimpleResponse simple response
type SimpleResponse struct {
}

// RespondErrorResp return error response
func (h *SimpleResponse) RespondErrorResp(c *gin.Context, err error, errCode int) {
	log.WithFields(log.Fields{
		"url":     c.Request.URL,
		"request": c.Value(ReqBodyLabel),
		"err":     err,
	})
	c.AbortWithStatusJSON(http.StatusOK, errors.Root(err).Error())
}

// RespondSuccessResp return success response
func (h *SimpleResponse) RespondSuccessResp(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, data)
}

// RespondSuccessPaginationResp return success response context of the pagination request
func (h *SimpleResponse) RespondSuccessPaginationResp(c *gin.Context, data interface{}, paginationProcessor *PaginationProcessor) {
	c.AbortWithStatusJSON(http.StatusOK, data)
}
