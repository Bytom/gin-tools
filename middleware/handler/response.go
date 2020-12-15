package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bytom/bytom/errors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Response describes the response standard. Code & Msg are always present.
// Data is present for a success response only.
type Response struct {
	Code       int             `json:"code"`
	Msg        string          `json:"msg"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationResp `json:"pagination,omitempty"`
}

// RespondErrorResp return error response
func (h *Handler) RespondErrorResp(c *gin.Context, err error) {
	log.WithFields(log.Fields{
		"url":     c.Request.URL,
		"request": c.Value(ReqBodyLabel),
		"err":     err,
	})
	c.AbortWithStatusJSON(http.StatusOK, h.formatErrResp(err))
}

// formatErrResp will find error code by specified error, then build the err response
func (h *Handler) formatErrResp(err error) Response {
	// default error response
	response := Response{
		Code: 300,
		Msg:  "request error",
	}

	root := errors.Root(err)
	if errCode, ok := h.errorCodes[root]; ok {
		response.Code = errCode
		response.Msg = root.Error()
	}
	return response
}

// RespondSuccessResp return success response
func (h *Handler) RespondSuccessResp(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, Response{Code: 200, Data: data})
}

// RespondSuccessPaginationResp return success response context of the pagination request
func (h *Handler) RespondSuccessPaginationResp(c *gin.Context, data interface{}, paginationProcessor *PaginationProcessor) {
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

// RespondErrorResp return error response
func (h *SimpleHandler) RespondErrorResp(c *gin.Context, err error) {
	log.WithFields(log.Fields{
		"url":     c.Request.URL,
		"request": c.Value(ReqBodyLabel),
		"err":     err,
	})
	c.AbortWithStatusJSON(http.StatusOK, h.formatErrResp(err).Msg)
}

// RespondSuccessResp return success response
func (h *SimpleHandler) RespondSuccessResp(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, data)
}
