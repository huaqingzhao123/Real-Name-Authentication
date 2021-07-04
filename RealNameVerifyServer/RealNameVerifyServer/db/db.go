package db

import (
	"RealNameVerifyServer/Config"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

var (
	pool                 *redis.Pool
	UserUpdateVerifyChan = make(chan *Config.UserAntiAddiction, 50000)
)

func InitDB() {
	pool = Config.GetRedisPool()
}

///从数据库查询账号是否存在
func IsAlreadyVerify(gameSign,account string) int32 {
	verifyHas := Config.GetStmt("VerifyReallyHas")
	if verifyHas == nil {
		logs.Error("mysql stmt出错， stmt并未提前生成")
		return 0
	}
	uID := fmt.Sprintf("%s-%s", gameSign,account)
	var num int32
	logs.Info("进入查询玩家信息是否存在的存储过程(db.31)")
	row := verifyHas.QueryRow(uID)
	if row!=nil{
		err:=row.Scan(&num)
		if err != nil {
			logs.Error("数据库操作出错%v", err.Error())
			return num
		}
	}
	logs.Info("退出查询玩家信息是否存在的存储过程(db.33)")
	return num
}

//查询玩家请求的实名信息是否已经被别人注册
func VerifyMesRepeat(gameSign string,cardId string)int32{
	//先在服务器拿一下
	isHave:=Config.JudgeUserAlreadyRegister(gameSign,cardId)
	if isHave{
		return 1
	}
	verifyHas := Config.GetStmt("VerifyRepeat")
	if verifyHas == nil {
		logs.Error("mysql stmt出错， stmt并未提前生成")
		return 0
	}
	var num int32
	err := verifyHas.QueryRow(gameSign,cardId).Scan(&num)
	if err != nil {
		logs.Error("数据库操作出错%v", err.Error())
		return num
	}
	return num
}

//根据基础时间和差值为实名认证系统设置登陆或者退出时间
func SetLoginOrExitTimeForVerify(baseHour,baseMinute, offsetMinute int) (lastExitTime string) {
	var nowHour = baseHour
	var nowMinute = baseMinute
	nowMinute+=offsetMinute
	if nowMinute>=60{
		nowHour+=nowMinute/60
		nowMinute=nowMinute%60
	}
	var newHour string
	var newMinute string
	if nowHour < 10 {
		newHour = fmt.Sprintf("0%d", nowHour)
	} else {
		newHour = fmt.Sprintf("%d", nowHour)
	}
	if nowMinute < 10 {
		newMinute = fmt.Sprintf("0%d", nowMinute)
	} else {
		newMinute = fmt.Sprintf("%d", nowMinute)
	}
	lastExitTime = fmt.Sprintf("%v%v", newHour, newMinute)
	return
}
//更新玩家的退出时间
func SetExitTimeWithNum(lastExitTime string,offsetMinute int)string{
	var exitHour = lastExitTime[0:2]
	var exitMinute = lastExitTime[2:4]
	exitHourInt, err := strconv.Atoi(exitHour)
	if err != nil {
		logs.Error("string转换Int类型失败")
	}
	exitMinuteInt, err := strconv.Atoi(exitMinute)
	if err != nil {
		logs.Error("string转换Int类型失败")
	}
	return SetLoginOrExitTimeForVerify(exitHourInt,exitMinuteInt,offsetMinute)
}
///将玩家实名信息存进redis
func storeUserMesToRedis(gameSign string, mes *Config.UserAntiAddiction) {
	var redisLua = Config.GetRedisLua(Config.SetVerifyMes)
	if redisLua == nil {
		logs.Error("redis未提前生成")
		return
	}
	con := pool.Get()
	defer con.Close()
	redisStr := fmt.Sprintf("%s_%s", gameSign, mes.Acc)
	_, err := redisLua.Do(con, redisStr, mes.Acc, mes.LastLoginTime, mes.LastExitTime, mes.CardID, mes.CanGoOn, mes.Age,mes.ResiduePurchase)
	logs.Info("存进redis的信息为%v%v%v%v%v%v%v",mes.Acc, mes.LastLoginTime, mes.LastExitTime, mes.CardID, mes.CanGoOn, mes.Age,mes.ResiduePurchase)
	logs.Info("redis写入%v玩家信息",redisStr)
	if err != nil {
		logs.Error("redis写入游戏:%s玩家:%s防沉迷信息出错%v", gameSign, mes.Acc, err)
	}
}

//更新redis和服务器缓存的玩家信息，防止覆盖redis导致表的生存时间被修改
func UpdateUserRedisMes(gameSign string, mes *Config.UserAntiAddiction,needSet bool) {
	keyNum := CheckUserVerifyMesExistRedis(gameSign, mes.Acc)
	//redis存在该玩家信息，调用更新方法
	if keyNum > 0 {
		var redisLua = Config.GetRedisLua(Config.UpdateVerifyMes)
		if redisLua == nil {
			logs.Error("redis未提前生成")
			return
		}
		con := pool.Get()
		defer con.Close()
		redisStr := fmt.Sprintf("%s_%s", gameSign, mes.Acc)
		_, err := redisLua.Do(con, redisStr, mes.Acc, mes.LastLoginTime, mes.LastExitTime, mes.CardID, mes.CanGoOn, mes.Age,mes.ResiduePurchase)
		logs.Info("存进redis的信息为%v%v%v%v%v%v%v",mes.Acc,mes.LastLoginTime, mes.LastExitTime, mes.CardID, mes.CanGoOn, mes.Age,mes.ResiduePurchase)
		if err != nil {
			logs.Error("redis写入游戏:%s玩家:%s防沉迷信息出错:%v", gameSign, mes.Acc, err)
		}
	} else {
		storeUserMesToRedis(gameSign, mes)
	}
	//只管理未成年,更新服务器上的信息
	if mes.Acc != "" && mes.Age<18&&needSet {
		//将数据缓存到服务器
		Config.UpdateUserVerifyMes(gameSign, mes.Minute, mes.Age, mes.CanGoOn,mes.ResiduePurchase, mes.CardID, mes.Acc, mes.LastLoginTime, mes.LastExitTime)
	}
}

//得到玩家实名认证信息
func GetUserVerifyMes(gameSign, account string) (mes *Config.UserAntiAddiction) {
	//先从服务器缓存拿
	mes = Config.GetUserVerifyMes(gameSign, account)
	logs.Info("服务器中此玩家的信息为:%v",mes)
	//服务器没有此玩家信息
	if mes == nil {
		//从redis拿
		mes = GetUserVerifyMesFromRedis(gameSign, account)
		logs.Info("redis中此玩家的信息为:%v",mes)
		//从数据库拿，再缓存到redis
		if mes == nil {
			num := IsAlreadyVerify(gameSign, account)
			logs.Info("数据库中此玩家的数量为:%d(db.152)",num)
			if num == 0 {
				logs.Error("sql找不到游戏:%s的玩家:%v的实名信息", gameSign, account)
				return nil
			} else {
				verifyMesGet := Config.GetStmt("VerifyRepeatIdty")
				if verifyMesGet == nil {
					logs.Error("mysql stmt出错， stmt并未提前生成")
					return nil
				}
				uID := fmt.Sprintf("%s-%s", gameSign, account)
				rows, sqlErr := verifyMesGet.Query(uID)
				if sqlErr != nil {
					logs.Error("数据库读取出错,可能没有该玩家的信息:%v", sqlErr.Error())
					return nil
				}
				var acc string
				var lastLoginTime string
				var lastExitTime string
				var id string
				var canGoOn int32
				var age int32
				var residuePurchase int32
				//var isReallyVerify int32
				mes = new(Config.UserAntiAddiction)
				for rows.Next() {
					if err := rows.Scan(&acc, &lastLoginTime, &lastExitTime, &id,
						&canGoOn, &age,&residuePurchase); err == nil {
						var strs=strings.Split(acc,"-")
						mes.GameSign = strs[0]
						mes.Acc=strs[1]
						mes.LastLoginTime = lastLoginTime
						mes.LastExitTime = lastExitTime
						mes.CardID = id
						mes.CanGoOn = canGoOn
						mes.Age = age
						mes.ResiduePurchase=residuePurchase
						logs.Info("即将返回的用户实名信息为:%v", mes)
					}
				}
				if mes.Acc != "" {
					//此时redis没有玩家信息,进行缓存
					mes.Minute,_=CalculateUserGameTime(mes.LastLoginTime,mes.LastExitTime)
					UpdateUserRedisMes(mes.GameSign, mes,true)
					return
				} else {
					logs.Error("找不到游戏:%s的玩家:%v的实名信息", gameSign, account)
					return nil
				}
			}
		}
		//只管理未成年,缓存到服务器
		if mes.Acc != "" && mes.Age<18 {
			//将数据缓存到服务器
			mes.Minute,_=CalculateUserGameTime(mes.LastLoginTime,mes.LastExitTime)
			Config.UpdateUserVerifyMes(gameSign, mes.Minute, mes.Age, mes.CanGoOn,mes.ResiduePurchase, mes.CardID, account, mes.LastLoginTime, mes.LastExitTime)
		}
	}
	return
}

//检测redis是否存在玩家的实名信息
func CheckUserVerifyMesExistRedis(gameSign, account string) int32 {
	var redisLua = Config.GetRedisLua(Config.CheckVerifyMes)
	if redisLua == nil {
		logs.Error("redis未提前生成")
		return 0
	}
	con := pool.Get()
	defer con.Close()
	redisStr := fmt.Sprintf("%s_%s", gameSign, account)
	keyNum, err := redis.Int(redisLua.Do(con, redisStr))
	if err != nil {
		logs.Error("redis读取游戏:%s玩家:%s防沉迷信息出错:%v", gameSign, account, err)
	}
	return int32(keyNum)
}

//从redis得到玩家实名信息
func GetUserVerifyMesFromRedis(gameSign, account string) (mes *Config.UserAntiAddiction) {
	mes = new(Config.UserAntiAddiction)
	keyNum := CheckUserVerifyMesExistRedis(gameSign, account)
	if keyNum > 0 {
		var redisLua = Config.GetRedisLua(Config.GetVerifyMes)
		if redisLua == nil {
			logs.Error("redis未提前生成")
			return
		}
		con := pool.Get()
		defer con.Close()
		redisStr := fmt.Sprintf("%s_%s", gameSign, account)
		mesInterfaces, err := redis.Values(redisLua.Do(con, redisStr))
		if err != nil {
			logs.Error("redis读取游戏:%s玩家:%s防沉迷信息出错%v", gameSign, account, err)
		} else {
			if mes.Acc, err = redis.String(mesInterfaces[0], err); err != nil {
				logs.Error("redis Acc转换游戏:%s玩家:%s防沉迷信息出错%v", gameSign, account, err)
				return nil
			}
			if mes.LastLoginTime, err = redis.String(mesInterfaces[1], err); err != nil {
				logs.Error("redis LastLoginTime转换游戏:%s玩家:%s防沉迷信息出错%v", gameSign, account, err)
				return nil
			}
			if mes.LastExitTime, err = redis.String(mesInterfaces[2], err); err != nil {
				logs.Error("redis LastExitTime转换游戏转换游戏:%s玩家:%s防沉迷信息出错%v", gameSign, account, err)
				return nil
			}
			if mes.CardID, err = redis.String(mesInterfaces[3], err); err != nil {
				logs.Error("redis CardID转换游戏:%s玩家:%s防沉迷信息出错", gameSign, account)
				return nil
			}
			var canGoOn, age,redisPurchase int
			if canGoOn, err = redis.Int(mesInterfaces[4], err); err != nil {
				logs.Error("redis canGoOn转换游戏:%s玩家:%s防沉迷信息出错%v", gameSign, account,err)
				return nil
			}
			if age, err = redis.Int(mesInterfaces[5], err); err != nil {
				logs.Error("redis age转换游戏:%s玩家:%s防沉迷信息出错", gameSign, account)
				return nil
			}

			if redisPurchase, err = redis.Int(mesInterfaces[6], err); err != nil {
				logs.Error("redis redisPurchase转换游戏:%s玩家:%s防沉迷信息出错", gameSign, account)
				return nil
			}
			mes.CanGoOn = int32(canGoOn)
			mes.Age = int32(age)
			mes.ResiduePurchase =int32(redisPurchase)
			return
		}
	} else {
		return nil
	}
	return
}

//更新sql玩家的数据信息
func HandleSqlUpdateUserVerify() {
	for {
		select {
		case req := <-UserUpdateVerifyChan:
			stVerify := Config.GetStmt("VerificationUpdate")
			if stVerify == nil {
				logs.Error("mysql stmt出错， stmt并未提前生成")
				return
			}
			//只有数据库存储的account为gameSign_acc
			var acc = fmt.Sprintf("%v-%v",req.GameSign,req.Acc)
			_, sqlErr := stVerify.Query(acc,req.LastLoginTime, req.LastExitTime, req.CardID, req.CanGoOn,
				req.Age,req.ResiduePurchase)
			//部分玩家redis数据还未消失，更新一下
			logs.Info("要存进数据库的信息为:%v",req)
			UpdateUserRedisMes(req.GameSign,req,false)
			if sqlErr != nil {
				logs.Error("玩家实名认证信息写入数据库出错%v", sqlErr.Error())
			}
			Config.DeleteUserVerifyMes(req.GameSign,req.Acc)
		}
	}
}

//每个月重置未成年玩家的消费金额
func HandleResetSqlPurchaseMes(){
	var t *time.Ticker
	for {
		//每个月重置未成年玩家的消费金额
		now := time.Now()
		workStart := time.Date(now.Year(), now.Month(), 1, 3, 0, 0, 0, time.Local).AddDate(0,1,0)
		///周期性定时，每三小时执行一次
		t = time.NewTicker(workStart.Sub(now))
		<-t.C
		logs.Info("更新sql玩家的消费金额")
		//更新玩家的实名认证信息
		sql:=Config.GetStmt("UpdatePurchaseNum")
		_, sqlErr := sql.Exec()
		if sqlErr!=nil{
			logs.Error("更新sql玩家消费金额出错：%v",sqlErr)
		}
	}
}
//获得未成年玩家剩余的游戏时间和已经游戏的时间,第一個返回值為剩餘時間,第二个已经游戏时间
func CalculateUserGameTime(lastloginTime, lastExitTime string) (int32,int32) {
	var now=time.Now()
	//22点到8点不提供游戏功能
	//if now.Hour()>21||now.Hour()<8{
	//	return  0,0
	//}
	if lastloginTime == "" || lastExitTime == "" {
		logs.Info("没有登陆时间或者退出时间")
		return 0,0
	} else {
		var loginHour = lastloginTime[0:2]
		loginHourInt, err := strconv.Atoi(loginHour)
		if err != nil {
			logs.Error("string转换Int类型失败")
		}
		var loginMinute = lastloginTime[2:4]
		loginMinuteInt, err := strconv.Atoi(loginMinute)
		if err != nil {
			logs.Error("string转换Int类型失败")
		}
		var exitHour = lastExitTime[0:2]
		var exitMinute = lastExitTime[2:4]
		exitHourInt, err := strconv.Atoi(exitHour)
		if err != nil {
			logs.Error("string转换Int类型失败")
		}
		exitMinuteInt, err := strconv.Atoi(exitMinute)
		if err != nil {
			logs.Error("string转换Int类型失败")
		}
		var minute int
		minute = exitMinuteInt - loginMinuteInt
		hour := exitHourInt - loginHourInt
		if hour > 0 && minute < 0 {
			minute += 60
			hour -= 1
		}
		//玩到第二天
		if hour < 0 {
			hour += 24
			if minute < 0 {
				minute += 60
				hour -= 1
			}
		}
		minute += hour * 60
		//根据节假日判断剩余的游戏时间
		//var nowWeekday=now.Weekday()
		var month=int32(now.Month())
		var dayStr,monthStr string
		if month<10{
			monthStr=fmt.Sprintf("0%v",month)
		}else {
			monthStr=string(month)
		}
		if now.Day()<10{
			dayStr=fmt.Sprintf("0%v",now.Day())
		}else {
			dayStr=string(now.Day())
		}
		var nowDay=fmt.Sprintf("%s%s",monthStr,dayStr)
		logs.Info("登陆时间%v，退出时间%v,minute:%v",lastloginTime,lastExitTime,minute)
		//节假日
		if Config.GetOneDayIsHoliday(nowDay){
			return  int32(180-minute),int32(minute)
		}
		//其它时间一天1.5小时
		return  int32(90-minute),int32(minute)
	}
}
