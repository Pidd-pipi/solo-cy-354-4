package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// seed populates default users, products and book exchanges on first boot.
func seed(ctx context.Context, db *gorm.DB, logger *slog.Logger) error {
	var total int64
	if err := db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return fmt.Errorf("seed count users: %w", err)
	}
	if total > 0 {
		logger.Info("seed skipped: users already exist", slog.Int64("count", total))
		return nil
	}
	hash := func(p string) (string, error) {
		b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	type seedUser struct {
		phone, password, nickname, role, campus string
		credit                                  int
	}
	defs := []seedUser{
		{phone: "13700000001", password: "123456", nickname: "小明同学", role: constants.UserRoleStudent, campus: "东校区", credit: 100},
		{phone: "13700000002", password: "123456", nickname: "阿珍", role: constants.UserRoleStudent, campus: "西校区", credit: 120},
		{phone: "13700000003", password: "123456", nickname: "二手达人", role: constants.UserRoleStudent, campus: "南校区", credit: 90},
		{phone: "13800000001", password: "admin123", nickname: "平台管理员", role: constants.UserRoleAdmin, campus: "东校区", credit: 300},
	}
	users := make([]model.User, 0, len(defs))
	for _, d := range defs {
		h, err := hash(d.password)
		if err != nil {
			return fmt.Errorf("seed hash user %s: %w", d.phone, err)
		}
		users = append(users, model.User{
			Phone: d.phone, PasswordHash: h, Nickname: d.nickname,
			Role: d.role, Campus: d.campus, CreditScore: d.credit,
		})
	}
	if err := db.WithContext(ctx).Create(&users).Error; err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	products := []model.Product{
		{SellerID: users[0].ID, Title: "高等数学第六版", Description: "九成新，有少量笔记", Price: 15, Category: constants.ProductCategoryBooks, Condition: "九成新", Campus: "东校区", TradeLocation: "图书馆门口", Images: "", Status: constants.ProductStatusOnSale},
		{SellerID: users[1].ID, Title: "iPad Air 5", Description: "95新，带笔", Price: 2800, Category: constants.ProductCategoryElectronics, Condition: "95新", Campus: "西校区", TradeLocation: "三食堂", Images: "", Status: constants.ProductStatusOnSale},
		{SellerID: users[2].ID, Title: "宿舍小台灯", Description: "暖光护眼", Price: 20, Category: constants.ProductCategoryDaily, Condition: "全新", Campus: "南校区", TradeLocation: "南门快递点", Images: "", Status: constants.ProductStatusOnSale},
		{SellerID: users[0].ID, Title: "毕业季正装一套", Description: "M码 黑色西服", Price: 180, Category: constants.ProductCategoryClothing, Condition: "九成新", Campus: "东校区", TradeLocation: "东门", Images: "", Status: constants.ProductStatusOnSale},
	}
	if err := db.WithContext(ctx).Create(&products).Error; err != nil {
		return fmt.Errorf("seed products: %w", err)
	}
	exchanges := []model.BookExchange{
		{UserID: users[0].ID, OfferBook: "数据结构", WantBook: "计算机网络", Description: "希望交换", Status: "open"},
		{UserID: users[1].ID, OfferBook: "计算机网络", WantBook: "数据结构", Description: "同城交换", Status: "open"},
	}
	if err := db.WithContext(ctx).Create(&exchanges).Error; err != nil {
		return fmt.Errorf("seed exchanges: %w", err)
	}
	logger.Info(fmt.Sprintf(constants.LogSeedingCompleted, len(users), len(products)))
	return nil
}
