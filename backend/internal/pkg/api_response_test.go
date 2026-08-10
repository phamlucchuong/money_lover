package pkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelope mirrors the wire shape of APIResponse. Data is json.RawMessage
// so the original Go type survives JSON round-trip — otherwise assert.Equal
// would fail on []int vs []interface{} after re-decoding.
type envelope struct {
	Data      json.RawMessage `json:"data,omitempty"`
	Error     *apiErrorWire   `json:"error,omitempty"`
	Meta      *paginationWire  `json:"meta,omitempty"`
	TimeStamp string          `json:"timestamp"`
}

type apiErrorWire struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

type paginationWire struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()
	var got envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got),
		"response body must be valid JSON envelope: %s", rec.Body.String())
	return got
}

func TestJSONOK(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantCode int
		wantData string // JSON-encoded expected data; "" means the data field must be absent
	}{
		{
			name:     "struct payload",
			data:     map[string]string{"id": "abc"},
			wantCode: http.StatusOK,
			wantData: `{"id":"abc"}`,
		},
		{
			name:     "slice payload",
			data:     []int{1, 2, 3},
			wantCode: http.StatusOK,
			wantData: `[1,2,3]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder()
			c := newTestContext(rec)

			err := JSONOK(c, tt.data)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			got := decodeEnvelope(t, rec)
			assert.JSONEq(t, tt.wantData, string(got.Data))
			assert.Nil(t, got.Error, "success response must not include error field")
			assert.Nil(t, got.Meta, "JSONOK must not include meta")
			assert.NotEmpty(t, got.TimeStamp, "timestamp must always be populated")
		})
	}
}

func TestJSONOK_NilData(t *testing.T) {
	// APIResponse has `json:"data,omitempty"`, so a nil payload must be
	// omitted from the envelope entirely (not encoded as literal "null").
	rec := newRecorder()
	c := newTestContext(rec)

	err := JSONOK(c, nil)
	require.NoError(t, err)

	body := rec.Body.String()
	assert.NotContains(t, body, `"data"`, "nil data must be omitted, not encoded as null")
	assert.NotEmpty(t, decodeEnvelope(t, rec).TimeStamp)
}

func TestJSONList(t *testing.T) {
	rec := newRecorder()
	c := newTestContext(rec)

	meta := PaginationMeta{
		Page:       2,
		PageSize:   20,
		TotalItems: 137,
		TotalPages: 7,
	}
	data := []string{"a", "b"}

	err := JSONList(c, data, meta)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	got := decodeEnvelope(t, rec)
	assert.JSONEq(t, `["a","b"]`, string(got.Data))
	require.NotNil(t, got.Meta, "JSONList must include meta")
	assert.Equal(t, meta.Page, got.Meta.Page)
	assert.Equal(t, meta.PageSize, got.Meta.PageSize)
	assert.Equal(t, meta.TotalItems, got.Meta.TotalItems)
	assert.Equal(t, meta.TotalPages, got.Meta.TotalPages)
	assert.Nil(t, got.Error)
	assert.NotEmpty(t, got.TimeStamp)
}

func TestJSONCreated(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantCode int
		wantData string
	}{
		{name: "struct", data: map[string]int{"id": 42}, wantCode: http.StatusCreated, wantData: `{"id":42}`},
		{name: "string", data: "ok", wantCode: http.StatusCreated, wantData: `"ok"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder()
			c := newTestContext(rec)

			err := JSONCreated(c, tt.data)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)

			got := decodeEnvelope(t, rec)
			assert.JSONEq(t, tt.wantData, string(got.Data))
			assert.Nil(t, got.Error)
			assert.Nil(t, got.Meta)
			assert.NotEmpty(t, got.TimeStamp)
		})
	}
}

func TestJSONError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		code        int
		message     string
		wantStatus  int
		wantCode    int
		wantMessage string
	}{
		{
			name:        "bad request",
			status:      http.StatusBadRequest,
			code:        4001,
			message:     "invalid input",
			wantStatus:  http.StatusBadRequest,
			wantCode:    4001,
			wantMessage: "invalid input",
		},
		{
			name:        "internal server error",
			status:      http.StatusInternalServerError,
			code:        5000,
			message:     "boom",
			wantStatus:  http.StatusInternalServerError,
			wantCode:    5000,
			wantMessage: "boom",
		},
		{
			name:        "not found",
			status:      http.StatusNotFound,
			code:        4040,
			message:     "missing",
			wantStatus:  http.StatusNotFound,
			wantCode:    4040,
			wantMessage: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newRecorder()
			c := newTestContext(rec)

			err := JSONError(c, tt.status, tt.code, tt.message)
			require.NoError(t, err)

			assert.Equal(t, tt.wantStatus, rec.Code)

			got := decodeEnvelope(t, rec)
			require.NotNil(t, got.Error, "error response must include error field")
			assert.Equal(t, tt.wantCode, got.Error.Code)
			assert.Equal(t, tt.wantStatus, got.Error.Status)
			assert.Equal(t, tt.wantMessage, got.Error.Message)
			assert.Nil(t, got.Data, "error response must not include data")
			assert.Nil(t, got.Meta)
			assert.NotEmpty(t, got.TimeStamp)
		})
	}
}
