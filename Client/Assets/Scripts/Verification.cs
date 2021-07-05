using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Text;
using UnityEngine;
using UnityEngine.Networking;
/// <summary>
/// 
/// </summary>
public class Verification : MonoSingleton<Verification>
{
    /// <summary>
    /// 标志玩家是否进入防沉迷，既是否能继续正常游戏
    /// </summary>
    public bool IsEnterAntiAddiction
    {
        get; private set;
    }
    /// <summary>
    /// 存储玩家年龄,支付时调用判断年龄，18岁以下单笔不可超过50元
    /// </summary>
    public int Age
    {
        get; private set;
    }
    /// <summary>
    /// 存储玩家剩余游戏时长，晚上22点至次日8点登陆不能提供游戏服务，那时登陆恒为0，需要提示时加以说明
    /// 结合Age使用，成年人恒为0
    /// </summary>
    public int ResidueGameTime
    {
        get; private set;
    }
    /// <summary>
    /// 当天可游戏的总时间
    /// </summary>
    public int TodayTotalTime
    {
        get; private set;
    }
    /// <summary>
    /// 存储玩家剩余的消费额度，需要和Age结合使用,成年人恒为0
    /// </summary>
    public int ResiduePurchase
    {
        get; private set;
    }
    /// <summary>
    /// 玩家账号，需要是此游戏中的唯一标识,如果没有可以也传身份证号，但是需要缓存方便下次登陆使用
    /// awake中赋值
    /// </summary>
    public string Account
    {
        get; private set;
    }
    /// <summary>
    /// 游戏标识，自定义，一个游戏用一个不能重复不能更改,setGameBaseMes中赋值
    /// </summary>
    public string GameSign
    {
        get; private set;
    }
    /// <summary>
    /// 请求的时间间隔
    /// </summary>
    public int Interval
    {
        get; private set;
    }
    /// <summary>
    /// 需要进行实名认证回调
    /// </summary>
    public Action NeedCollectRealNameHandler;
    /// <summary>
    /// 实名认证信息通过回调
    /// </summary>
    public Action VerifyMesSuccessHandler;
    /// <summary>
    /// 实名认证信息已经被注册需要重新认证回调
    /// </summary>
    public Action NeedCollectAgainHandler;
    /// <summary>
    /// 收到玩家防沉迷状态回调,可以在这里通过IsEnterAntiAddiction的值判断玩家的防沉迷状态并做相应处理
    /// </summary>
    public Action AntiAddictionStatusHandler;

    /// <summary>
    /// 统计请求次数，该值不能被重置为1，
    /// 如果此脚本有可能被销毁，需要将此值放在数据类中供此类使用
    /// </summary>
    public int Times = 1;

    /// <summary>
    /// 服务器地址，更换为服务器部署后的地址
    /// </summary>
    string _url = "";

    /// <summary>
    /// 定义1分钟发起一次请求
    /// </summary>
    int _interval = 60;



    /// <summary>
    /// 持有运行的持续post的协程
    /// </summary>
    Coroutine _coroutine;



    /// <summary>
    /// 游戏信息初始化调用一次设置游戏标识
    /// </summary>
    /// <param name="gameSign">游戏标识</param>
    /// <param name="account">游戏唯一账号id,单机游戏可不传此参数</param>
    public void SetGameBaseMes(string gameSign, string account="")
    {
        //单机游戏拿AndroidID当作唯一账号
        if (account == "")
        {
            var javaClass = new AndroidJavaClass("com.wingjoy.ghostshock.inter.myapplication");
            account = javaClass.CallStatic<string>("getAndroidId");
        }
        GameSign = gameSign;
        Account = account;
        //定义一次请求的时间间隔，默认600s
        if (Interval == 0)
        {
            Interval = _interval;
        }
        //更新当天可以游戏的时间
        TotalTimeToday();
    }

    /// <summary>
    /// 计算当天还有多久
    /// </summary>
    public void TotalTimeToday()
    {
        TodayTotalTime = 90;
        var listHoliday = new Dictionary<int, List<int>>();
        listHoliday.Add(6, new List<int>());
        listHoliday[6].Add(12);
        listHoliday[6].Add(13);
        listHoliday[6].Add(14);
        listHoliday.Add(9, new List<int>());
        listHoliday[9].Add(19);
        listHoliday[9].Add(20);
        listHoliday[9].Add(21);
        listHoliday.Add(10, new List<int>());
        listHoliday[10].Add(1);
        listHoliday[10].Add(2);
        listHoliday[10].Add(3);
        listHoliday[10].Add(4);
        listHoliday[10].Add(5);
        listHoliday[10].Add(6);
        listHoliday[10].Add(7);
        var nowDay = DateTime.Now.Day;
        var nowMonth = DateTime.Now.Month;
        //节假日
        if (listHoliday.ContainsKey(nowMonth) && listHoliday[nowMonth].Contains(nowDay))
        {
            TodayTotalTime = 180;
        }
    }

    /// <summary>
    /// 玩家实名信息输入后客户端先检测是否有效，确认有效调用此请求玩家检查此信息是否已经被实名认证
    /// </summary>
    /// <param name="CardId"></param>
    /// <param name="times"></param>
    public void RequestVerifyStatus(string CardId)
    {
        if (GameSign == "" || Account == "")
        {
            Debug.LogError("请先调用setGameBaseMes初始化游戏信息");
            return;
        }
        string postData = JsonUtility.ToJson(new RequestVerifyStatusMes()
        {
            account = Account,
            userId = CardId,
            gameSign = GameSign,
            times = Times
        });
        StartCoroutine(PostData(postData));
    }


    /// <summary>
    /// 登陆成功调用此方法传入查询玩家是否实名，返回状态信息，客户端进行处理和缓存
    /// 在游戏内需隔一段时间或者再次调用此方法更新状态信息,已经在此脚本OnApplicationPause进行了处理
    /// </summary>
    /// <param name="gameSign"></param>
    /// <param name="account"></param>
    /// <param name="isFirst">是否为登陆后第一次发送</param>
    /// <param name="purcharseNum">支付的金额，未成年玩家支付后应调用此方法</param>
    public void RequestVerifyAntiAddiction(int purcharseNum = 0)
    {
        if (GameSign == "" || Account == "")
        {
            Debug.LogError("请先调用setGameBaseMes初始化游戏信息");
            return;
        }
        if (Age > 18 || (IsEnterAntiAddiction == true && purcharseNum<= 0) || purcharseNum > ResiduePurchase)
        {
            //Debug.LogError("此玩家不是未成年玩家或者已经进入防沉迷或者已无消费额度，无需再发起此请求!");
            return;
        }
        var reqMes = new RequestVerifyStatusMes()
        {
            account = Account,
            gameSign = GameSign,
            consumeNum = purcharseNum,
            times = Times,
            interval = Interval
        };
        //有消费的时候请求的游戏时间为0
        if (purcharseNum > 0)
        {
            reqMes.interval = 0;
        }
        string postData = JsonUtility.ToJson(reqMes);

        StartCoroutine(PostData(postData));
        if (_coroutine == null)
        {
            _coroutine = StartCoroutine(PersistentPost());
        }
    }

    /// <summary>
    /// 是否可以让玩家支付
    /// </summary>
    /// <param name="payMoney"></param>
    /// <returns></returns>
    public bool BeforePurchase(float payMoney)
    {
        //if (MicroHandler.LoginType == Login.LoginRequest.Types.LoginType.Youkelogin&&Age==0)
        //{
        //    AntiAddictionManager.Instance.Tips("未进行实名认证无法进行充值");
        //    return false;
        //}
        if (Age >= 16 && Age < 18)
        {
            //单笔超过100，则不能支付
            //if (Mathf.CeilToInt(payMoney) > 100)
            //{
            //    //提示单笔支付超过100或者每月可消费额度已不足
            //    AntiAddictionManager.Instance.Tips("单笔支付不可超过100");
            //    return false;
            //}
            if (Mathf.CeilToInt(payMoney) > 100 || ResiduePurchase < Mathf.CeilToInt(payMoney))
            {
                //每月可消费额度已不足
                AntiAddictionManager.Instance.Tips($"根据国家新闻出版署《关于防止未成年人沉迷网络游戏的通知》," +
                    $"您的充值金额达到了防沉迷限制上限,无法进行充值【16岁至18岁,单笔上限100元,每月上限400元】");
                return false;
            }
        }
        if (Age >= 8 && Age < 16)
        {
            //单笔超过50，则不能支付
            //if (Mathf.CeilToInt(payMoney) > 50)
            //{
            //    //提示单笔支付不可超过50
            //    AntiAddictionManager.Instance.Tips("单笔支付不可超过50");
            //    return false;
            //}
            if (Mathf.CeilToInt(payMoney) > 50 || ResiduePurchase < Mathf.CeilToInt(payMoney))
            {
                //每月可消费额度已不足
                AntiAddictionManager.Instance.Tips($"根据国家新闻出版署《关于防止未成年人沉迷网络游戏的通知》," +
                    $"您的充值金额达到了防沉迷限制上限,无法进行充值【8岁至16岁,单笔上限50元,每月上限200元】");
                return false;
            }
        }
        if (Age < 8)
        {
            //提示当前年龄不可支付
            AntiAddictionManager.Instance.Tips("根据国家新闻出版署《关于防止未成年人沉迷网络游戏的通知》,未满8周岁无法进行充值");
            return false;
        }
        return true;
    }

    /// <summary>
    /// 玩家支付成功后请求服务器更新信息
    /// </summary>
    public void PurcharseSuccess(float payMoney)
    {

        if (Age < 18)
        {
            RequestVerifyAntiAddiction(Mathf.CeilToInt(payMoney));
        }
    }

    /// <summary>
    /// Post请求
    /// </summary>
    /// <param name="postData"></param>
    /// <returns></returns>
    IEnumerator PostData(string postData)
    {
        //form.AddField("gameSign", DateTime.Now.ToString());
        using (UnityWebRequest webRequest = new UnityWebRequest(_url, UnityWebRequest.kHttpVerbPOST))
        {
            Debug.Log($"要发送的信息为:{postData}");
            UploadHandler uploader = new UploadHandlerRaw(System.Text.Encoding.Default.GetBytes(postData));
            webRequest.uploadHandler = uploader;
            webRequest.uploadHandler.contentType = "application/json";  //设置HTTP协议的请求头，默认的请求头HTTP服务器无法识别
                                                                        //这里需要创建新的对象用于存储请求并响应后返回的消息体，否则报空引用的错误
            DownloadHandler downloadHandler = new DownloadHandlerBuffer();
            webRequest.downloadHandler = downloadHandler;
            // Request and wait for the desired page.
            yield return webRequest.SendWebRequest();

            if (webRequest.isNetworkError || webRequest.isHttpError)
            {
                Debug.LogError($"请求出错" + webRequest.error + webRequest.isHttpError);
            }
            else
            {
                var responseMes = JsonUtility.FromJson<ResponseVerifyStatusMes>(webRequest.downloadHandler.text);
                Debug.Log($"收到的信息为:{webRequest.downloadHandler.text}");

                //回调处理
                if (responseMes.status > 0)
                {
                    switch (responseMes.status)
                    {

                        case 1:         //1:提交的实名信息有效，认证成功
                            Times++;
                            SetProperty(responseMes);
                            VerifyMesSuccessHandler?.Invoke();
                            break;
                        case 2:         //玩家还未认证，需要认证的处理
                            NeedCollectRealNameHandler?.Invoke();
                            break;
                        case 3:          //信息重复提示重新认证的处理
                            NeedCollectAgainHandler?.Invoke();
                            break;
                        case 5:         //玩家之前已经认证过，并返回状态信息
                            Times++;
                            SetProperty(responseMes);
                            AntiAddictionStatusHandler?.Invoke();
                            break;
                    }
                }

            }
        }
    }


    /// <summary>
    /// 收到玩家防沉迷状态后更新脚本中存储的信息
    /// </summary>
    /// <param name="mes"></param>
    void SetProperty(ResponseVerifyStatusMes mes)
    {
        Age = mes.age;
        ResidueGameTime = mes.gameTime;
        ResiduePurchase = mes.residuePurchase;
        if (ResidueGameTime <= 0)
        {
            ResidueGameTime = 0;
        }
        if (ResiduePurchase <= 0)
        {
            ResiduePurchase = 0;
        }
        IsEnterAntiAddiction = mes.canGoOn == 1 ? false : true;
    }

    /// <summary>
    ///客户端从进入游戏开始每隔一段时间向服务器请求一次状态
    /// </summary>
    /// <returns></returns>
    IEnumerator PersistentPost()
    {
        //等待第一次请求完成
        while (Times == 1)
        {
            yield return null;
        }
    //AntiAddictionManager.Instance.YoungManger(true);
    Persistent:
        while (_interval > 0)
        {
            yield return new WaitForSeconds(1f);
            _interval--;
        }
        _interval = 60;
        if (GameSign == "" || Account == "")
        {
            Debug.LogError("请先调用setGameBaseMes初始化游戏信息");
            goto Persistent;
        }
        RequestVerifyAntiAddiction();
        goto Persistent;
    }
    /// <summary>
    /// 发送Get请求，此处没用先放着，请求实名状态必须用post
    /// </summary>
    /// <param name="url"></param>
    /// <param name="data"></param>
    /// <returns></returns>
    string HttpGet(string url, string data)
    {
        try
        {
            //创建Get请求

            url = url + (data == "" ? "" : "?") + data;

            HttpWebRequest request = (HttpWebRequest)WebRequest.Create(url);

            request.Method = "GET";

            request.ContentType = "text/html;charset=UTF-8";

            //接受返回来的数据

            HttpWebResponse response = (HttpWebResponse)request.GetResponse();

            Stream stream = response.GetResponseStream();

            StreamReader streamReader = new StreamReader(stream, Encoding.GetEncoding("utf-8"));

            string retString = streamReader.ReadToEnd();
            Debug.Log($"得到的信息为:{retString}");
            streamReader.Close();

            stream.Close();

            response.Close();

            return retString;

        }
        catch (Exception e)
        {
            Debug.LogError(e.Message);
            return "";
        }

    }


    private void Update()
    {
        if (_coroutine == null)
        {
            Debug.Log("pauseStartCoroutine");
            _coroutine = StartCoroutine(PersistentPost());
        }
    }
    private void OnApplicationPause(bool pause)
    {
        //切出游戏
        if (pause)
        {
            if (_coroutine != null)
            {
                Debug.Log("pauseStopCoroutine");
                StopCoroutine(_coroutine);
                _coroutine = null;
            }
        }
        //切入游戏运行，每隔一段时间post一次
        else
        {
            if (_coroutine == null)
            {
                Debug.Log("pauseStartCoroutine");
                _coroutine = StartCoroutine(PersistentPost());
            }
        }
    }


    /// <summary>
    /// 请求消息
    /// userId!=""时发送玩家填写的实名认证信息到服务器验证
    /// gameExitTime!=""时为得到玩家游戏时长和是否进入防沉迷
    /// </summary>
    [Serializable]
    struct RequestVerifyStatusMes
    {
        public string account;
        public string userId;
        public string gameSign;
        public int consumeNum;
        public int times;
        public int interval;
    }
    /// <summary>
    /// 返回消息
    /// </summary>
    [Serializable]
    struct ResponseVerifyStatusMes
    {
        /// <summary>
        /// 实名认证状态,1已经成功实名,2未实名，需要进行实名,3实名信息重复，提示重新实名
        /// </summary>
        public int status;
        /// <summary>
        /// 标志是否未成年,1未成年
        /// </summary>
        public int age;
        /// <summary>
        /// 玩家在线时长
        /// </summary>
        public int gameTime;
        /// <summary>
        /// 是否进入了防沉迷
        /// </summary>
        public int canGoOn;
        /// <summary>
        /// 剩余的消费额度，使用时应首先判断年龄<18,成年玩家恒等于0
        /// </summary>
        public int residuePurchase;
    }
}


