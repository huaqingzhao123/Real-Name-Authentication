package main

import (
	"RealNameVerifyServer/Config"
	"RealNameVerifyServer/db"
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils"
)

//玩家发送的实名认证信息
type VerifyMes struct {
	Accout string `json:"account"` //玩家在游戏内的账号
	UserId string `json:"userId"`  //身份证号
	//Name   		string 		`json:"name"`     		//姓名 -----暂时不需要
	GameSign   string `json:"gameSign"`   //游戏标示
	ConsumeNum int32  `json:"consumeNum"` //玩家消费的额度，规整为整数传递
	Times      int32  `json:"times"`      //登陆时第一次发送统计登陆时间
	Interval  int32   `json:"interval"`      //请求的时间间隔
}

/*type AddictionVerify struct {
	Accout   string 		`json:"account"` 	 	//玩家在游戏内的账号
	GameSign string 		`json:"gameSign"` 		//游戏标示
	GameExitTime string 	`json:"gameExitTime"`	//玩家的退出时间，请求玩家防沉迷状态时发送任意内容都可，只做标志区分查询实名状态
}
type AddictionResult struct {
	GameTime int 		`json:"gameTime"`  //玩家的游戏时间
	CanGoOn  int 		`json:"canGoOn"` //是否进入防沉迷
}
*/

//返回玩家实名认证结果
type VerifyResult struct {
	Status   int32 `json:"status"`   //玩家实名认证状态,1成功,2未实名,3信息重复
	Age      int32 `json:"age"`      //玩家年龄
	GameTime int32 `json:"gameTime"` //玩家的游戏时间
	CanGoOn  int32 `json:"canGoOn"`  //是否进入防沉迷
	//CanPurchase     int32 		`json:"canPurchase"`			//是否购买成功
	ResiduePurchase int32 `json:"residuePurchase"` //当月剩余消费额度
}

//处理请求
func MessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	//存储发过来的
	var receiveMes = new(VerifyMes)
	err := json.NewDecoder(r.Body).Decode(receiveMes)
	if err != nil {
		logs.Error("读取json信息失败%v", err)
		return
	}
	logs.Info("收到的信息为:%v", receiveMes)
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	var UserId = receiveMes.UserId // r.FormValue("userId")身份证号码
	var Accout = receiveMes.Accout //r.FormValue("account")
	var interal=receiveMes.Interval//时间间隔
	//var gameExitTime= receiveMes.GameExitTime //r.FormValue("gameExitTime")
	//登陆后发来的实名认证消息
	if UserId != "" {
		//到数据库查询玩家信息是否重复,不重复就存储
		var gameSign = receiveMes.GameSign
		num := db.VerifyMesRepeat(gameSign, UserId)
		//信息重复，返回重新实名
		if num > 0 {
			_ = json.NewEncoder(w).Encode(&VerifyResult{Status: 3, Age: 0})
		} else {
			//进行认证，存储信息,判断是否未成年
			var age, residuePurchase = EveryDayUpdateAge(receiveMes.UserId)
			var lastLoginTime = db.SetLoginOrExitTimeForVerify(time.Now().Hour(), time.Now().Minute(), 0)
			//设置玩家剩余游戏时间
			var minute,canGoOn int32
			canGoOn=1
			if age<18{
				minute, _ = db.CalculateUserGameTime(lastLoginTime, lastLoginTime)
				if minute<=0{
					canGoOn=2
				}
			}else {
				minute=0
			}
			//信息存储到redis保存一天，晚上统一写进sql
			var verifyMes = &Config.UserAntiAddiction{
				Acc:             receiveMes.Accout,
				GameSign:        gameSign,
				LastLoginTime:   lastLoginTime,
				LastExitTime:    lastLoginTime,
				CardID:          receiveMes.UserId,
				Age:             age,
				Minute:          minute,
				CanGoOn:         canGoOn,
				ResiduePurchase: residuePurchase,
			}
			logs.Info("玩家的信息为:%v,是否未成年:%v", receiveMes, age)
			//缓存信息进行管理，晚上三点统一存进sql
			//注册当天都将数据缓存到服务器，晚上需要存进数据库
			Config.UpdateUserVerifyMes(gameSign, verifyMes.Minute, verifyMes.Age, verifyMes.CanGoOn, verifyMes.ResiduePurchase, verifyMes.CardID, Accout, verifyMes.LastLoginTime, verifyMes.LastExitTime)
			db.UpdateUserRedisMes(gameSign, verifyMes,true)
			//判断当前时间为晚上十点到第二天八点将时间置为0
			var hour=time.Now().Hour()
			if hour>=22||hour<8{
				if age<18{
					canGoOn=2
				}
				_ = json.NewEncoder(w).Encode(&VerifyResult{Status: 1, Age: age, GameTime: 0, CanGoOn: canGoOn, ResiduePurchase: residuePurchase})
			}else {
				_ = json.NewEncoder(w).Encode(&VerifyResult{Status: 1, Age: age, GameTime: minute, CanGoOn: 1, ResiduePurchase: residuePurchase})
			}
		}
	} else if Accout != "" { //请求玩家状态信息
		var gameSign = receiveMes.GameSign
		//登陆时判断玩家是否实名,返回实名状态信息并缓存未成年信息进行管理
		var mes = db.GetUserVerifyMes(gameSign, Accout)
		//已经实名
		if mes != nil {
			//每天第一次请求时更新一下年龄
			if receiveMes.Times == 1 {
				var age, _ = EveryDayUpdateAge(mes.CardID)
				mes.Age = age
			}
			//用上次的记录计算一下玩家的剩余正常游戏时间
			//如果是登陆后第一次发送，用之前的使用时间作为差值更新登陆时间和退出时间
			//如果也是今天第一次登陆应该是两个时间应该是默认值"0000",得到今天剩余的最大时间90或180
			if mes.Age < 18 {
				var minute, useMinute int32
				if receiveMes.Times == 1 {
					minute, useMinute = db.CalculateUserGameTime(mes.LastLoginTime, mes.LastExitTime)
					mes.Minute = minute
					//更新登陆时间
					mes.LastLoginTime = db.SetLoginOrExitTimeForVerify(time.Now().Hour(), time.Now().Minute(), 0)
					//直接将退出时间更新到与登陆时间差为今天游戏时间useMinute的地方，后面直接在此基础加
					mes.LastExitTime = db.SetLoginOrExitTimeForVerify(time.Now().Hour(), time.Now().Minute(), int(useMinute))
				} else {
					//不是登陆后第一次发送就更新退出时间，剩余时间用上一次计算的剩余时间减去post时间
					mes.LastExitTime = db.SetExitTimeWithNum(mes.LastExitTime, int(interal/60))
					mes.Minute -= interal/60
				}
				//更新玩家的消费额度
				if receiveMes.ConsumeNum > 0 {
					mes.ResiduePurchase -= receiveMes.ConsumeNum
				}

				logs.Info("玩家剩余额度:%v,剩余游戏时间:%v,已经游戏时间%v", mes.ResiduePurchase, mes.Minute, useMinute)
			} else {
				//成年玩家不处理游戏时间和消费额度,理论上应该只有登陆时请求一次信息
				mes.Minute = 0
				mes.ResiduePurchase = 0
			}
			//var minute=db.CalculateUserGameTime(mes.LastLoginTime,mes.LastExitTime)
			if mes.Minute <= 0 && mes.Age < 18 {
				mes.Minute = 0
				mes.CanGoOn = 2
			}
			//判断当前时间为晚上十点到第二天八点将时间置为0
			var hour=time.Now().Hour()
			var cangoOn=mes.CanGoOn
			if hour>=22||hour<8{
				logs.Info("到了晚上十点后或者早上八点前，未成年不能登录游戏")
				if mes.Age<18{
					cangoOn=2
				}
				_ = json.NewEncoder(w).Encode(VerifyResult{
					Status:          5,
					Age:             mes.Age,
					GameTime:        0,
					CanGoOn:         cangoOn,
					ResiduePurchase: mes.ResiduePurchase,
				})
			}else {
				_ = json.NewEncoder(w).Encode(VerifyResult{
					Status:          5,
					Age:             mes.Age,
					GameTime:        mes.Minute,
					CanGoOn:         cangoOn,
					ResiduePurchase: mes.ResiduePurchase,
				})
			}
			db.UpdateUserRedisMes(gameSign, mes,true)
		} else {
			//未实名，强制他实名
			logs.Info("此玩家未实名返回状态2")
			_ = json.NewEncoder(w).Encode(VerifyResult{Status: 2})
		}
	} else {
		http.Error(w, "args are wrong", 400)
	}
}

func main() {
	logs.Info("版本:0601-1")
	//初始化数据库
	Config.InitDatabase()
	db.InitDB()
	//处理数据库信息
	go func() {
		db.HandleSqlUpdateUserVerify()
	}()
	//重置玩家消费金额
	go func() {
		db.HandleResetSqlPurchaseMes()
	}()
	go func() {
		//handleSqlThreeHour()
	}()
	go func() {
		HandleSqlEveryDay()
	}()
	http.HandleFunc("/", MessageHandler)
	http.ListenAndServe(":80", nil)
}

//三小时更新一下玩家数据库数据
func handleSqlThreeHour() {
	var t *time.Ticker
	for {
		//定义每三小时更新
		now := time.Now()
		hh, _ := time.ParseDuration("3h")
		//workStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.Local).Add(mm)
		workStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.Local).Add(hh)
		///周期性定时，每三小时执行一次
		t = time.NewTicker(workStart.Sub(now))
		<-t.C
		logs.Info("更新数据库玩家的实名信息")
		//更新玩家的实名认证信息
		for _, gameMeses := range Config.UserVerifyMap.Items() {
			for _, mes := range gameMeses.(*utils.BeeMap).Items() {
				db.UserUpdateVerifyChan <- mes.(*Config.UserAntiAddiction)
			}
		}
	}
}

//每天三点清理服务器数据,重置未成年的游戏时间
func HandleSqlEveryDay() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			errStr := string(buf[:l])
			logs.Info("写入sql任务Panic： %v -- %s", err, errStr)
		}
	}()
	var t *time.Ticker
	for {
		//定义每天三点更新
		now := time.Now()
		//mm, _ := time.ParseDuration("1m")
		//workStart := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, time.Local).Add(mm)
		workStart := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, time.Local).AddDate(0, 0, 1)
		///周期性定时，每天三点执行一次
		t = time.NewTicker(workStart.Sub(now))
		<-t.C
		//更新玩家的实名认证信息
		for _, gameMeses := range Config.UserVerifyMap.Items() {
			for _, mes := range gameMeses.(*utils.BeeMap).Items() {
				//重置玩家的登陆时间和退出时间以重置每日游戏时间
				mes.(*Config.UserAntiAddiction).LastLoginTime = "0000"
				mes.(*Config.UserAntiAddiction).LastExitTime = "0000"
				mes.(*Config.UserAntiAddiction).CanGoOn = 1
				db.UserUpdateVerifyChan <- mes.(*Config.UserAntiAddiction)
			}
		}
	}
}

///得到玩家年龄和每月消费额度
func JudgeAge(year, month, day int) (int32, int32) {
	var nowY = time.Now().Year()
	var nowMonth = time.Now().Month()
	var nowDay = time.Now().Day()
	var offY = nowY - year
	var offM = int(nowMonth) - month
	var offD = nowDay - day
	if offD < 0 {
		offD += 30
		offM -= 1
	}
	if offM < 0 {
		offM += 12
		offY = offY - 1
	}
	//8岁以下禁止消费，18岁以上无限制
	if offY >= 18 || offY < 8 {
		return int32(offY), 0
	} else if offY < 16 {
		return int32(offY), 200
	} else if offY < 18 {
		return int32(offY), 400
	}
	return int32(offY), 0
}

//每天更新玩家的年龄
func EveryDayUpdateAge(cardID string) (int32, int32) {
	var yearStr = cardID[6:10]
	var monthStr = cardID[10:12]
	var dayStr = cardID[12:14]
	var day, month, year int
	var err error
	if day, err = strconv.Atoi(dayStr); err != nil {
		logs.Error("string转换Int类型失败")
	}
	if month, err = strconv.Atoi(monthStr); err != nil {
		logs.Error("string转换Int类型失败")
	}
	if year, err = strconv.Atoi(yearStr); err != nil {
		logs.Error("string转换Int类型失败")
	}
	//得到玩家年龄和对应的消费额度
	var age, residuePurchase = JudgeAge(year, month, day)
	return age, residuePurchase
}
