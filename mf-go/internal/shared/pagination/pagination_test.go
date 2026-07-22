package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromRequest_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	params := FromRequest(r)

	assert.Equal(t, DefaultPage, params.Page)
	assert.Equal(t, DefaultPerPage, params.PerPage)
}

func TestFromRequest_CustomValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?page=3&per_page=50", nil)
	params := FromRequest(r)

	assert.Equal(t, 3, params.Page)
	assert.Equal(t, 50, params.PerPage)
}

func TestFromRequest_MaxPerPage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?per_page=500", nil)
	params := FromRequest(r)

	assert.Equal(t, MaxPerPage, params.PerPage)
}

func TestFromRequest_MaxPage(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/test?page=999999999", nil)
	params := FromRequest(r)

	assert.Equal(t, MaxPage, params.Page)
}

func TestParams_Offset(t *testing.T) {
	params := Params{Page: 3, PerPage: 20}
	assert.Equal(t, 40, params.Offset())
}

func TestNewResult(t *testing.T) {
	data := []string{"a", "b", "c"}
	params := Params{Page: 1, PerPage: 10}
	result := NewResult(data, params, 25)

	assert.Len(t, result.Data, 3)
	assert.Equal(t, 25, result.TotalCount)
	assert.Equal(t, 3, result.TotalPages)
}
