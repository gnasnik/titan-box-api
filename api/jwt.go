package api

import (
	"database/sql"
	"errors"
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-box-api/core/dao"
	xerrors "github.com/gnasnik/titan-box-api/core/errors"
	"github.com/gnasnik/titan-box-api/core/generated/model"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

const (
	loginStatusFailure = iota
	loginStatusSuccess
)

type login struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Chaptcode string `json:"chaptcode"`
	Host      string `json:"host"`
}

type loginResponse struct {
	Token string `json:"token"`
}

var identityKey = "id"

func jwtGinMiddleware(secretKey string) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:             "User",
		Key:               []byte(secretKey),
		Timeout:           30 * 24 * time.Hour,
		MaxRefresh:        30 * 24 * time.Hour,
		IdentityKey:       identityKey,
		SendAuthorization: true,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &model.User{
				Username: claims[identityKey].(string),
			}
		},
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			c.JSON(http.StatusOK, loginResponse{
				Token: token,
			})
		},
		LogoutResponse: func(c *gin.Context, code int) {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
			})
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginParams login
			if err := c.BindJSON(&loginParams); err != nil {
				return "", fmt.Errorf("invalid input params")
			}

			if loginParams.Username == "" {
				return "", jwt.ErrMissingLoginValues
			}

			user, err := dao.GetUserByUsername(c.Request.Context(), loginParams.Username)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, xerrors.ErrUserNotFound
			}

			if err != nil {
				log.Errorf("get user by username: %v", err)
				return nil, xerrors.ErrInternalServer
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginParams.Password)); err != nil {
				return nil, xerrors.ErrInvalidPassword
			}

			return &model.User{Uid: user.Uid, Username: user.Username}, nil
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    401,
				"msg":     message,
				"success": false,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {
			c.Next()
		},
	})
}
