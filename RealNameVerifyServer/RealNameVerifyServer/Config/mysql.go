package Config

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	//注册mysql 的 驱动
	_ "github.com/go-sql-driver/mysql"
)

// defaultMysqlConfig mysql 配置
type defaultMysqlConfig struct {
	URL               string
	Enable            bool
	MaxIdleConnection int
	MaxOpenConnection int
}

var (
	dbInited  bool
	mdb       sync.Mutex
	defaultdb = &defaultMysqlConfig{
		Enable: false,
		//URL需要拼接，在init里面进行拼接
		//URL:               "root:Wingjoy1234@tcp(mysqlserver:3306)/ghostswitch",
		MaxIdleConnection: 5,
		MaxOpenConnection: 50,
	}
	mysqlDB *sql.DB

	accountSQLStr = map[string]string{
		"VerifyRepeatIdty":   `call  verifyRepeatIdty(?)`,
		"VerificationUpdate": `call  verificationUpdate(?,?,?,?,?,?,?)`,
		"VerifyReallyHas":    `call  verifyReallyHas(?)`,
		"VerifyRepeat":       `call verifyRepeat(?,?)`,
		"UpdatePurchaseNum":  `call updatePurchaseNum()`,
	}
	st = map[string]*sql.Stmt{}
)

func GetDefaultMysqlConfig() *defaultMysqlConfig {
	return defaultdb
}

// GetDB 获取db
func GetDB() *sql.DB {
	return mysqlDB
}

func GetStmt(stName string) *sql.Stmt {
	return st[stName]
}

func initMysql(forceEnable bool) error {
	mdb.Lock()
	defer mdb.Unlock()
	var err error
	if dbInited {
		err = fmt.Errorf("mysql模块已经初始化过")
		fmt.Println(err.Error())
		return err
	}

	defaultdb.Enable = forceEnable
	defaultdb.URL = fmt.Sprintf("root:Wingjoy1234@tcp(realname-mysql:3306)/Verification?collation=utf8mb4_unicode_ci&charset=utf8mb4")
	// 如果配置声明使用log
	if defaultdb.Enable {
		err = initMysqlInternal()
		if err != nil {
			fmt.Println("初始化mysql失败%v", err.Error())
		}
	}

	dbInited = true
	return err
}

func initMysqlInternal() (err error) {
	fmt.Println("初始化mysql...")
	// 创建连接
	mysqlDB, err = sql.Open("mysql", defaultdb.URL)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// 最大连接数
	mysqlDB.SetMaxOpenConns(defaultdb.MaxOpenConnection)

	// 最大闲置数
	mysqlDB.SetMaxIdleConns(defaultdb.MaxIdleConnection)

	//最大闲置时间
	mysqlDB.SetConnMaxLifetime(time.Minute)

	//初始化stmt
	for query, statement := range accountSQLStr {
		prepared, errMysql := mysqlDB.Prepare(statement)
		//checkError(err, "stmt生成失败")
		if errMysql != nil {
			return errMysql
		}
		st[query] = prepared
	}

	// 激活链接
	if err = mysqlDB.Ping(); err != nil {
		fmt.Println(err)
		return
	}
	return
}
