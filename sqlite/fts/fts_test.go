package test

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"testing"
)

const TableSQL = `CREATE TABLE IF NOT EXISTS chat (_id INTEGER PRIMARY KEY AUTOINCREMENT,sender_nickname TEXT,data BLOB)`
const SearchTableSQL = `CREATE VIRTUAL TABLE IF NOT EXISTS chat_fts USING fts5(sender_nickname, data, content=chat, content_rowid=_id, tokenize='simple')`

const TriggerSQL = `
CREATE TRIGGER IF NOT EXISTS chat_fts_i AFTER INSERT ON chat BEGIN
INSERT INTO chat_fts(rowid, sender_nickname, data) VALUES (new._id, new.sender_nickname, new.data);
END;
CREATE TRIGGER IF NOT EXISTS chat_fts_d AFTER DELETE ON chat BEGIN
INSERT INTO chat_fts(chat_fts, rowid, sender_nickname, data) VALUES('delete', old._id, old.sender_nickname, old.data);
END;
CREATE TRIGGER IF NOT EXISTS chat_fts_u AFTER UPDATE ON chat BEGIN
INSERT INTO chat_fts(chat_fts, rowid, sender_nickname, data) VALUES('delete', old._id, old.sender_nickname, old.data);
INSERT INTO chat_fts(rowid, sender_nickname, data) VALUES (new._id, new.sender_nickname, new.data);
END;
`
const InsertSQL = `INSERT INTO chat(sender_nickname, data) VALUES(?, ?);`
const SearchSQL = `SELECT rowid,sender_nickname,data FROM chat_fts where chat_fts match ? ORDER BY rank;`

var dataList = []string{
	"不想上学的举高手", "[emoji:255]\x1a\x0b\x00",
	"Dianjixz", "/笑哭/笑哭",
	"Drifter", "rv64是啥",
	"哦仙人", "等等，现在有啥可以用的RV64？",
	"L0/1/2_泽畔无材", "[QQ红包]",
	"ㅤ", "/魔鬼笑/魔鬼笑/魔鬼笑",
	"qq@Αρηδ", "64位ri scv",
	"名字不重要", "Risc-v64",
	"L0/1/2_泽畔无材", "可以跑linux的",
	"QiqiStudio", "[图片]",
	"。。。。。", "难道是k510",
	"哦仙人", "最近有出这种板卡么？",
	"\u202d\u202d\u202d", "完全没听过 你们太厉害了",
	"林夕木易", "[图片]",
	"林夕木易", "正在编译",
	"qq@Αρηδ", "riscv的单片机 性能很垃圾。高端的不知道性能如何。",
}

func TestFTS5(t *testing.T) {
	sql.Register("sqlite3_simple",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"G:\\Solutions\\ocean\\sqlite\\fts\\simple",
			},
		})

	db, err := sql.Open("sqlite3_simple", "fts_test")
	require.NoError(t, err)
	defer db.Close()

	// 建表
	_, err = db.Exec(TableSQL)
	require.NoError(t, err)
	// 索引表
	_, err = db.Exec(SearchTableSQL)
	require.NoError(t, err)
	// 触发器
	_, err = db.Exec(TriggerSQL)
	require.NoError(t, err)

	for i := 0; i < len(dataList); i += 2 {
		db.Exec(InsertSQL, dataList[i], dataList[i+1])
	}

	//rows, err := db.Query(SearchSQL, "不")
	rows, err := db.Query(`SELECT rowid, sender_nickname, data FROM chat_fts where chat_fts match simple_query('人') ORDER BY rank`)
	require.NoError(t, err)

	defer rows.Close()
	for rows.Next() {
		var rowid int64
		var senderNickname string
		var data string
		rows.Scan(&rowid, &senderNickname, &data)
		fmt.Println(rowid, senderNickname, data)
	}

}
