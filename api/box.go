package api

import (
	"fmt"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-box-api/core/dao"
	"github.com/gnasnik/titan-box-api/core/generated/model"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"time"
)

func QueryBoxListGet(c *gin.Context) {
	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("pageSize"), 10, 64)

	ctx := c.Request.Context()
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	boxIds := c.QueryArray("boxIds")
	if boxIds == nil {
		boxIds = c.QueryArray("boxIds[]")
	}

	supplierBoxIds := c.QueryArray("supplierBoxIds")
	if supplierBoxIds == nil {
		supplierBoxIds = c.QueryArray("supplierBoxIds[]")
	}

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	total, boxes, err := dao.GetBoxesList(ctx, username, boxIds, supplierBoxIds, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box list: %v", err)
		return
	}

	out := GetBoxListResponse{
		Boxes: boxes,
		Total: strconv.Itoa(int(total)),
	}

	c.JSON(http.StatusOK, &out)
}

func QueryBoxListPost(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	type QueryBoxListRequest struct {
		Page           string   `json:"page"`
		PageSize       string   `json:"pageSize"`
		BoxIds         []string `json:"boxIds"`
		SupplierBoxIds []string `json:"supplierBoxIds"`
		Isp            []string `json:"isp"`
		Province       []string `json:"province"`
		ProcessStatus  []string `json:"processStatus"`
		Online         []string `json:"online"`
		Fuzzy          bool     `json:"fuzzy"`
		Remarks        []string `json:"remarks"`
	}

	var requestParam QueryBoxListRequest
	if err := c.BindJSON(&requestParam); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		log.Errorf("get box list: %v", err)
		return
	}

	page, _ := strconv.ParseInt(requestParam.Page, 10, 64)
	pageSize, _ := strconv.ParseInt(requestParam.PageSize, 10, 64)

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	ctx := c.Request.Context()

	total, boxes, err := dao.GetBoxesList(ctx, username, requestParam.BoxIds, requestParam.SupplierBoxIds, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box list: %v", err)
		return
	}

	out := GetBoxListResponse{
		Boxes: boxes,
		Total: strconv.Itoa(int(total)),
	}

	c.JSON(http.StatusOK, &out)
}

func QueryBoxIncomeV2Get(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	start := c.Query("start")
	end := c.Query("end")
	page, _ := strconv.ParseInt(c.Query("pageIndex"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.Query("pageSize"), 10, 64)
	boxIds := c.QueryArray("boxIds")
	if boxIds == nil {
		boxIds = c.QueryArray("boxIds[]")
	}

	if start == "" {
		start = time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	}

	if end == "" {
		end = time.Now().Format(time.DateOnly)
	}

	supplierBoxIds := c.QueryArray("supplierBoxIds")
	if supplierBoxIds == nil {
		supplierBoxIds = c.QueryArray("supplierBoxIds[]")
	}

	remarks := c.QueryArray("remarks")
	if remarks == nil {
		remarks = c.QueryArray("remarks[]")
	}

	ctx := c.Request.Context()
	totalNum, total, boxIncome, err := dao.GetBoxIncomeV2(ctx, username, boxIds, remarks, supplierBoxIds, start, end, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box list: %v", err)
		return
	}

	out := GetBoxIncomeResponse{
		BoxIncome: boxIncome,
		Total:     strconv.Itoa(int(total)),
		TotalNum:  strconv.Itoa(int(totalNum)),
	}

	c.JSON(http.StatusOK, &out)
}

func QueryBoxIncomeV2Post(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	type QueryBoxIncomeRequest struct {
		Start          string   `json:"start"`
		End            string   `json:"end"`
		BoxIds         []string `json:"boxIds"`
		Remarks        []string `json:"remarks"`
		SupplierBoxIds []string `json:"supplierBoxIds"`
		Page           string   `json:"pageIndex"`
		PageSize       string   `json:"pageSize"`
	}

	var requestParam QueryBoxIncomeRequest
	if err := c.BindJSON(&requestParam); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		log.Errorf("get box list: %v", err)
		return
	}

	ctx := c.Request.Context()
	page, _ := strconv.ParseInt(requestParam.Page, 10, 64)
	pageSize, _ := strconv.ParseInt(requestParam.PageSize, 10, 64)

	if requestParam.Start == "" {
		requestParam.Start = time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
	}

	if requestParam.End == "" {
		requestParam.End = time.Now().Format(time.DateOnly)
	}

	if page == 0 {
		page = 1
	}

	if pageSize == 0 {
		pageSize = 10
	}

	totalNum, total, boxIncome, err := dao.GetBoxIncomeV2(ctx, username, requestParam.BoxIds, requestParam.Remarks, requestParam.SupplierBoxIds, requestParam.Start, requestParam.End, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box list: %v", err)
		return
	}

	out := GetBoxIncomeResponse{
		BoxIncome: boxIncome,
		Total:     strconv.Itoa(int(total)),
		TotalNum:  strconv.Itoa(int(totalNum)),
	}

	c.JSON(http.StatusOK, &out)
}

func QueryBoxBandwidthGet(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	date := c.Query("date")
	startHour := c.Query("startHour")
	endHour := c.Query("endHour")
	boxIds := c.QueryArray("boxIds")
	if boxIds == nil {
		boxIds = c.QueryArray("boxIds[]")
	}

	supplierBoxIds := c.QueryArray("supplierBoxIds")
	if supplierBoxIds == nil {
		supplierBoxIds = c.QueryArray("supplierBoxIds[]")
	}

	if date == "" || boxIds == nil {
		c.JSON(http.StatusBadRequest, errors.New("BadRequest"))
		return
	}

	var startStr, endStr string

	if startHour != "" && endHour != "" {
		startStr = fmt.Sprintf("%s %s:00:00", date, startHour)
		endStr = fmt.Sprintf("%s %s:59:59", date, endHour)
	} else {
		startStr = fmt.Sprintf("%s 00:00:00", date)
		endStr = fmt.Sprintf("%s 23:59:59", date)
	}

	start, _ := time.Parse(time.DateTime, startStr)
	end, _ := time.Parse(time.DateTime, endStr)

	ctx := c.Request.Context()
	bandwidths, err := dao.GetBoxBandwidth(ctx, username, boxIds, supplierBoxIds, start.Unix(), end.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box bandwidth: %v", err)
		return
	}

	tempBoxIds := make(map[string]*BoxBandwidths)
	for _, b := range bandwidths {
		if _, existing := tempBoxIds[b.BoxId]; !existing {
			tempBoxIds[b.BoxId] = &BoxBandwidths{
				BoxId:         b.BoxId,
				SupplierBoxId: b.SupplierBoxId,
				Bandwidths:    make([]*model.BoxBandwidth, 0),
			}
		}

		tempBoxIds[b.BoxId].Bandwidths = append(tempBoxIds[b.BoxId].Bandwidths, b)
	}

	var boxBandwidths []*BoxBandwidths
	for _, b := range tempBoxIds {
		boxBandwidths = append(boxBandwidths, b)
	}

	c.JSON(http.StatusOK, GetBoxBandwidthResponse{
		BoxBandwidths: boxBandwidths,
	})
}

func QueryBoxBandwidthPost(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	type QueryBoxBandwidthRequest struct {
		BoxId          []string `json:"boxId"`
		SupplierBoxIds []string `json:"supplierBoxIds"`
		Date           string   `json:"date"`
		StartHour      string   `json:"startHour"`
		EndHour        string   `json:"endHour"`
	}

	var requestParam QueryBoxBandwidthRequest
	if err := c.BindJSON(&requestParam); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		log.Errorf("get box list: %v", err)
		return
	}

	remarks := c.QueryArray("remarks")
	if remarks == nil {
		remarks = c.QueryArray("remarks[]")
	}

	if requestParam.Date == "" || requestParam.BoxId == nil {
		c.JSON(http.StatusBadRequest, errors.New("BadRequest"))
		return
	}

	var startStr, endStr string

	if requestParam.StartHour != "" && requestParam.EndHour != "" {
		startStr = fmt.Sprintf("%s %s:00:00", requestParam.Date, requestParam.StartHour)
		endStr = fmt.Sprintf("%s %s:59:59", requestParam.Date, requestParam.EndHour)
	} else {
		startStr = fmt.Sprintf("%s 00:00:00", requestParam.Date)
		endStr = fmt.Sprintf("%s 23:59:59", requestParam.Date)
	}

	start, _ := time.Parse(time.DateTime, startStr)
	end, _ := time.Parse(time.DateTime, endStr)

	ctx := c.Request.Context()
	bandwidths, err := dao.GetBoxBandwidth(ctx, username, requestParam.BoxId, requestParam.SupplierBoxIds, start.Unix(), end.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box bandwidth: %v", err)
		return
	}

	tempBoxIds := make(map[string]*BoxBandwidths)
	for _, b := range bandwidths {
		if _, existing := tempBoxIds[b.BoxId]; !existing {
			tempBoxIds[b.BoxId] = &BoxBandwidths{
				BoxId:         b.BoxId,
				SupplierBoxId: b.SupplierBoxId,
				Bandwidths:    make([]*model.BoxBandwidth, 0),
			}
		}

		tempBoxIds[b.BoxId].Bandwidths = append(tempBoxIds[b.BoxId].Bandwidths, b)
	}

	var boxBandwidths []*BoxBandwidths
	for _, b := range tempBoxIds {
		boxBandwidths = append(boxBandwidths, b)
	}

	c.JSON(http.StatusOK, GetBoxBandwidthResponse{
		BoxBandwidths: boxBandwidths,
	})
}

func QueryBoxQualityGet(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	date := c.Query("date")
	startHour := c.Query("startHour")
	endHour := c.Query("endHour")
	boxIds := c.QueryArray("boxIds")
	if boxIds == nil {
		boxIds = c.QueryArray("boxIds[]")
	}

	supplierBoxIds := c.QueryArray("supplierBoxIds")
	if supplierBoxIds == nil {
		supplierBoxIds = c.QueryArray("supplierBoxIds[]")
	}

	if date == "" || boxIds == nil {
		c.JSON(http.StatusBadRequest, errors.New("BadRequest"))
		return
	}

	var startStr, endStr string

	if startHour != "" && endHour != "" {
		startStr = fmt.Sprintf("%s %s:00:00", date, startHour)
		endStr = fmt.Sprintf("%s %s:59:59", date, endHour)
	} else {
		startStr = fmt.Sprintf("%s 00:00:00", date)
		endStr = fmt.Sprintf("%s 23:59:59", date)
	}

	start, _ := time.Parse(time.DateTime, startStr)
	end, _ := time.Parse(time.DateTime, endStr)

	ctx := c.Request.Context()
	qualities, err := dao.GetBoxQualities(ctx, username, boxIds, supplierBoxIds, start.Unix(), end.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box qualities: %v", err)
		return
	}

	tempBoxIds := make(map[string]*BoxQualities)
	for _, b := range qualities {
		if _, existing := tempBoxIds[b.BoxId]; !existing {
			tempBoxIds[b.BoxId] = &BoxQualities{
				BoxId:         b.BoxId,
				SupplierBoxId: b.SupplierBoxId,
				Qualities:     make([]*model.BoxQuality, 0),
			}
		}

		tempBoxIds[b.BoxId].Qualities = append(tempBoxIds[b.BoxId].Qualities, b)
	}

	var boxQualities []*BoxQualities
	for _, b := range tempBoxIds {
		boxQualities = append(boxQualities, b)
	}

	c.JSON(http.StatusOK, GetBoxQualitiesResponse{
		BoxQualities: boxQualities,
	})
}

func QueryBoxQualityPost(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	type QueryBoxBandwidthRequest struct {
		BoxId          []string `json:"boxId"`
		SupplierBoxIds []string `json:"supplierBoxIds"`
		Date           string   `json:"date"`
		StartHour      string   `json:"startHour"`
		EndHour        string   `json:"endHour"`
	}

	var requestParam QueryBoxBandwidthRequest
	if err := c.BindJSON(&requestParam); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		log.Errorf("get box list: %v", err)
		return
	}

	if requestParam.Date == "" || requestParam.BoxId == nil {
		c.JSON(http.StatusBadRequest, errors.New("BadRequest"))
		return
	}

	var startStr, endStr string

	if requestParam.StartHour != "" && requestParam.EndHour != "" {
		startStr = fmt.Sprintf("%s %s:00:00", requestParam.Date, requestParam.StartHour)
		endStr = fmt.Sprintf("%s %s:59:59", requestParam.Date, requestParam.EndHour)
	} else {
		startStr = fmt.Sprintf("%s 00:00:00", requestParam.Date)
		endStr = fmt.Sprintf("%s 23:59:59", requestParam.Date)
	}

	start, _ := time.Parse(time.DateTime, startStr)
	end, _ := time.Parse(time.DateTime, endStr)

	ctx := c.Request.Context()
	qualities, err := dao.GetBoxQualities(ctx, username, requestParam.BoxId, requestParam.SupplierBoxIds, start.Unix(), end.Unix())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.New("InternalServerError"))
		log.Errorf("get box qualities: %v", err)
		return
	}

	tempBoxIds := make(map[string]*BoxQualities)
	for _, b := range qualities {
		if _, existing := tempBoxIds[b.BoxId]; !existing {
			tempBoxIds[b.BoxId] = &BoxQualities{
				BoxId:         b.BoxId,
				SupplierBoxId: b.SupplierBoxId,
				Qualities:     make([]*model.BoxQuality, 0),
			}
		}

		tempBoxIds[b.BoxId].Qualities = append(tempBoxIds[b.BoxId].Qualities, b)
	}

	var boxQualities []*BoxQualities
	for _, b := range tempBoxIds {
		boxQualities = append(boxQualities, b)
	}

	c.JSON(http.StatusOK, GetBoxQualitiesResponse{
		BoxQualities: boxQualities,
	})
}
