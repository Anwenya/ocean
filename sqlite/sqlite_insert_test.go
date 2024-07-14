package test

import (
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"testing"
	"time"
)

type User struct {
	Id    int64  `gorm:"primaryKey"`
	Cid   int64  `gorm:"index"`
	Tid   int64  `gorm:"index"`
	Eid   int64  `gorm:"index"`
	Pid   int64  `gorm:"index"`
	Oid   int64  `gorm:"index"`
	Extra string `gorm:"type:text"`
}

func TestInsertData(t *testing.T) {

	//dsn := "test_smc_sqlcipher_aes256.db?_cipher=sqlcipher&_key=123456"
	dsn := "test_smc_chacha20.db?_cipher=chacha20&_key=123456"

	// 连接到数据库（如果不存在则创建）
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// 自动迁移模式，创建表
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}

	// 批量插入随机生成的JSON字符串
	var users []User
	for i := 0; i < 1000000; i++ {
		users = append(users, User{Extra: RandomJSON()})
	}
	log.Println("生成数据完成,开始插入")
	startTime := time.Now()
	for i := 0; i < 1000000; i += 1000 {
		u := users[i : i+1000]
		result := db.Create(&u)
		if result.Error != nil {
			log.Fatal(result.Error)
		}
	}

	log.Printf("批量插入完成,耗时:%f\n", time.Since(startTime).Seconds())
}

// RandomJSON 生成随机的JSON字符串
func RandomJSON() string {
	data := map[string]interface{}{
		"id":         rand.Intn(1000),
		"name":       randomString(10),
		"age":        rand.Intn(100),
		"time":       time.Now().Format(time.RFC3339),
		"email":      randomString(5) + "@example.com",
		"address":    randomString(15),
		"city":       randomString(8),
		"country":    randomString(6),
		"phone":      randomString(10),
		"zipcode":    randomString(5),
		"latitude":   rand.Float64()*180 - 90,
		"longitude":  rand.Float64()*360 - 180,
		"company":    randomString(12),
		"website":    "https://" + randomString(10) + ".com",
		"job_title":  randomString(10),
		"department": randomString(8),
		"birthday":   time.Now().AddDate(-rand.Intn(50), 0, 0).Format("2006-01-02"),
		"notes":      randomChineseString(rand.Intn(500) + 50),
		"active":     rand.Intn(2) == 1,
		"created_at": time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	return string(jsonData)
}

// randomString 生成指定长度的随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// randomChineseString 生成指定长度的随机中文字符串
func randomChineseString(length int) string {
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	runes := make([]rune, length)
	for i := range runes {
		runes[i] = rune(seededRand.Intn(0x9FFF-0x4E00) + 0x4E00)
	}
	return string(runes)
}
