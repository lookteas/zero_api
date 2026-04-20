package logic

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"api/internal/types"
	"api/model"
)

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func userToInfo(user *model.Users) types.UserInfo {
	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}

	mobile := ""
	if user.Mobile.Valid {
		mobile = user.Mobile.String
	}

	return types.UserInfo{
		Id:       user.Id,
		Account:  user.Account,
		Email:    email,
		Mobile:   mobile,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Status:   int64(user.Status),
	}
}

func loginRespFromUser(user *model.Users) *types.LoginResp {
	return &types.LoginResp{
		Code:    0,
		Message: "ok",
		Data: types.LoginData{
			AccessToken:  "dev-token-user-1",
			RefreshToken: "dev-refresh-token-user-1",
			AccessExpire: 86400,
			User:         userToInfo(user),
		},
	}
}

func buildUser(account, password, nickname, email, mobile string) *model.Users {
	return &model.Users{
		Account:      account,
		Email:        nullString(email),
		Mobile:       nullString(mobile),
		PasswordHash: hashPassword(password),
		Nickname:     nickname,
		Avatar:       "",
		Status:       1,
		LastLoginAt:  sql.NullTime{Time: time.Now(), Valid: true},
	}
}
