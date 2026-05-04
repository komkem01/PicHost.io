package base

import (
	"errors"
	"strings"

	"github.com/uptrace/bun"
)

type Pagination struct {
	Page    int    `form:"page"`
	Size    int    `form:"size"`
	OrderBy string `form:"order_by"`
	SortBy  string `form:"sort_by"`
}

var DEFAULT_PAGINATION = Pagination{
	Page:    1,
	Size:    10,
	OrderBy: `asc`,
	SortBy:  `created_at`,
}

const (
	defaultMinLimit = 10
	defaultMaxLimit = 100
)

// Paginate use by `/list` request
type RequestPaginate struct {
	// Required. Must be >= 1
	Size int `json:"size" form:"size"`

	// Required. Must be >= 0
	Page int `json:"page" form:"page"`

	// (optional, required search_by)
	// Search string. Ignore if len(search) < 3 and search_by is not given.
	// By default, it will search exact keyword. To use ambiguous search, set fuzzy to true.
	Search string `json:"search" form:"search" binding:"omitempty"`

	// (optional, required search)
	// Column name to search. Ignore if search is not specify.
	// Must be one of the given column
	SearchBy string `json:"search_by" form:"search_by" binding:"omitempty"`

	// (optional, required search_by & search)
	// If fuzzy is set to true, it will use "% LIKE %" command to search.
	// If fuzzy is set to false, it will use "? = ?" command to find the exact keyword
	// Default is false
	Fuzzy bool `json:"fuzzy" form:"fuzzy" binding:"omitempty"`

	// (optional, required order_by)
	// Column name to sort. Ignore if search is not specify. Must be one of the given column
	SortBy string `json:"sort_by" form:"sort_by" binding:"omitempty"`

	// (optional, required sort_by)
	// Must be one of these: ["ASC", "asc", "DESC", "desc"]
	OrderBy string `json:"order_by" form:"order_by" binding:"omitempty"`

	StartDate int64 `json:"start_date" form:"start_date" binding:"omitempty"`

	EndDate int64 `json:"end_date" form:"end_date" binding:"omitempty"`

	Date int64 `json:"date" form:"date" binding:"omitempty"`
}

var (
	ErrInvalidSort         = errors.New("paginate: invalid sort_by cols")
	ErrInvalidOrderBy      = errors.New("paginate: invalid order_by attr")
	ErrInvalidSearchLength = errors.New("paginate: invalid search length < 3")
	ErrInvalidSearchCol    = errors.New("paginate: invalid search_by col")
)

func IsPagErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "paginate: ")
}

func (p *RequestPaginate) GetPage() int64 {
	if p.Page < 1 {
		return 1
	}
	return int64(p.Page)
}

func (p *RequestPaginate) GetSize() int64 {
	if p.Size < 10 {
		return defaultMinLimit
	} else if p.Size > 100 {
		return defaultMaxLimit
	}
	return int64(p.Size)
}

func (p *RequestPaginate) SetOffsetLimit(selQ *bun.SelectQuery) {
	offset := (p.GetPage() - 1) * p.GetSize()
	selQ.Offset(int(offset)).Limit(int(p.GetSize()))
}

func (p *RequestPaginate) SetSearchBy(selQ *bun.SelectQuery, acceptCol []string) error {
	if p.Search == `` || p.SearchBy == `` {
		return nil
	}
	search := p.Search
	searchBy := p.SearchBy

	search = removeMalicious(search)

	if len(search) < 3 {
		return ErrInvalidSearchLength
	}

	if !containsStringList(acceptCol, searchBy) {
		// log.Info("search_by: [%s]", searchBy)
		return ErrInvalidSearchCol
	}

	if p.Fuzzy {
		// log.Info("use fuzzy search:[%s] by [%s]", search, searchBy)
		selQ.Where("?::text LIKE ?", bun.Ident(searchBy), "%"+search+"%")
		return nil
	}
	selQ.Where("? = ?", bun.Ident(searchBy), search)
	return nil
}

func (p *RequestPaginate) SetSortOrder(selQ *bun.SelectQuery, acceptCol []string) error {
	if p.SortBy == `` {
		return nil
	}

	orderBy := strings.ToUpper(p.OrderBy)
	sortBy := p.SortBy

	if orderBy != "ASC" && orderBy != "DESC" {
		orderBy = `ASC`
	}

	if !containsStringList(acceptCol, sortBy) {
		return ErrInvalidSort
	}

	selQ.OrderExpr("? "+orderBy, bun.Ident(sortBy))
	return nil
}

var (
	chQuestionMark = []rune("?")[0]
	chPercent      = []rune("%")[0]
)

func removeMalicious(s string) string {
	r := []rune(s)
	var sb strings.Builder
	for i := 0; i < len(r); i++ {
		if r[i] != chQuestionMark && r[i] != chPercent {
			sb.WriteRune(r[i])
		}
	}
	return strings.TrimSpace(sb.String())
}

func containsStringList(list []string, key string) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == key {
			return true
		}
	}
	return false
}
