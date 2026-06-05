package usecase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"openbridge/backend/internal/config"
	"time"

	"gorm.io/gorm"
)

type UserUseCase struct {
	config *config.Config
	db     *gorm.DB
}

func NewUserUseCase(config *config.Config, db *gorm.DB) *UserUseCase {
	return &UserUseCase{
		config: config,
		db:     db,
	}
}

func (uc *UserUseCase) Login(username, password string) error {
	
	// HTTP 客户端配置，设置超时时间为10秒
	client := &http.Client{Timeout: 10 * time.Second}

	// 构造登录请求的payload，包含用户名和密码
	payload := map[string]string{
		"username":   username,
		"password":   password,
	}

	// 将payload转换为JSON格式
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 创建一个新的HTTP POST请求，目标URL为OpenList的登录接口，并将JSON数据作为请求体
	req, err := http.NewRequest("POST", uc.config.OpenList.BaseURL + "/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// 设置请求头，指定内容类型为JSON，并设置User-Agent
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "OpenBridge/1.0")

	// 发送HTTP请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// 解析响应体，提取登录结果
	// {
    // "code": 200,
    // "message": "success",
    // "data": {
    //     "token": "xxxx"
	// }
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// 将登录成功后返回的Token保存到配置中，以便后续请求使用
	uc.config.OpenList.Token = result.Data.Token

	return nil
}

func (uc *UserUseCase) Reset() error { 

	err := uc.ClearAllTables(uc.db)
	if err != nil {
	    return err
	}
	return nil
}

func (uc *UserUseCase) ClearAllTables(db *gorm.DB) error {
    // 获取所有表名
    var tables []string
    db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables)
    
    // 事务清空
    return db.Transaction(func(tx *gorm.DB) error {
        tx.Exec("PRAGMA foreign_keys = OFF")
        for _, table := range tables {
            tx.Exec(fmt.Sprintf("DELETE FROM %s", table))
            // tx.Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name = ?", table))
        }
        tx.Exec("PRAGMA foreign_keys = ON")
        tx.Exec("VACUUM")
        return nil
    })
}