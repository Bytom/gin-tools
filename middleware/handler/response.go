package handler

import (
	"fmt"
	"net/http"
	"strings"

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
func RespondErrorResp(c *gin.Context, err error, formatErrResp FormatErrResp) {
	log.WithFields(log.Fields{
		"url":     c.Request.URL,
		"request": c.Value(ReqBodyLabel),
		"err":     err,
	})
	c.AbortWithStatusJSON(http.StatusOK, formatErrResp(err))
}

// RespondSuccessResp return success response
func RespondSuccessResp(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(http.StatusOK, Response{Code: 200, Data: data})
}

// RespondSuccessPaginationResp return success response context of the pagination request
func RespondSuccessPaginationResp(c *gin.Context, data interface{}, paginationProcessor *PaginationProcessor) {
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
