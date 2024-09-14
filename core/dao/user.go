package dao

import (
	"context"
	"github.com/gnasnik/titan-box-api/core/generated/model"
)

func CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO user (username, password, appKey, appSecret, supplierType, phoneNumber, billingCycle, parentId, distPercent, canInvite, inviterType, createdAt)
			VALUES (:username, :password, :appKey, :appSecret, :supplierType, :phoneNumber, :billingCycle, :parentId, :distPercent, :canInvite, :inviterType, now())`

	_, err := DB.NamedExecContext(ctx, query, user)

	return err
}

func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT * FROM user WHERE username = ?`

	var out model.User
	if err := DB.QueryRowxContext(ctx, query, username).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}
