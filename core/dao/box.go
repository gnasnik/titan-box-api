package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-box-api/core/generated/model"
	"github.com/jmoiron/sqlx"
)

func BulkUpsertBoxes(ctx context.Context, boxes []*model.Box) error {
	query := `INSERT INTO box (username, boxId, supplierBoxId, online, tcpNatType, udpNatType, publicIp, privateIp, isp, province,
		city, cpuArch, cpuCores, memorySize, os, pluginVersion, pluginDeployTime, processStatus,
		fault, upload, download, diskUsage, upnp, notDeployReason, reportUpBandwidth, planTask, pressBandwidth, remark,
		icmpv6Out, createdAt, updatedAt )
	VALUES(:username, :boxId, :supplierBoxId, :online, :tcpNatType, :udpNatType, :publicIp, :privateIp, :isp, :province,
		:city, :cpuArch, :cpuCores, :memorySize, :os, :pluginVersion, :pluginDeployTime, :processStatus,
		:fault, :upload, :download, :diskUsage, :upnp, :notDeployReason, :reportUpBandwidth, :planTask, :pressBandwidth, :remark,
		:icmpv6Out, now(), now() ) ON DUPLICATE KEY UPDATE 
		online = VALUES(online), tcpNatType = VALUES(tcpNatType), udpNatType = VALUES(udpNatType), publicIp = VALUES(publicIp),
		privateIp = VALUES(privateIp), isp = VALUES(isp), province = VALUES(province), city = VALUES(city), cpuArch = VALUES(cpuArch),
	    cpuCores = VALUES(cpuCores), memorySize = VALUES(memorySize), os = VALUES(os), pluginVersion = VALUES(pluginVersion), 
	    pluginDeployTime = VALUES(pluginDeployTime),pluginDeployTime = VALUES(pluginDeployTime),
	    processStatus = VALUES(processStatus), fault = VALUES(fault), upload = VALUES(upload), download = VALUES(download), diskUsage = VALUES(diskUsage),
	 	upnp = VALUES(upnp), notDeployReason = VALUES(notDeployReason), reportUpBandwidth = VALUES(reportUpBandwidth), planTask = VALUES(planTask), 
		pressBandwidth = VALUES(pressBandwidth), remark = VALUES(remark), icmpv6Out = VALUES(icmpv6Out), updatedAt = now()`

	if _, err := DB.NamedExecContext(ctx, query, boxes); err != nil {
		return err
	}

	return nil
}

func BulkUpsertBoxDiskInfo(ctx context.Context, diskInfo []*model.DiskInfo) error {
	query := `INSERT INTO box_diskinfo(boxId, supplierBoxId, diskId, diskSize, diskMedia, diskUsed)
	VALUES(:boxId, :supplierBoxId, :diskId, :diskSize, :diskMedia, :diskUsed)  ON DUPLICATE KEY UPDATE 
	diskSize = VALUES(diskSize), diskMedia = VALUES(diskMedia), diskUsed = VALUES(diskUsed)`

	if _, err := DB.NamedExecContext(ctx, query, diskInfo); err != nil {
		return err
	}

	return nil
}

func BulkUpsertBoxDayIncome(ctx context.Context, diskInfo []*model.BoxIncome) error {
	query := `INSERT INTO box_income(username, boxId, supplierBoxId, date, remark, bw, bwAmount, amount, activityIncome, distPercent, inviterId, updatedAt)
	VALUES(:username, :boxId, :supplierBoxId, :date, :remark, :bw, :bwAmount, :amount, :activityIncome, :distPercent, :inviterId, now())  ON DUPLICATE KEY UPDATE 
	remark = VALUES(remark), bw = VALUES(bw), amount = VALUES(amount), updatedAt = now()`

	if _, err := DB.NamedExecContext(ctx, query, diskInfo); err != nil {
		return err
	}

	return nil
}

func BulkUpsertBoxBandwidth(ctx context.Context, bandwidths []*model.BoxBandwidth) error {
	query := `INSERT INTO box_bandwidth(username, boxId, supplierBoxId, time, upload, download, updatedAt)
	VALUES(:username, :boxId, :supplierBoxId, :time, :upload, :download, now())  ON DUPLICATE KEY UPDATE 
	upload = VALUES(upload), download = VALUES(download), updatedAt = now()`

	if _, err := DB.NamedExecContext(ctx, query, bandwidths); err != nil {
		return err
	}

	return nil
}

func BulkUpsertBoxQualities(ctx context.Context, qualities []*model.BoxQuality) error {
	query := `INSERT INTO box_quality(username, boxId, supplierBoxId, time, packetLoss, tcpNatType, udpNatType, cpuUsage, memoryUsage, diskUsage, updatedAt)
	VALUES(:username, :boxId, :supplierBoxId, :time, :packetLoss, :tcpNatType, :udpNatType, :cpuUsage, :memoryUsage, :diskUsage, now())  ON DUPLICATE KEY UPDATE 
	packetLoss = VALUES(packetLoss), tcpNatType = VALUES(tcpNatType), udpNatType = VALUES(udpNatType), cpuUsage = VALUES(cpuUsage), memoryUsage = VALUES(memoryUsage), 
	diskUsage = VALUES(diskUsage), updatedAt = now()`

	if _, err := DB.NamedExecContext(ctx, query, qualities); err != nil {
		return err
	}

	return nil
}

func GetBoxesList(ctx context.Context, username string, boxIds []string, supplierBoxIds []string, page, pageSize int64) (int64, []*model.Box, error) {
	query := `select b.*, ifnull(d.diskSize,'') as diskSize, ifnull(d.diskMedia,'') as diskMedia, ifnull(d.diskUsed,'') as diskUsed from (%s) b left join box_diskinfo d on b.boxId = d.boxId `

	var (
		where = `where username = ? `
		args  = []interface{}{username}
	)

	if len(boxIds) > 0 {
		boxQuery, boxArgs, err := sqlx.In(` and boxId in (?)`, boxIds)
		if err != nil {
			return 0, nil, err
		}
		where += boxQuery
		args = append(args, boxArgs...)
	}

	if len(supplierBoxIds) > 0 {
		supplierQuery, supplierArgs, err := sqlx.In(` and supplierBoxId in (?)`, supplierBoxIds)
		if err != nil {
			return 0, nil, err
		}
		where += supplierQuery
		args = append(args, supplierArgs...)
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	subQry := `SELECT * from box `
	countQry := `SELECT count(1) from box ` + where

	var total int64
	if err := DB.GetContext(ctx, &total, countQry, args...); err != nil {
		return 0, nil, err
	}

	query = fmt.Sprintf(query, subQry+where+fmt.Sprintf(" limit %d offset %d", limit, offset))

	type BoxAndDisk struct {
		model.Box
		model.DiskInfo
	}

	var bds []*BoxAndDisk
	if err := DB.SelectContext(ctx, &bds, query, args...); err != nil {
		return 0, nil, err
	}

	var out []*model.Box
	boxGroup := make(map[string]*model.Box)
	for _, d := range bds {
		_, ok := boxGroup[d.Box.BoxId]
		if !ok {
			boxGroup[d.Box.BoxId] = &d.Box
			boxGroup[d.Box.BoxId].DiskInfos = make([]*model.DiskInfo, 0)
			out = append(out, boxGroup[d.Box.BoxId])
		}
		boxGroup[d.Box.BoxId].DiskInfos = append(boxGroup[d.Box.BoxId].DiskInfos, &d.DiskInfo)
	}

	return total, out, nil
}

func GetBoxIncomeV2(ctx context.Context, username string, boxIds, remarks, supplierBoxIds []string, start, end string, page, pageSize int64) (int64, int64, []*model.BoxIncome, error) {
	query := `select * from box_income `

	var (
		where = `where username = ? and date >= ? and date <= ? `
		args  = []interface{}{username, start, end}
	)

	if len(boxIds) > 0 {
		boxQuery, boxArgs, err := sqlx.In(` and boxId in (?)`, boxIds)
		if err != nil {
			return 0, 0, nil, err
		}
		where += boxQuery
		args = append(args, boxArgs...)
	}

	if len(supplierBoxIds) > 0 {
		supplierQuery, supplierArgs, err := sqlx.In(` and supplierBoxId in (?)`, supplierBoxIds)
		if err != nil {
			return 0, 0, nil, err
		}
		where += supplierQuery
		args = append(args, supplierArgs...)
	}

	if len(remarks) > 0 {
		remarkQuery, remarkArgs, err := sqlx.In(` and remarks in (?)`, remarks)
		if err != nil {
			return 0, 0, nil, err
		}
		where += remarkQuery
		args = append(args, remarkArgs...)
	}

	countQry := `SELECT sum(amount) as total, count(1) as totalNum from box_income ` + where
	type Count struct {
		Total    int64 `db:"total"`
		TotalNum int64 `db:"totalNum"`
	}

	var count Count
	if err := DB.GetContext(ctx, &count, countQry, args...); err != nil {
		return 0, 0, nil, err
	}

	limit := pageSize
	offset := (page - 1) * pageSize
	query = query + where + fmt.Sprintf(" order by date desc limit %d offset %d", limit, offset)

	var out []*model.BoxIncome
	if err := DB.SelectContext(ctx, &out, query, args...); err != nil {
		return 0, 0, nil, err
	}

	return count.TotalNum, count.Total, out, nil
}

func GetBoxBandwidth(ctx context.Context, username string, boxIds, supplierBoxIds []string, start, end int64) ([]*model.BoxBandwidth, error) {
	query := `select * from box_bandwidth `

	var (
		where = `where username = ? and time >= ? and time <= ? `
		args  = []interface{}{username, start, end}
	)

	if len(boxIds) > 0 {
		boxQuery, boxArgs, err := sqlx.In(` and boxId in (?)`, boxIds)
		if err != nil {
			return nil, err
		}
		where += boxQuery
		args = append(args, boxArgs...)
	}

	if len(supplierBoxIds) > 0 {
		supplierQuery, supplierArgs, err := sqlx.In(` and supplierBoxId in (?)`, supplierBoxIds)
		if err != nil {
			return nil, err
		}
		where += supplierQuery
		args = append(args, supplierArgs...)
	}

	query = query + where
	fmt.Println("query=>", query)

	var out []*model.BoxBandwidth
	if err := DB.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, err
	}

	return out, nil
}

func GetBoxQualities(ctx context.Context, username string, boxIds, supplierBoxIds []string, start, end int64) ([]*model.BoxQuality, error) {
	query := `select * from box_quality `

	var (
		where = `where username = ? and time >= ? and time <= ? `
		args  = []interface{}{username, start, end}
	)

	if len(boxIds) > 0 {
		boxQuery, boxArgs, err := sqlx.In(` and boxId in (?)`, boxIds)
		if err != nil {
			return nil, err
		}
		where += boxQuery
		args = append(args, boxArgs...)
	}

	if len(supplierBoxIds) > 0 {
		supplierQuery, supplierArgs, err := sqlx.In(` and supplierBoxId in (?)`, supplierBoxIds)
		if err != nil {
			return nil, err
		}
		where += supplierQuery
		args = append(args, supplierArgs...)
	}

	query = query + where
	fmt.Println("query=>", query)

	var out []*model.BoxQuality
	if err := DB.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, err
	}

	return out, nil
}

func GetUserKeys(ctx context.Context) ([]*model.PaiUserKey, error) {
	query := `select * from pai_userkey`

	var out []*model.PaiUserKey
	if err := DB.SelectContext(ctx, &out, query); err != nil {
		return nil, err
	}

	return out, nil
}

func GetUserKeyByAPIKey(ctx context.Context, key string) (*model.PaiUserKey, error) {
	query := `select * from pai_userkey where apiKey = ?`

	var out model.PaiUserKey
	if err := DB.GetContext(ctx, &out, query, key); err != nil {
		return nil, err
	}

	return &out, nil
}
