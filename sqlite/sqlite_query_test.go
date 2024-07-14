package test

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"testing"
	"time"
)

func TestSortByJson(t *testing.T) {
	//dsn := "test_normal.db"
	//dsn := "test_smc_sqlcipher_aes256.db?_cipher=sqlcipher&_key=123456"
	dsn := "test_smc_chacha20.db?_cipher=chacha20&_key=123456"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("开始查询")
	startTime := time.Now()
	var users []User
	err = db.Raw(
		`SELECT * FROM users where json_extract(extra, '$.notes') like "%檎%" 
        	AND id > 100 AND id < 100000
			ORDER BY json_extract(extra, '$.age') ASC, json_extract(extra, '$.notes') ASC
			limit 0, 1000`,
	).Scan(&users).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("查询结束 耗时%f\n", time.Since(startTime).Seconds())
	//for _, user := range users {
	//	log.Println(user.Extra)
	//}
	log.Println(len(users))
}
