package test

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type User struct {
	Id          int64
	Name        string
	Description string
}

func getDB() *sql.DB {
	sql.Register(
		"sqlite3_simple",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"G:\\Solutions\\ocean\\sqlite\\fts\\simple",
			},
		},
	)

	db, err := sql.Open(
		"sqlite3_simple",
		"fts_cipher_test.db?_cipher=chacha20&_key=123456",
	)
	if err != nil {
		panic(err)
	}
	return db
}

func TestCipherFTS5CreateTable(t *testing.T) {
	db := getDB()
	defer db.Close()

	// 建表
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS user (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	name TEXT,
    	description TEXT
	)
    `)
	require.NoError(t, err)

	// 索引表
	_, err = db.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS user_fts USING fts5(
	       name, 
	       description, 
	       content=user,
	       content_rowid=id, 
	       tokenize="simple"
   )
	`)
	require.NoError(t, err)

	// 触发器
	_, err = db.Exec(`
	CREATE TRIGGER IF NOT EXISTS user_fts_i AFTER INSERT ON user BEGIN
	INSERT INTO user_fts(rowid, name, description) 
	VALUES (new.id, new.name, new.description);
	END;
	
	CREATE TRIGGER IF NOT EXISTS user_fts_d AFTER DELETE ON user BEGIN
	INSERT INTO user_fts(user_fts, rowid, name, description) 
	VALUES('delete', old.id, old.name, old.description);
	END;
	
	CREATE TRIGGER IF NOT EXISTS user_fts_u AFTER UPDATE ON user BEGIN
	INSERT INTO user_fts(user_fts, rowid, name, description) 
	VALUES('delete', old.id, old.name, old.description);

	INSERT INTO user_fts(rowid, name, description) 
	VALUES (new.id, new.name, new.description);
	END;
	`)
	require.NoError(t, err)
}

func TestCipherFTS5Insert(t *testing.T) {
	db := getDB()
	defer db.Close()

	// 写入数据
	userList := []User{
		{
			Name:        "小明",
			Description: "今天天气很不错",
		},
		{
			Name:        "小红",
			Description: "明天就放假了",
		},
		{
			Name:        "小黑",
			Description: "今年去哪玩?",
		},
		{
			Name:        "小白",
			Description: "今天下雨了",
		},
		{
			Name:        "小强",
			Description: "今天去哪里玩?",
		},
	}

	//file, err := os.Open("G:\\Environments\\area_code_2024.csv\\area_code_2024.csv")
	//require.NoError(t, err)
	//defer file.Close()
	//
	//// 创建一个新的 Scanner
	//scanner := bufio.NewScanner(file)

	//var userList []User

	// 按行读取文件内容
	//for scanner.Scan() {
	//	line := scanner.Text()
	//	lines := strings.Split(line, ",")
	//	userList = append(userList, User{Name: lines[0], Description: lines[1]})
	//}

	fmt.Println("count ", len(userList))

	for _, user := range userList {
		_, err := db.Exec(`INSERT INTO user(name, description) VALUES(?,?)`, user.Name, user.Description)
		require.NoError(t, err)
	}
}

func TestCipherFTS5Search(t *testing.T) {
	db := getDB()
	defer db.Close()

	startTime := time.Now()
	rows, err := db.Query(`
	SELECT rowid, name, description 
	FROM user_fts 
	WHERE user_fts MATCH simple_query('beicun')
	ORDER BY rank
	LIMIT 500;
	`)
	require.NoError(t, err)
	defer rows.Close()
	fmt.Println(time.Since(startTime).Milliseconds())
	var user User

	for rows.Next() {
		err := rows.Scan(&user.Id, &user.Name, &user.Description)
		require.NoError(t, err)
		fmt.Println(user)
	}
}
