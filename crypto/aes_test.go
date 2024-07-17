package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func encrypt(key []byte, plaintext []byte) []byte {
	// 创建一个AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个加密器
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal(err)
	}

	// 使用加密器加密数据
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext
}

func decrypt(key []byte, ciphertext []byte) []byte {
	// 创建一个AES加密块
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}

	// 解析初始化向量
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 创建一个解密器
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

func TestAES(t *testing.T) {
	// key
	key := []byte("B6gCt0DETDo8y5sq")

	// 读取配置文件
	configData, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// 加密配置文件数据
	encryptedData := encrypt(key, configData)

	// 写入配置文件
	err = os.WriteFile("config-encrypted.txt", encryptedData, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// 读取加密后的配置文件数据
	encryptedData, err = os.ReadFile("config-encrypted.txt")
	if err != nil {
		log.Fatal(err)
	}
	//
	//解密配置文件数据
	decryptedData := decrypt(key, encryptedData)

	// 使用解密后的数据进行操作
	// ...
	a := string(decryptedData)
	fmt.Println(a)

}
