package gingcpapigatewaymw

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// gatewayUserInfoHeader is the header that contains the user info.
const gatewayUserInfoHeader = "X-Apigateway-Api-Userinfo"

// gatewayUserInfoCtx is a context key for the GatewayUserInfo.
const gatewayUserInfoCtx = "gatewayUserInfo"

type GatewayUserInfo struct {
	Sub   uuid.UUID `json:"sub" binding:"required"`
	Email string    `json:"email" binding:"required,email"`
}

type GatewayCtx struct {
	log *logrus.Logger
}

// NewGatewayCtx creates a new GatewayCtx middleware.
func NewGatewayCtx(log *logrus.Logger) *GatewayCtx {
	return &GatewayCtx{log: log}
}

// Require is a middleware handler that extracts user info from the request.
func (m GatewayCtx) Require() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		encodedUser := ctx.GetHeader(gatewayUserInfoHeader)
		if encodedUser == "" {
			m.log.WithContext(ctx).Warn("UserInfo not available")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		decodedBytes, err := base64.RawURLEncoding.DecodeString(encodedUser)
		if err != nil {
			m.log.WithContext(ctx).WithError(err).Warn("UserInfo is not base64 encoded")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		var userInfo GatewayUserInfo

		err = json.NewDecoder(bytes.NewReader(decodedBytes)).Decode(&userInfo)
		if err != nil {
			m.log.WithContext(ctx).WithError(err).Warn("UserInfo can not be decoded")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		err = binding.Validator.ValidateStruct(&userInfo)
		if err != nil {
			m.log.WithContext(ctx).WithError(err).Warn("UserInfo is not valid")
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
func GetGatewayUserID(ctx *gin.Context) uuid.UUID {
	userInfo := GetGatewayUserInfo(ctx)
	return userInfo.Sub
}
