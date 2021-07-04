package Config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/utils"
)

var (
	HelmReleaseName    = "release"
	UserVerifyMap      = utils.NewBeeMap() //key是游戏标示,value是游戏未成年用户实名信息
	holidaysConfigPath = "/config/节假日.json"
	HolidaysData       *Holidays
)

type Holidays struct {
	Days []string
}

func InitDatabase() {
	err := initMysql(true)
	if err != nil {
		fmt.Printf("初始化mysql失败：%s\n", err.Error())
		//return
	}
	err = initRedis(true)
	if err != nil {
		fmt.Printf("初始化redis失败：%s\n", err.Error())
		return
	}
	HolidaysData = &Holidays{Days: make([]string, 0)}
	bytes, err := ioutil.ReadFile(holidaysConfigPath)
	if err != nil {
		logs.Error("读取节假日配置出错：%v", err)
	}

	err = json.Unmarshal(bytes, HolidaysData)
	if err != nil {
		logs.Error("解析节假日json出错：%v", err)
	}
	logs.Info("节假日%v", HolidaysData.Days)

}

//判断一天是否是节假日,格式:0329
func GetOneDayIsHoliday(day string) bool {
	for _, d := range HolidaysData.Days {
		if d == day {
			return true
		}
	}
	return false
}

///玩家的实名认证信息
type UserAntiAddiction struct {
	Acc             string
	GameSign        string
	LastLoginTime   string
	LastExitTime    string
	CardID          string //身份证号
	Age             int32
	Minute          int32
	CanGoOn         int32
	ResiduePurchase int32
	//IsReallyVerify bool
	//IsNeedUpdateSql bool
	//IsNeedUpdateTime   bool
}

//添加一个用户的年龄信息
func UpdateUserVerifyMes(gameSign string, minute, age, canGO, residuePurchase int32, cardId, acc, lastLoginTime, lastExitTime string) {
	if !UserVerifyMap.Check(gameSign) {
		UserVerifyMap.Set(gameSign, utils.NewBeeMap())
	}
	//map键是gameSign,value是一个键是account,value是实名信息的map
	if !(UserVerifyMap.Get(gameSign).(*utils.BeeMap).Check(acc)) {
		UserVerifyMap.Get(gameSign).(*utils.BeeMap).Set(acc, &UserAntiAddiction{
			Acc:             acc,
			GameSign:        gameSign,
			LastLoginTime:   lastLoginTime,
			LastExitTime:    lastExitTime,
			CardID:          cardId, //身份证号
			Age:             age,
			Minute:          minute,
			CanGoOn:         canGO,
			ResiduePurchase: residuePurchase,
			//IsReallyVerify:   isrellyVerify,
			//IsNeedUpdateSql:  false,
			//IsNeedUpdateTime: true,
		})
	} else {
		var mes = UserVerifyMap.Get(gameSign).(*utils.BeeMap).Get(acc).(*UserAntiAddiction)
		mes.Acc = acc
		mes.LastLoginTime = lastLoginTime
		mes.LastExitTime = lastExitTime
		mes.CardID = cardId //身份证号
		mes.Age = age
		mes.Minute = minute
		mes.CanGoOn = canGO
		//mes.IsReallyVerify = isrellyVerify
	}
}

//得到一个用户的防沉迷信息
func GetUserVerifyMes(gameSign, account string) *UserAntiAddiction {
	if UserVerifyMap.Check(gameSign) {
		if UserVerifyMap.Get(gameSign).(*utils.BeeMap).Check(account) {
			var addiction = UserVerifyMap.Get(gameSign).(*utils.BeeMap).Get(account)
			return addiction.(*UserAntiAddiction)
		} else {
			logs.Info("列表中不存在用户:%v的防沉迷信息", account)
			return nil
		}
	} else {
		logs.Info("列表中不存在游戏:%v的防沉迷信息", gameSign)
		return nil
	}
}

func JudgeUserAlreadyRegister(gameSign,cardId string)bool{
	if UserVerifyMap.Check(gameSign){
         var nbMap=UserVerifyMap.Get(gameSign).(*utils.BeeMap)
         for _,info:=range nbMap.Items(){
         	if info.(*UserAntiAddiction).CardID==cardId{
         		return  true
			}
		 }
		 return  false
	}else  {
		return  false
	}
}
//删除一个用户的防沉迷信息
func DeleteUserVerifyMes(gameSign, account string) {
	if UserVerifyMap.Check(gameSign) {
		if UserVerifyMap.Get(gameSign).(*utils.BeeMap).Check(account) {
			UserVerifyMap.Get(gameSign).(*utils.BeeMap).Delete(account)
		} else {
			logs.Error("列表中不存在用户:%v的防沉迷信息", account)
		}
	} else {
		logs.Error("列表中不存在游戏:%v的防沉迷信息", gameSign)
	}
}

//删除一个游戏的防沉迷信息
func DeleteGameVerifyMes(gameSign string) {
	if UserVerifyMap.Check(gameSign) {
		UserVerifyMap.Delete(gameSign)
	} else {
		logs.Error("列表中不存在游戏:%v的防沉迷信息", gameSign)
	}
}
