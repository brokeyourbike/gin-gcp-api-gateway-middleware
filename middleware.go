package gcpapigatewaymw

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// gatewayUserInfoHeader is the header that contains the user info.
const gatewayUserInfoHeader = "X-Apigateway-Api-Userinfo"

// gatewayUserInfoCtx is a context key for the GatewayUserInfo.
const gatewayUserInfoCtx = "gatewayUserInfo"

type GatewayUserInfo struct {
	Sub   string `json:"sub" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// Middleware creates a new GatewayCtx middleware.
func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		encodedUser := ctx.GetHeader(gatewayUserInfoHeader)
		if encodedUser == "" {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		decodedBytes, err := base64.RawURLEncoding.DecodeString(encodedUser)
		if err != nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		var userInfo GatewayUserInfo

		err = json.NewDecoder(bytes.NewReader(decodedBytes)).Decode(&userInfo)
		if err != nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		err = binding.Validator.ValidateStruct(&userInfo)
		if err != nil {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		ctx.Set(gatewayUserInfoCtx, userInfo)
		ctx.Next()
	}
}

// GetGatewayUserInfo returns the user info from the context.
func GetGatewayUserInfo(ctx *gin.Context) GatewayUserInfo {
	userInfo := ctx.MustGet(gatewayUserInfoCtx).(GatewayUserInfo)
	return userInfo
}

// GetGatewayUserID returns the user ID from the context.
func GetGatewayUserID(ctx *gin.Context) string {
	userInfo := GetGatewayUserInfo(ctx)
	return userInfo.Sub
}
