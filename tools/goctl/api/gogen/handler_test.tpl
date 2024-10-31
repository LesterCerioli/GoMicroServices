package {{.PkgName}}

import (
	"bytes"
	{{if .HasRequest}}"encoding/json"{{end}}
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	{{.ImportPackages}}
)

{{if .HasDoc}}{{.Doc}}{{end}}
func Test{{.HandlerName}}(t *testing.T) {
	// new service context
	c := config.Config{}
	svcCtx := svc.NewServiceContext(c)

	tests := []struct {
		name       string
		reqBody    interface{}
		wantStatus int
		wantResp   string
		setupMocks func()
	}{
		{
			name:    "invalid request body",
			reqBody: "invalid",
			wantStatus: http.StatusBadRequest,
			wantResp:   `{"code":400,"msg":"invalid request"}`, // Adjust based on actual error response
			setupMocks: func() {
				// No setup needed for this test case
			},
		},
		{
			name: "handler error",
			{{if .HasRequest}}reqBody: types.{{.RequestType}}{
				//TODO: add fields here
			},
			{{end}}wantStatus: http.StatusUnauthorized,
			wantResp:   `{"code":401,"msg":"unauthorized"}`, // Adjust based on actual error response
			setupMocks: func() {
				// Mock login logic to return an error
			},
		},
		{
			name: "handler successful",
			{{if .HasRequest}}reqBody: types.{{.RequestType}}{
				//TODO: add fields here
			},
			{{end}}wantStatus: http.StatusOK,
			wantResp:   `{"code":0,"msg":"success","data":{}}`, // Adjust based on actual success response
			setupMocks: func() {
				// Mock login logic to return success
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			reqBody := []byte{}
			{{if .HasRequest}}reqBody, err := json.Marshal(tt.reqBody)
			require.NoError(t, err){{end}}
			req, err := http.NewRequest("POST", "/ut", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := {{.HandlerName}}(svcCtx)
			handler.ServeHTTP(rr, req)
			t.Log(rr.Body.String())
			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.JSONEq(t, tt.wantResp, rr.Body.String())
		})
	}
}