package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bytom/bytom/errors"
	"github.com/gin-gonic/gin"
)

const (
	defaultStartStr = "0"
	defaultLimitStr = "10"
	maxPageLimit    = 1000
)

var (
	errParsePaginationStart = fmt.Errorf("parse pagination start")
	errParsePaginationLimit = fmt.Errorf("parse pagination limit")
)

// PaginationResult used to return the pagination info
type PaginationResult struct {
	data  interface{}
	total uint64
}

// NewPaginationResult is a factory method of PaginationResult
func NewPaginationResult(data interface{}, total uint64) *PaginationResult {
	return &PaginationResult{data: data, total: total}
}

// PaginationQuery represent an argument of pagination query
type Pagination struct {
	Start uint64 `json:"start"`
	Limit uint64 `json:"limit"`
	Total uint64 `json:"total,omitempty"`
}

// PaginationQuery is the conditions for paging query
type PaginationQuery Pagination

// PaginationResp is the response struct to pagination datas
type PaginationResp struct {
	*Pagination
	Links links `json:"_links"`
}

type links struct {
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// ParsePagination request meets the standard on https://developer.atlassian.com/server/confluence/pagination-in-the-rest-api/
func ParsePagination(c *gin.Context) (*PaginationQuery, error) {
	startStr := c.DefaultQuery("start", defaultStartStr)
	limitStr := c.DefaultQuery("limit", defaultLimitStr)

	start, err := strconv.ParseUint(startStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errParsePaginationStart)
	}

	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errParsePaginationLimit)
	}

	if limit > maxPageLimit {
		limit = maxPageLimit
	}

	return &PaginationQuery{
		Start: start,
		Limit: limit,
	}, nil
}

// PaginationProcessor is the middle result of paging query
type PaginationProcessor struct {
	*Pagination
	HasNext bool
	HasPrev bool
}

// NewPaginationProcessor create a new PaginationProcessor
func NewPaginationProcessor(query *PaginationQuery, size int, total uint64) *PaginationProcessor {
	return &PaginationProcessor{
		Pagination: &Pagination{
			Start: query.Start,
			Limit: query.Limit,
			Total: total,
		},
		HasNext: size == int(query.Limit),
		HasPrev: 0 != int(query.Start),
	}
}

// getLinks return the calculated PaginationProcessor links
func (p *PaginationProcessor) getLinks(baseURL string) links {
	l := links{}
	if p.HasNext {
		// To efficiently build a string using Write methods
		// https://stackoverflow.com/questions/1760757/how-to-efficiently-concatenate-strings-in-go
		// https://tip.golang.org/pkg/strings/#Builder
		var b strings.Builder
		fmt.Fprintf(&b, "%s?limit=%d&start=%d", baseURL, p.Limit, p.Start+p.Limit)
		l.Next = b.String()
	}

	if p.HasPrev {
		var b strings.Builder
		prevStart := p.Start - p.Limit
		if prevStart < 0 {
			prevStart = 0
		}
		fmt.Fprintf(&b, "%s?limit=%d&start=%d", baseURL, p.Limit, prevStart)
		l.Prev = b.String()
	}

	return l
}
