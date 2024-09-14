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
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// API endpoint for fetching boxes
	paiNetBaseUrl = "https://openapi.painet.work"

	defaultPageSize = 200
)

type DataService struct {
	Interval time.Duration
}

type Response struct {
	Code    int64  `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

func NewDataService() *DataService {
	return &DataService{
		Interval: 30 * time.Minute,
	}
}

func (d *DataService) Run(ctx context.Context) {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	d.startSync()

	for {
		select {
		case <-ticker.C:
			d.startSync()
		case <-ctx.Done():
			break
		}
	}
}

func (d *DataService) startSync() {
	var eg errgroup.Group

	apiKeys, err := dao.GetUserKeys(context.Background())
	if err != nil {
		log.Errorf("get user keys: %v", err)
		return
	}

	for _, userKey := range apiKeys {
		if userKey.Status != 0 {
			continue
		}

		//d.startSyncBoxDayHistory(userKey)

		eg.Go(func() error {
			if err := d.syncBoxList(userKey); err != nil {
				return errors.Wrapf(err, "sync box list error")
			}
			return nil
		})

		eg.Go(func() error {
			start := time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
			end := time.Now().Format(time.DateOnly)

			if err := d.syncBoxIncome(userKey, start, end); err != nil {
				return errors.Wrapf(err, "sync box income error")
			}
			return nil
		})

		eg.Go(func() error {
			today := time.Now().Format(time.DateOnly)
			if err := d.syncBoxBandwidth(userKey, today); err != nil {
				return errors.Wrapf(err, "sync box bandwidth error")
			}
			return nil
		})

		eg.Go(func() error {
			today := time.Now().Format(time.DateOnly)
			if err := d.syncBoxQualities(userKey, today); err != nil {
				return errors.Wrapf(err, "sync box qualities error")
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		log.Errorf("Synchronization error: %v", err)
	}
}

type GetBoxListResponse struct {
	Boxes []*model.Box `json:"boxes"`
	Total string       `json:"total"`
}

func (d *DataService) syncBoxList(userKey *model.PaiUserKey) error {
	var (
		count int64 = 0
		ctx         = context.Background()
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes on page: %d", page)

		url := fmt.Sprintf("%s/boxsupplier/v1/box/list?page=%d&pageSize=%d", paiNetBaseUrl, page, defaultPageSize)
		response, err := d.doRequest(userKey, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes from page %d", page)
		}

		var res GetBoxListResponse
		if err := json.Unmarshal(response, &res); err != nil {
			return err
		}

		total, err := strconv.ParseInt(res.Total, 10, 64)
		if err != nil {
			return err
		}

		if total == 0 {
			return nil
		}

		if err := d.saveBoxList(ctx, userKey.Username, res.Boxes); err != nil {
			return err
		}

		count += int64(len(res.Boxes))
		if count >= total {
			break
		}
	}

	log.Info("Synchronization of boxes list completed successfully.")

	return nil
}

func (d *DataService) saveBoxList(ctx context.Context, username string, boxes []*model.Box) error {
	var diskInfos []*model.DiskInfo

	for _, box := range boxes {
		box.Username = username

		for _, diskInfo := range box.DiskInfos {
			diskInfo.BoxId = box.BoxId
			diskInfo.SupplierBoxId = box.SupplierBoxId
			diskInfos = append(diskInfos, diskInfo)
		}
		diskInfos = append(diskInfos)
	}

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

func (d *DataService) syncBoxIncome(userKey *model.PaiUserKey, start, end string) error {
	var (
		count int64 = 0
		ctx         = context.Background()
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes income")

		url := fmt.Sprintf("%s/boxsupplier/v1/supplier/income_v2?start=%s&end=%s&pageSize=%d&pageIndex=%d", paiNetBaseUrl, start, end, defaultPageSize, page)
		response, err := d.doRequest(userKey, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes income")
		}

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

		if err := d.saveBoxIncome(ctx, userKey.Username, res.BoxIncome); err != nil {
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

func (d *DataService) doRequest(userKey *model.PaiUserKey, url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	sign := generateMD5Hash(userKey.APISecret, userKey.Username, timestamp)

	request.Header.Set("ak", userKey.APIKey)
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

func (d *DataService) StartSyncBoxHistory(userKey *model.PaiUserKey) {
	nowYear, nowMonth, nowDay := time.Now().Date()
	startTime := time.Date(2024, 8, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(nowYear, nowMonth, nowDay, 0, 0, 0, 0, time.Local)

	for startTime.Before(endTime) {
		start := startTime.Format(time.DateOnly)
		end := startTime.AddDate(0, 0, 10).Format(time.DateOnly)
		startTime = startTime.AddDate(0, 0, 10)

		if err := d.syncBoxIncome(userKey, start, end); err != nil {
			log.Infof("syncBoxIncome: start: %s, end: %s, %v", start, end, err)
		}
	}

}

func (d *DataService) StartSyncBoxDayHistory(userKey *model.PaiUserKey) {
	nowYear, nowMonth, nowDay := time.Now().Date()
	startTime := time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local)
	endTime := time.Date(nowYear, nowMonth, nowDay, 0, 0, 0, 0, time.Local)

	for startTime.Before(endTime) {
		start := startTime.Format(time.DateOnly)
		end := startTime.AddDate(0, 0, 1).Format(time.DateOnly)
		startTime = startTime.AddDate(0, 0, 1)

		fmt.Println(" ==>", start)
		if err := d.syncBoxBandwidth(userKey, start); err != nil {
			fmt.Println("==>", err)
			log.Infof("syncBoxBandwidth: start: %s, end: %s, %v", start, end, err)
		}

		if err := d.syncBoxQualities(userKey, start); err != nil {
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
	var bandwidths []*model.BoxBandwidth
	for _, box := range boxBandwidths {
		for _, bandwidth := range box.Bandwidths {
			bandwidths = append(bandwidths, &model.BoxBandwidth{
				Username:      username,
				BoxId:         box.BoxId,
				SupplierBoxId: box.SupplierBoxId,
				Upload:        bandwidth.Upload,
				Download:      bandwidth.Download,
				Time:          bandwidth.Time,
			})
		}
	}

	err := dao.BulkUpsertBoxBandwidth(ctx, bandwidths)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataService) syncBoxBandwidth(userKey *model.PaiUserKey, date string) error {
	var (
		count    int64 = 0
		ctx            = context.Background()
		pageSize       = 100
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes bandwidth")

		total, boxes, err := dao.GetBoxesList(ctx, userKey.Username, nil, nil, int64(page), int64(pageSize))
		if err != nil {
			return errors.Wrapf(err, "query boxes list")
		}

		var boxIds []string
		for _, box := range boxes {
			boxIds = append(boxIds, "boxId="+box.BoxId)
		}

		url := fmt.Sprintf("%s/boxsupplier/v1/box/bandwidth?date=%s&%s", paiNetBaseUrl, date, strings.Join(boxIds, "&"))

		response, err := d.doRequest(userKey, url)
		if err != nil {
			return errors.Wrapf(err, "Failed to fetch boxes income")
		}

		//fmt.Println("=>", string(response))

		var res GetBoxBandwidthResponse
		if err := json.Unmarshal(response, &res); err != nil {
			return err
		}

		if len(res.BoxBandwidths) == 0 {
			return nil
		}

		if err := d.saveBoxBandwidth(ctx, userKey.Username, res.BoxBandwidths); err != nil {
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
	var qualities []*model.BoxQuality
	for _, box := range boxQualities {
		for _, quality := range box.Qualities {
			qualities = append(qualities, &model.BoxQuality{
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

	err := dao.BulkUpsertBoxQualities(ctx, qualities)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataService) syncBoxQualities(userKey *model.PaiUserKey, date string) error {
	var (
		count    int64 = 0
		ctx            = context.Background()
		pageSize       = 100
	)

	for page := 1; ; page++ {
		log.Infof("Starting to query boxes qualities")

		total, boxes, err := dao.GetBoxesList(ctx, userKey.Username, nil, nil, int64(page), int64(pageSize))
		if err != nil {
			return errors.Wrapf(err, "query boxes list")
		}

		var boxIds []string
		for _, box := range boxes {
			boxIds = append(boxIds, "boxId="+box.BoxId)
		}

		url := fmt.Sprintf("%s/boxsupplier/v1/box/quality?date=%s&%s", paiNetBaseUrl, date, strings.Join(boxIds, "&"))

		response, err := d.doRequest(userKey, url)
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

		if err := d.saveBoxQualities(ctx, userKey.Username, res.BoxQualities); err != nil {
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
