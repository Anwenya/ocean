package test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestSqliteEncrypt(t *testing.T) {

	// 不指定版本号默认使用AES-256bit加密
	//dsn := "file:test_cipher_1.db?_cipher=sqlcipher&_key=123456"
	dsn := "file:test_cipher_chacha20.db?_cipher=chacha20&_key=123456"

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	//Migrate the schema
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	//Create
	err = db.Create(&User{Extra: RandomJSON()}).Error
	require.NoError(t, err)

	var user User

	err = db.Find(&user).Error
	require.NoError(t, err)
	fmt.Println(user)
}
