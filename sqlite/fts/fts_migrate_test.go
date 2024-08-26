package test

import (
	"database/sql"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"testing"
	"time"
)

type SyncDepartmentInfo struct {
	Id             int64 `gorm:"primaryKey"`
	RecordId       string
	GaDepartment   string `gorm:"uniqueIndex"` //唯一
	Alias          string
	ParentResource string
	SortNo         string
	DeleteStatus   int
}

type SyncCaseTypeInfo struct {
	Id             int64 `gorm:"primaryKey"`
	RecordId       string
	GaCaseType     string `gorm:"uniqueIndex"`
	CaseTypeName   string
	BusinessClass  string
	ParentResource string
	SortNo         string
	DeleteStatus   int
}

func TestFTS5Simple(t *testing.T) {
	driverName := "sqlite3_simple"
	extensions := []string{
		"xxx\\simple.dll",
	}
	RegisterDriver(driverName, extensions)
	t.Log("注册驱动成功")

	//fromDSN := "xxx"
	toDSN := "xxx?_cipher=chacha20&_key=xxx"

	//fromDB, err := ConnectDB(driverName, fromDSN)
	//require.NoError(t, err)

	toDB, err := ConnectDB(driverName, toDSN)
	require.NoError(t, err)
	t.Log("连接数据库成功")

	MigrateSchema(t, toDB)
	t.Log("建表成功")

	CreateVirtualTable(t, toDB)
	t.Log("创建虚拟表成功")

	CreateTrigger(t, toDB)
	t.Log("创建触发器成功")

	//MigrateData(t, fromDB, toDB)
	//QueryData(t, toDB)
	//QueryDepartmentData(t, toDB)
	//QueryCaseData(t, toDB)
}

func MigrateData(t *testing.T, fromDB, toDB *gorm.DB) {
	t.Log("开始迁移数据")
	start := time.Now()
	defer func() {
		t.Log("迁移数据结束 ", time.Since(start).Milliseconds())
	}()

	var syncDepartmentInfoList []*SyncDepartmentInfo
	err := fromDB.Find(&syncDepartmentInfoList).Error
	require.NoError(t, err)

	err = toDB.Create(syncDepartmentInfoList).Error
	require.NoError(t, err)

	var syncCaseTypeInfoList []*SyncCaseTypeInfo
	err = fromDB.Find(&syncCaseTypeInfoList).Error
	require.NoError(t, err)

	err = toDB.Create(syncCaseTypeInfoList).Error
	require.NoError(t, err)
}

func MigrateSchema(t *testing.T, db *gorm.DB) {
	err := db.AutoMigrate(&SyncDepartmentInfo{}, &SyncCaseTypeInfo{})
	require.NoError(t, err)
}

func RegisterDriver(driverName string, extensions []string) {
	sql.Register(
		driverName,
		&sqlite3.SQLiteDriver{
			Extensions: extensions,
		},
	)
}

func ConnectDB(driverName, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Dialector{
		DriverName: driverName,
		DSN:        dsn,
	}, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		CreateBatchSize: 512,
	})
	return db, err

}

func CreateVirtualTable(t *testing.T, db *gorm.DB) {
	// 索引表
	err := db.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS sync_department_info_fts USING fts5(
	       alias, 
	       ga_department, 
	       content=sync_department_info,
	       content_rowid=id, 
	       tokenize="simple"
    )
    `).Error
	require.NoError(t, err)

	err = db.Exec(`
	CREATE VIRTUAL TABLE IF NOT EXISTS sync_case_type_info_fts USING fts5(
	       ga_case_type, 
	       case_type_name, 
           record_id UNINDEXED,
           parent_resource UNINDEXED,
	       content=sync_case_type_info,
	       content_rowid=id, 
	       tokenize="simple"
    )
    `).Error
	require.NoError(t, err)
}

func CreateTrigger(t *testing.T, db *gorm.DB) {
	_sql := `
    CREATE TRIGGER IF NOT EXISTS sync_department_info_fts_i AFTER INSERT ON sync_department_info BEGIN
	INSERT INTO sync_department_info_fts(rowid, alias, ga_department) 
	VALUES (new.id, new.alias, new.ga_department);
	END;
	
	CREATE TRIGGER IF NOT EXISTS sync_department_info_fts_d AFTER DELETE ON sync_department_info BEGIN
	INSERT INTO sync_department_info_fts(sync_department_info_fts, rowid, alias, ga_department) 
	VALUES('delete', old.id, old.alias, old.ga_department);
	END;
	
	CREATE TRIGGER IF NOT EXISTS sync_department_info_fts_u AFTER UPDATE ON sync_department_info BEGIN
	INSERT INTO sync_department_info_fts(sync_department_info_fts, rowid, alias, ga_department) 
	VALUES('delete', old.id, old.alias, old.ga_department);

	INSERT INTO sync_department_info_fts(rowid, alias, ga_department) 
	VALUES (new.id, new.alias, new.ga_department);
	END;
    `
	err := db.Exec(_sql).Error
	require.NoError(t, err)

	_sql = `
    CREATE TRIGGER IF NOT EXISTS sync_case_type_info_fts_i AFTER INSERT ON sync_case_type_info BEGIN
	INSERT INTO sync_case_type_info_fts(rowid, ga_case_type, case_type_name, record_id, parent_resource) 
	VALUES (new.id, new.ga_case_type, new.case_type_name, new.record_id, new.parent_resource);
	END;
	
	CREATE TRIGGER IF NOT EXISTS sync_case_type_info_fts_d AFTER DELETE ON sync_case_type_info BEGIN
	INSERT INTO sync_case_type_info_fts(sync_case_type_info_fts, rowid, ga_case_type, case_type_name, record_id, parent_resource) 
	VALUES('delete', old.id, old.ga_case_type, old.case_type_name, old.record_id, old.parent_resource);
	END;
	
	CREATE TRIGGER IF NOT EXISTS sync_case_type_info_fts_u AFTER UPDATE ON sync_case_type_info BEGIN
	INSERT INTO sync_case_type_info_fts(sync_case_type_info_fts, rowid, ga_case_type, case_type_name, record_id, parent_resource) 
	VALUES('delete', old.id, old.ga_case_type, old.case_type_name, old.record_id, old.parent_resource);

	INSERT INTO sync_case_type_info_fts(rowid, ga_case_type, case_type_name, record_id, parent_resource) 
	VALUES (new.id, new.ga_case_type, new.case_type_name, new.record_id, new.parent_resource);
	END;
    `
	err = db.Exec(_sql).Error
	require.NoError(t, err)
}

func QueryDepartmentData(t *testing.T, db *gorm.DB) {
	//var s string
	// "今" AND "天" AND "天" AND "气" AND "真" AND "不" AND "错"
	//db.Raw("select simple_query('今天天气真不错')").Scan(&s)
	//fmt.Println(s)
	var sdis []SyncDepartmentInfo
	db = db.Debug()
	err := db.Raw(`
	SELECT rowid as id, ga_department, alias 
	FROM sync_department_info_fts 
	WHERE sync_department_info_fts MATCH simple_query("xxx")
	ORDER BY rank
	Limit ?
	`, 10).Scan(&sdis).Error
	require.NoError(t, err)
	for _, sdi := range sdis {
		t.Log(sdi)
	}
}

func QueryCaseData(t *testing.T, db *gorm.DB) {
	var s string
	//"今" AND "天" AND "天" AND "气" AND "真" AND "不" AND "错"
	db.Raw(`select simple_query("xxx")`).Scan(&s)
	fmt.Println(s)
	var sdis []SyncCaseTypeInfo
	db = db.Debug()
	err := db.Raw(`
    SELECT rowid as id, ga_case_type, case_type_name, record_id, parent_resource
    FROM sync_case_type_info_fts
    WHERE sync_case_type_info_fts MATCH simple_query("xxx")
    ORDER BY rank
    Limit ?
    `, 10).Scan(&sdis).Error
	require.NoError(t, err)
	fmt.Println(sdis)
}
