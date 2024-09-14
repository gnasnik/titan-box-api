package api

import (
	"database/sql"
	"errors"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-box-api/core/dao"
	xerrors "github.com/gnasnik/titan-box-api/core/errors"
	"github.com/gnasnik/titan-box-api/core/generated/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type registerParams struct {
	Username    string `json:"username"`
	PhoneNumber string `json:"phoneNumber"`
	VerifyCode  string `json:"verifyCode"`
	Password    string `json:"password"`
	Src         string `json:"src"`
	Public      bool   `json:"public"`
	Code        string `json:"code"`
}

func UserRegister(c *gin.Context) {
	var params registerParams
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, respError(xerrors.ErrInvalidParams))
		return
	}

	userInfo := &model.User{
		Username:    params.Username,
		PhoneNumber: params.PhoneNumber,
	}

	passwd := params.Password
	if userInfo.Username == "" {
		c.JSON(http.StatusBadRequest, respError(xerrors.ErrInvalidParams))
		return
	}

	_, err := dao.GetUserByUsername(c.Request.Context(), userInfo.Username)
	if err == nil {
		c.JSON(http.StatusInternalServerError, respError(xerrors.ErrUserExist))
		return
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("GetUserByUsername: %v", err)
		c.JSON(http.StatusInternalServerError, respError(xerrors.ErrInvalidParams))
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, respError(xerrors.ErrInternalServer))
		return
	}
	userInfo.Password = string(passHash)
	userInfo.AppKey = strings.ReplaceAll(uuid.NewString(), "-", "")
	userInfo.AppSecret = strings.ReplaceAll(uuid.NewString(), "-", "")

	err = dao.CreateUser(c.Request.Context(), userInfo)
	if err != nil {
		log.Errorf("create user : %v", err)
		c.JSON(http.StatusInternalServerError, respError(xerrors.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, JsonObject{
		"msg": "success",
	})
}

func QueryUserInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	user, err := dao.GetUserByUsername(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusBadRequest, respError(xerrors.ErrUserNotFound))
		return
	}
	c.JSON(http.StatusOK, user)
}
