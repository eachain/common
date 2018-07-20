// mysql 实现有依赖其它包，项目中用时请自行替换。

package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/eachain/common/logger"
	"github.com/eachain/common/utils"
)

const mysqlBiz = "__mysql__" // 用于 utils.Go

const (
	// CheckDBInterval 是检查 DB 可用性的时间间隔。
	checkDBInterval = 3 * time.Second

	// DefaultLifeTime 是默认的wait_timeout秒数。
	defaultLifeTime = 10
)

var BizNotFound = errors.New("Biz not found")

// ConfigInfo 包含mysql初始化信息。
type ConfigInfo struct {
	// BizName 是业务名，用于区分不同的业务或连接。
	BizName string
	// DSN 格式：[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// DSN 示例：username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8&timeout=5s&readTimeout=5s
	// &writeTimeout=5s&parseTime=true&&loc=Local
	DSN string
	// MaxOpenConns 是最大打开连接数。
	MaxOpenConns int
	// MaxOpenConns 是最大空闲连接数。
	MaxIdleConns int
	// LifeTime 是SetConnMaxLifetime参数的值。
	LifeTime time.Duration
}

type dbclient struct {
	db   *sql.DB
	info ConfigInfo
}

// clients 只在初始化的时候变更
// TODO: 要变更配置，需要加锁
var clients = make(map[string]*dbclient)

// Init 初始化并注册所有需要的sql.DB。
func Init(infos []ConfigInfo) {
	for _, info := range infos {
		db, err := dial(info)
		if err != nil {
			panic(info.BizName + " Open:" + err.Error())
		}

		cli := &dbclient{
			db:   db,
			info: info,
		}
		utils.Go(mysqlBiz, cli.refresh)
		clients[info.BizName] = cli
	}
}

func Get(biz string) (*sql.DB, error) {
	cli := clients[biz]
	if cli == nil {
		return nil, BizNotFound
	}

	return cli.db, nil
}

func dial(info ConfigInfo) (*sql.DB, error) {
	db, err := sql.Open("mysql", info.DSN)
	if err != nil {
		return nil, err
	}
	if info.MaxOpenConns != 0 {
		db.SetMaxOpenConns(info.MaxOpenConns)
	}
	if info.MaxIdleConns != 0 {
		db.SetMaxIdleConns(info.MaxIdleConns)
	}
	db.SetConnMaxLifetime(info.LifeTime)
	return db, nil
}

func (cli *dbclient) refresh(ctx context.Context) {
	ping := time.NewTicker(checkDBInterval)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ping.C:
			err := cli.db.PingContext(ctx)
			if err != nil {
				logger.Warnf("mysql: ping: %v", err)
			}
		}
	}
}

// CloseAll 应该在主线程中被调用，比如进程退出前。
func CloseAll() {
	utils.StopAndWait(mysqlBiz)

	for _, cli := range clients {
		cli.db.Close()
	}
}
