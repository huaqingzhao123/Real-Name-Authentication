package Config

import (
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

// defaultRedisConfig redis 配置
type defaultRedisConfig struct {
	Enabled  bool
	Conn     string
	Password string
}

type RedisLuaEnum int

const (
	GetVerifyMes    RedisLuaEnum = 53 //得到用户的认证信息
	SetVerifyMes    RedisLuaEnum = 54 //设置用户的认证信息
	CheckVerifyMes  RedisLuaEnum = 55 //检查用户的认证信息
	UpdateVerifyMes RedisLuaEnum = 56 //更新用户的认证信息

	CacheTime = 172800 //缓存2天
)

var (
	clientPool   *redis.Pool
	mredis       sync.RWMutex
	redisInited  bool
	defaultRedis = &defaultRedisConfig{
		Enabled: false,
		//初始化时拼接 ，格式 helm-rediscache
		// Conn:     "redisserver:6379/0",
		Password: "basketballPWD",
	}
	registeredLuaScripts map[RedisLuaEnum]*redis.Script
)

func GetDefaultRedisConfig() *defaultRedisConfig {
	return defaultRedis
}

func initRedis(forceEnable bool) error {
	mredis.Lock()
	defer mredis.Unlock()
	var err error
	if redisInited {
		err = fmt.Errorf("redis模块已经初始化过")
		fmt.Println(err.Error())
		return err
	}
	defaultRedis.Conn = "realname-redis:6379/0" //"localhost:6379"//fmt.Sprintf("localhost:6379/0", HelmReleaseName)
	defaultRedis.Enabled = forceEnable
	// 如果配置声明使用log
	if defaultRedis.Enabled {
		initRedisInternal()
	}

	redisInited = true
	return nil
}

func initRedisInternal() {
	// 打开才加载
	redisURL := fmt.Sprintf("redis://:%s@%s", defaultRedis.Password, defaultRedis.Conn)
	fmt.Println("初始化Redis...")

	clientPool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   5000, //redis默认支持10000
		IdleTimeout: 300 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {

			c, err := redis.DialURL(redisURL)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	//注册redis的map，缓存一天
	registeredLuaScripts = map[RedisLuaEnum]*redis.Script{
		SetVerifyMes: redis.NewScript(0,
			`redis.call("hmset", ARGV[1], "account",ARGV[2],"lastLgTime", ARGV[3],"lastExtTime",ARGV[4],
			"userId",ARGV[5],"CanGoOn",ARGV[6],"age",ARGV[7],"residuePurchase",ARGV[8])
			redis.call("expire",ARGV[1],86410) 
			`), //,"IsReallyVerify",ARGV[8]
		GetVerifyMes: redis.NewScript(0, `
			return redis.call("HMGET",ARGV[1],"account","lastLgTime","lastExtTime","userId","CanGoOn","age","residuePurchase")
			`),
		CheckVerifyMes: redis.NewScript(0, `
			return redis.call("hlen",ARGV[1])`),
		UpdateVerifyMes: redis.NewScript(0, `
		redis.call("hmset", ARGV[1], "account",ARGV[2],"lastLgTime", ARGV[3],"lastExtTime",ARGV[4],
		"userId",ARGV[5],"CanGoOn",ARGV[6],"age",ARGV[7],"residuePurchase",ARGV[8])
		`),
	}
}

// GetRedis 获取redis
func GetRedisPool() *redis.Pool {
	return clientPool
}

func GetRedisLua(luaType RedisLuaEnum) *redis.Script {
	if retScript, bExist := registeredLuaScripts[luaType]; bExist {
		return retScript
	}
	return nil
}
