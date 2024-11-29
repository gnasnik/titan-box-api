package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gnasnik/titan-box-api/core/dao"
	"github.com/gnasnik/titan-box-api/core/generated/model"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// API endpoint for fetching boxes
	paiNetBaseUrl = "https://openapi.painet.work"

	defaultPageSize = 200
)

type DataService struct {
	Interval time.Duration

	lk sync.Mutex
	//Fetchers map[string]DataFetcher
}

func NewDataService() *DataService {
	return &DataService{
		Interval: 10 * time.Minute,
		//Fetchers: make(map[string]DataFetcher),
	}
}

type Response struct {
	Code    int64  `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func (d *DataService) Run(ctx context.Context) {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(time.Local),
	)

	c.AddFunc("0 0 10 * * *", d.startSyncTimer)

	c.Start()

	d.startSyncTicker()

	for {
		select {
		case <-ticker.C:
			d.startSyncTicker()
		case <-ctx.Done():
			break
		}
	}
}

func (d *DataService) startSyncTicker() {
	paiNetInfo, err := dao.GetUserKeys(context.Background())
	if err != nil {
		log.Errorf("get user keys: %v", err)
		return
	}

	ctx := context.Background()

	for _, pi := range paiNetInfo {
		if pi.Status == 1 {
			continue
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func(pi *model.PaiNetInfo) {
			defer wg.Done()

			fmt.Println("start-> sync box list", pi.PaiUsername)
			if err := d.syncBoxList(ctx, pi); err != nil {
				log.Errorf("sync box list error: %v", err)
			}
		}(pi)

		wg.Add(1)
		go func(pi *model.PaiNetInfo) {
			defer wg.Done()

			start := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
			end := time.Now().Format(time.DateOnly)

			if err := d.syncBoxIncome(pi, start, end); err != nil {
				log.Errorf("sync box income error: %v", err)
			}

		}(pi)

		wg.Wait()
	}

}

func (d *DataService) startSyncTimer() {
	paiNetInfo, err := dao.GetUserKeys(context.Background())
	if err != nil {
		log.Errorf("get user keys: %v", err)
		return
	}

	for _, pi := range paiNetInfo {
		if pi.Status == 1 {
			continue
		}

		var wg sync.WaitGroup

		wg.Add(1)
		go func(pi *model.PaiNetInfo) {
			defer wg.Done()

			yesterday := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
			if err := d.syncBoxBandwidth(pi, yesterday); err != nil {
				log.Errorf("sync box bandwidth error: %v", err)
			}

		}(pi)

		wg.Add(1)
		go func(pi *model.PaiNetInfo) {
			defer wg.Done()

			yesterday := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
			if err := d.syncBoxQualities(pi, yesterday); err != nil {
				log.Errorf("sync box qualities error: %v", err)
			}

		}(pi)

		wg.Wait()
	}

}

type GetBoxListResponse struct {
	Boxes []*model.Box `json:"boxes"`
	Total string       `json:"total"`
}

func getBox(ctx context.Context, pi *model.PaiNetInfo, page, size int) (*GetBoxListResponse, error) {
	url := fmt.Sprintf("%s/boxsupplier/v1/box/list?page=%d&pageSize=%d", paiNetBaseUrl, page, size)
	response, err := doRequest(pi, url)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to fetch boxes size")
	}

	var res GetBoxListResponse
	if err := json.Unmarshal(response, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (d *DataService) syncBoxList(ctx context.Context, pi *model.PaiNetInfo) error {

	var (
		count int64 = 0
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes income")

		response, err := getBox(ctx, pi, page, defaultPageSize)
		if err != nil {
			//return errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
			//errChan <- errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
			log.Errorf("Failed to fetch boxes from page %d", page)
			return errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
		}

		total, err := strconv.ParseInt(response.Total, 10, 64)
		if err != nil {
			return err
		}

		if total == 0 {
			return nil
		}

		var diskInfos []*model.DiskInfo

		for _, box := range response.Boxes {
			box.Username = pi.PaiUsername

			for _, diskInfo := range box.DiskInfos {
				diskInfo.BoxId = box.BoxId
				diskInfo.SupplierBoxId = box.SupplierBoxId
				diskInfos = append(diskInfos, diskInfo)
			}
			diskInfos = append(diskInfos)
		}

		if err := SaveBoxList(ctx, response.Boxes, diskInfos); err != nil {
			//errChan <- errors.Wrapf(err, "Failed to save boxes from page %d", page)
			log.Errorf("Failed to save boxes from page %d", page)
			return errors.Wrapf(err, "Failed to save boxes from page %d", page)
		}

		count += int64(len(response.Boxes))
		if count >= total {
			break
		}
	}

	log.Info("Synchronization of boxes list completed successfully.")

	//response, err := getBox(ctx, pi, 1, 1)
	//if err != nil {
	//	return errors.Wrapf(err, "Failed to fetch boxes from page max size")
	//}
	//
	//total, err := strconv.ParseInt(response.Total, 10, 64)
	//if err != nil {
	//	return err
	//}
	//
	//if total == 0 {
	//	fmt.Println("get box list, total 0")
	//	return nil
	//}

	//wg := sync.WaitGroup{}
	//boxesChan := make(chan *GetBoxListResponse)
	//doneChan := make(chan interface{}, 1)
	//errChan := make(chan error)

	//maxPage := int(total) / defaultPageSize

	//for page := 1; page <= 1; page++ {
	//	//wg.Add(1)
	//
	//	log.Infof("Starting to query boxes on page: %d", page)
	//
	//	response, err = getBox(ctx, pi, page, defaultPageSize)
	//	if err != nil {
	//		//return errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
	//		//errChan <- errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
	//		log.Errorf("Failed to fetch boxes from page %d", page)
	//		return errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
	//	}
	//
	//	var diskInfos []*model.DiskInfo
	//
	//	for _, box := range response.Boxes {
	//		box.Username = pi.PaiUsername
	//
	//		for _, diskInfo := range box.DiskInfos {
	//			diskInfo.BoxId = box.BoxId
	//			diskInfo.SupplierBoxId = box.SupplierBoxId
	//			diskInfos = append(diskInfos, diskInfo)
	//		}
	//		diskInfos = append(diskInfos)
	//	}
	//
	//	if err := SaveBoxList(ctx, response.Boxes, diskInfos); err != nil {
	//		//errChan <- errors.Wrapf(err, "Failed to save boxes from page %d", page)
	//		log.Errorf("Failed to save boxes from page %d", page)
	//		return errors.Wrapf(err, "Failed to save boxes from page %d", page)
	//	}
	//
	//}

	//wg.Wait()

	//go func() {
	//	wg.Wait()
	//
	//	doneChan <- nil
	//}()
	//
	//for {
	//	select {
	//	case <-ctx.Done():
	//		return ctx.Err()
	//	case err := <-errChan:
	//		return err
	//	case boxesRes := <-boxesChan:
	//		fmt.Println("+++=>", "get", len(boxesRes.Boxes))
	//		var diskInfos []*model.DiskInfo
	//
	//		for _, box := range boxesRes.Boxes {
	//			box.Username = pi.PaiUsername
	//
	//			for _, diskInfo := range box.DiskInfos {
	//				diskInfo.BoxId = box.BoxId
	//				diskInfo.SupplierBoxId = box.SupplierBoxId
	//				diskInfos = append(diskInfos, diskInfo)
	//			}
	//			diskInfos = append(diskInfos)
	//		}
	//
	//		if err := SaveBoxList(ctx, boxesRes.Boxes, diskInfos); err != nil {
	//			return err
	//		}
	//	case <-doneChan:
	//		log.Info("Synchronization of boxes list completed successfully.")
	//		return nil
	//	}
	//}
	return nil
}

func SaveBoxList(ctx context.Context, boxes []*model.Box, diskInfos []*model.DiskInfo) error {
	err := dao.BulkUpsertBoxes(ctx, boxes)
	if err != nil {
		return err
	}

	if len(diskInfos) == 0 {
		return nil
	}

	err = dao.BulkUpsertBoxDiskInfo(ctx, diskInfos)
	if err != nil {
		return err
	}

	return nil
}

type GetBoxIncomeResponse struct {
	BoxIncome []*model.BoxIncome `json:"list"`
	Total     string             `json:"total"`
	TotalNum  string             `json:"totalNum"`
}

func (d *DataService) syncBoxIncome(pi *model.PaiNetInfo, start, end string) error {
	var (
		count int64 = 0
		ctx         = context.Background()
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes income")

		url := fmt.Sprintf("%s/boxsupplier/v1/supplier/income_v2?start=%s&end=%s&pageSize=%d&pageIndex=%d", paiNetBaseUrl, start, end, defaultPageSize, page)

		fmt.Println("url=>", url)
		response, err := doRequest(pi, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes income")
		}

		fmt.Println(string(response))

		var res GetBoxIncomeResponse
		if err := json.Unmarshal(response, &res); err != nil {
			return err
		}

		total, err := strconv.ParseInt(res.TotalNum, 10, 64)
		if err != nil {
			return err
		}

		if total == 0 {
			return nil
		}

		if err := d.saveBoxIncome(ctx, pi.PaiUsername, res.BoxIncome); err != nil {
			return err
		}

		count += int64(len(res.BoxIncome))
		if count >= total {
			break
		}
	}

	log.Info("Synchronization of boxes income completed successfully.")

	return nil
}

func (d *DataService) saveBoxIncome(ctx context.Context, username string, boxIncome []*model.BoxIncome) error {
	for _, b := range boxIncome {
		b.Username = username
	}

	err := dao.BulkUpsertBoxDayIncome(ctx, boxIncome)
	if err != nil {
		return err
	}
	return nil
}

func generateMD5Hash(as, account string, timestamp int64) string {
	text := fmt.Sprintf("%s#%d#%s", as, timestamp, account)
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func doRequest(pi *model.PaiNetInfo, url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	sign := generateMD5Hash(pi.APISecret, pi.PaiUsername, timestamp)

	request.Header.Set("ak", pi.APIKey)
	request.Header.Set("timestamp", fmt.Sprintf("%d", timestamp))
	request.Header.Set("sign", sign)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (d *DataService) StartSyncBoxList(pi *model.PaiNetInfo) {
	if err := d.syncBoxList(context.Background(), pi); err != nil {
		log.Infof("syncBoxList: %v", err)
	}

}

func (d *DataService) StartSyncBoxIncomeHistoryFrom(pi *model.PaiNetInfo, date string) {
	var startTime, endTime time.Time

	if date == "" {
		year, month, day := time.Now().Date()
		startTime = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		endTime = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	} else {
		dateTime, err := time.Parse(time.DateOnly, date)
		if err != nil {
			log.Errorf("parse date: %v", err)
			return
		}

		startTime = dateTime
		endTime = time.Now()
	}

	for startTime.Before(endTime) {
		start := startTime.Format(time.DateOnly)
		end := startTime.AddDate(0, 0, 10).Format(time.DateOnly)
		startTime = startTime.AddDate(0, 0, 10)

		fmt.Println(" ==>", start)
		if err := d.syncBoxIncome(pi, start, end); err != nil {
			log.Infof("syncBoxIncome: start: %s, end: %s, %v", start, end, err)
		}
	}

}

func (d *DataService) StartSyncBoxDayBandwidthHistoryFrom(pi *model.PaiNetInfo, date string) {
	var startTime, endTime time.Time

	if date == "" {
		year, month, day := time.Now().Date()
		startTime = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		endTime = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	} else {
		dateTime, err := time.Parse(time.DateOnly, date)
		if err != nil {
			log.Errorf("parse date: %v", err)
			return
		}

		startTime = dateTime
		endTime = time.Now()
	}

	for startTime.Before(endTime) {
		start := startTime.Format(time.DateOnly)
		end := startTime.AddDate(0, 0, 1).Format(time.DateOnly)
		startTime = startTime.AddDate(0, 0, 1)

		if err := d.syncBoxBandwidth(pi, start); err != nil {
			log.Infof("syncBoxBandwidth: start: %s, end: %s, %v", start, end, err)
		}
	}

}

func (d *DataService) StartSyncBoxDayQualitiesHistoryFrom(pi *model.PaiNetInfo, date string) {
	var startTime, endTime time.Time

	if date == "" {
		year, month, day := time.Now().Date()
		startTime = time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		endTime = time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	} else {
		dateTime, err := time.Parse(time.DateOnly, date)
		if err != nil {
			log.Errorf("parse date: %v", err)
			return
		}

		startTime = dateTime
		endTime = time.Now()
	}

	for startTime.Before(endTime) {
		start := startTime.Format(time.DateOnly)
		end := startTime.AddDate(0, 0, 1).Format(time.DateOnly)
		startTime = startTime.AddDate(0, 0, 1)

		if err := d.syncBoxQualities(pi, start); err != nil {
			fmt.Println("==>", err)
			log.Infof("syncBoxQualities: start: %s, end: %s, %v", start, end, err)
		}
	}

}

type GetBoxBandwidthResponse struct {
	BoxBandwidths []*BoxBandwidths `json:"boxBandwidths"`
}

type BoxBandwidths struct {
	BoxId         string                `json:"boxId"`
	SupplierBoxId string                `json:"supplierBoxId"`
	Bandwidths    []*model.BoxBandwidth `json:"bandwidths"`
}

func (d *DataService) saveBoxBandwidth(ctx context.Context, username string, boxBandwidths []*BoxBandwidths) error {
	bandwidthMap := make(map[string][]*model.BoxBandwidth)

	for _, box := range boxBandwidths {
		for _, bandwidth := range box.Bandwidths {

			if _, existing := bandwidthMap[box.BoxId]; !existing {
				bandwidthMap[box.BoxId] = make([]*model.BoxBandwidth, 0)
			}

			bandwidthMap[box.BoxId] = append(bandwidthMap[box.BoxId], &model.BoxBandwidth{
				Username:      username,
				BoxId:         box.BoxId,
				SupplierBoxId: box.SupplierBoxId,
				Upload:        bandwidth.Upload,
				Download:      bandwidth.Download,
				Time:          bandwidth.Time,
			})
		}
	}

	for _, bandwidths := range bandwidthMap {
		err := dao.BulkUpsertBoxBandwidth(ctx, bandwidths)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DataService) syncBoxBandwidth(pi *model.PaiNetInfo, date string) error {
	var (
		count    int64 = 0
		ctx            = context.Background()
		pageSize       = 100
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes bandwidth")

		total, boxes, err := dao.GetBoxesList(ctx, pi.PaiUsername, nil, nil, int64(page), int64(pageSize))
		if err != nil {
			return errors.Wrapf(err, "query boxes list")
		}

		var boxIds []string
		for _, box := range boxes {
			boxIds = append(boxIds, "boxId="+box.BoxId)
		}

		url := fmt.Sprintf("%s/boxsupplier/v1/box/bandwidth?date=%s&%s", paiNetBaseUrl, date, strings.Join(boxIds, "&"))

		fmt.Println(url)
		response, err := doRequest(pi, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes income")
		}

		fmt.Println("=>", string(response))

		var res GetBoxBandwidthResponse
		if err := json.Unmarshal(response, &res); err != nil {
			return err
		}

		if len(res.BoxBandwidths) == 0 {
			return nil
		}

		if err := d.saveBoxBandwidth(ctx, pi.PaiUsername, res.BoxBandwidths); err != nil {
			return err
		}

		count += int64(len(res.BoxBandwidths))
		if count >= total {
			break
		}
	}

	log.Info("Synchronization of boxes bandwidth completed successfully.")

	return nil
}

type GetBoxQualitiesResponse struct {
	BoxQualities []*BoxQualities `json:"boxQualities"`
}

type BoxQualities struct {
	BoxId         string              `json:"boxId"`
	SupplierBoxId string              `json:"supplierBoxId"`
	Qualities     []*model.BoxQuality `json:"qualities"`
}

func (d *DataService) saveBoxQualities(ctx context.Context, username string, boxQualities []*BoxQualities) error {

	qualitiesMap := make(map[string][]*model.BoxQuality)

	for _, box := range boxQualities {
		for _, quality := range box.Qualities {

			if _, existing := qualitiesMap[box.BoxId]; !existing {
				qualitiesMap[box.BoxId] = make([]*model.BoxQuality, 0)
			}

			qualitiesMap[box.BoxId] = append(qualitiesMap[box.BoxId], &model.BoxQuality{
				Username:      username,
				BoxId:         box.BoxId,
				SupplierBoxId: box.SupplierBoxId,
				PacketLoss:    quality.PacketLoss,
				TcpNatType:    quality.TcpNatType,
				UdpNatType:    quality.UdpNatType,
				CpuUsage:      quality.CpuUsage,
				MemoryUsage:   quality.MemoryUsage,
				DiskUsage:     quality.DiskUsage,
				Time:          quality.Time,
			})
		}
	}

	for _, bandwidths := range qualitiesMap {
		err := dao.BulkUpsertBoxQualities(ctx, bandwidths)
		if err != nil {
			return err
		}
	}

	return nil

}

func (d *DataService) syncBoxQualities(pi *model.PaiNetInfo, date string) error {
	var (
		count    int64 = 0
		ctx            = context.Background()
		pageSize       = 100
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes qualities")

		total, boxes, err := dao.GetBoxesList(ctx, pi.PaiUsername, nil, nil, int64(page), int64(pageSize))
		if err != nil {
			return errors.Wrapf(err, "query boxes list")
		}

		var boxIds []string
		for _, box := range boxes {
			boxIds = append(boxIds, "boxId="+box.BoxId)
		}

		url := fmt.Sprintf("%s/boxsupplier/v1/box/quality?date=%s&%s", paiNetBaseUrl, date, strings.Join(boxIds, "&"))

		response, err := doRequest(pi, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes income")
		}

		//fmt.Println("=>", string(response))

		var res GetBoxQualitiesResponse
		if err := json.Unmarshal(response, &res); err != nil {
			return err
		}

		if len(res.BoxQualities) == 0 {
			return nil
		}

		if err := d.saveBoxQualities(ctx, pi.PaiUsername, res.BoxQualities); err != nil {
			return err
		}

		count += int64(len(res.BoxQualities))
		if count >= total {
			break
		}
	}

	log.Info("Synchronization of boxes qualities completed successfully.")

	return nil
}
