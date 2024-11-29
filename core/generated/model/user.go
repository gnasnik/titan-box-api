package model

import "time"

type User struct {
	Uid          int       `json:"uid" db:"uid"`
	Username     string    `json:"username" db:"username"`
	Password     string    `json:"-" db:"password"`
	AppKey       string    `json:"appKey" db:"appKey"`
	AppSecret    string    `json:"appSecret" db:"appSecret"`
	SupplierType int       `json:"supplierType" db:"supplierType"`
	PhoneNumber  string    `json:"phoneNumber" db:"phoneNumber"`
	BillingCycle string    `json:"billingCycle" db:"billingCycle"`
	ParentId     string    `json:"parentId" db:"parentId"`
	DistPercent  int       `json:"distPercent" db:"distPercent"`
	CanInvite    bool      `json:"canInvite" db:"canInvite"`
	InviterType  int       `json:"inviterType" db:"inviterType"`
	CreatedAt    time.Time `json:"-" db:"createdAt"`
}

type PaiNetInfo struct {
	APIKey      string `db:"apiKey"`
	APISecret   string `db:"apiSecret"`
	Username    string `db:"username"`
	PaiUsername string `db:"paiUsername"`
	Status      int64  `db:"status"`
}
