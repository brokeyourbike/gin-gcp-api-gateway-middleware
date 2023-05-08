package gcpapigatewaymw_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gcpapigatewaymw "github.com/brokeyourbike/gin-gcp-api-gateway-middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

func TestMiddleware(t *testing.T) {
	noRequiredFields := base64.RawURLEncoding.EncodeToString([]byte(`{"a": "b"}`))
	valid := base64.RawURLEncoding.EncodeToString([]byte(`{"sub": "c2133353-6547-4429-a453-4c8fa2fdbacd", "email": "john@doe.com"}`))

	tests := []struct {
		name       string
		headers    map[string]string
		wantStatus int
	}{
		{
			"missing header",
			map[string]string{},
			http.StatusForbidden,
		},
		{
			"header is not base64 encoded",
			map[string]string{"X-Apigateway-Api-Userinfo": "not-base64-encoded"},
			http.StatusForbidden,
		},
		{
			"encoded value is not json",
			map[string]string{"X-Apigateway-Api-Userinfo": "SSBhbSBub3QgSlNPTg=="}, // I am not JSON
			http.StatusForbidden,
		},
		{
			"json does not contain required fields",
			map[string]string{"X-Apigateway-Api-Userinfo": noRequiredFields},
			http.StatusForbidden,
		},
		{
			"valid json",
			map[string]string{"X-Apigateway-Api-Userinfo": valid},
			http.StatusOK,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range test.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router := gin.New()
			router.Use(gcpapigatewaymw.Middleware())
			router.GET("/", func(ctx *gin.Context) {
				info := gcpapigatewaymw.GetGatewayUserInfo(ctx)
				assert.Equal(t, "c2133353-6547-4429-a453-4c8fa2fdbacd", info.Sub.String())

				id := gcpapigatewaymw.GetGatewayUserID(ctx)
				assert.Equal(t, "c2133353-6547-4429-a453-4c8fa2fdbacd", id.String())

				ctx.String(http.StatusOK, "the end.")
			})
			router.ServeHTTP(w, req)

			assert.Equal(t, test.wantStatus, w.Code)
		})
	}
}
