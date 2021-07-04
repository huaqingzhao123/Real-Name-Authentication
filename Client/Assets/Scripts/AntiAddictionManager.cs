
using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

public class AntiAddictionManager : MonoSingleton<AntiAddictionManager>
{
    Coroutine _coroutine;
    private Transform _youngTip;
    private Transform _TipPanel;
    private Text _TipText;
    float youkeLimitTime = 0;
    //游戏进入后先调用一次
    private float ShowTime = 0;
    private void Awake()
    {
        _TipPanel = transform.Find("YoungTipPanel");
        _TipText = _TipPanel.Find("tip").GetComponent<Text>();
        //处理未成年到时间退出游戏
        Verification.Instance.AntiAddictionStatusHandler += () =>
        {
            YoungManger(false);
            //进入防沉迷，各自游戏可各自统计玩家状态，此处为幽灵服务器统计玩家防沉迷
            if (Verification.Instance.IsEnterAntiAddiction)
            {
                //MicroHandler.Inst.CollectAntiAddictionStatus();
            }
            //时间到了显示一次面板
            if (Verification.Instance.ResidueGameTime <= 0 && Verification.Instance.Age < 18 && Verification.Instance.Age > 0)
            {
                ShowTips();
            }
            //if (_youngTip == null)
            //{
            //    _youngTip = transform.Find("YoungTip");
            //}
            //if (Verification.Instance.Age < 18)
            //{
            //    if (Verification.Instance.IsEnterAntiAddiction)
            //    {
            //        //游戏时间到或者在十点至第二天八点登陆游戏
            //        _youngTip.Find("1").gameObject.SetActive(false);
            //        _youngTip.Find("Content").gameObject.SetActive(false);
            //        //十点到第二天八点不提供游戏服务
            //        if (JudgeTime())
            //        {
            //            _youngTip.Find("GameStop").gameObject.SetActive(true);
            //        }
            //        //如果不是十点到第二天八点显示时间用完
            //        else
            //        {
            //            _youngTip.Find("GameOver").gameObject.SetActive(true);
            //        }
            //        _youngTip.gameObject.SetActive(true);
            //        //强制退出游戏
            //        _youngTip.Find("ConfirmBtn").GetComponent<Button>().onClick.AddListener(() =>
            //        {
            //            Application.Quit();
            //        });
            //    }
            //}
        };
        _coroutine = StartCoroutine(TipAnti());
    }

    private void Update()
    {
        if (_coroutine == null)
        {
            _coroutine = StartCoroutine(TipAnti());
        }
    }
    public void YoungManger(bool isFirst = false)
    { 
        if (_youngTip == null)
        {
            _youngTip = transform.Find("YoungTip");
        }
        ShowTime = Time.frameCount;
        //小于18岁
        if (Verification.Instance.Age < 18)
        {
            if (Verification.Instance.ResidueGameTime > 0)
            {
                //Debug.LogError("显示面板");
                _youngTip.Find("1").GetComponent<Text>().text = $"您的年龄为:{Verification.Instance.Age}岁";
                //设置提示的文本
                //小于8岁
                if (Verification.Instance.Age < 8)
                {
                    _youngTip.Find("Content/3").GetComponent<Text>().text = $"2.小于8周岁不能使用支付功能";
                }
                //大于8小于16岁
                else if (Verification.Instance.Age < 16)
                {
                    _youngTip.Find("Content/3").GetComponent<Text>().text = $"2.单笔支付不能大于50，月消费不能超过200";
                }
                //大于16小于18岁
                else
                {
                    _youngTip.Find("Content/3").GetComponent<Text>().text = $"2.单笔支付不能大于100，月消费不能超过400";
                }
                _youngTip.Find("Content/2").GetComponent<Text>().text = $"当天可游戏时间为{Verification.Instance.TodayTotalTime}分，剩余:{Verification.Instance.ResidueGameTime}分";
                _youngTip.Find("Content/4").GetComponent<Text>().text = $"当月剩余消费额度:{Verification.Instance.ResiduePurchase}";
            }
            else
            {
                //游戏时间到或者在十点至第二天八点登陆游戏
                _youngTip.Find("1").gameObject.SetActive(false);
                _youngTip.Find("Content").gameObject.SetActive(false);
                //十点到第二天八点不提供游戏服务
                if (JudgeTime())
                {
                    _youngTip.Find("GameStop").gameObject.SetActive(true);
                }
                //如果不是十点到第二天八点显示时间用完
                else
                {
                    _youngTip.Find("GameOver").gameObject.SetActive(true);
                }
                //强制退出
                _youngTip.Find("ConfirmBtn").GetComponent<Button>().onClick.AddListener(() =>
                {
                    Application.Quit();
                });

            }
        }
        else
        {
            _youngTip.gameObject.SetActive(false);
        }
        if (isFirst)
        {
            _youngTip.gameObject.SetActive(true);
        }
    }

    public void ShowTips()
    {
        if (_youngTip == null)
        {
            _youngTip = transform.Find("YoungTip");
        }
        _youngTip.gameObject.SetActive(true);
    }


    public void Tips(string tip)
    {
        _TipText.text = tip;
        _TipPanel.gameObject.SetActive(true);
    }
    /// <summary>
    /// 判断当前时间
    /// </summary>
    /// <returns></returns>
    bool JudgeTime()
    {
        var hour = DateTime.Now.Hour;
        if (hour < 8 || hour >= 22)
        {
            return true;
        }
        return false;
    }

    /// <summary>
    /// 每分钟调用一次，为了在十点的时候强制退出游戏
    /// </summary>
    /// <returns></returns>
    IEnumerator TipAnti()
    {
        while (true)
        {
            if (Verification.Instance.Age < 18 && Verification.Instance.Age > 0)
            {
                if (JudgeTime())
                {
                    if (_youngTip == null)
                    {
                        _youngTip = transform.Find("YoungTip");
                    }
                    _youngTip.Find("1").gameObject.SetActive(false);
                    _youngTip.Find("Content").gameObject.SetActive(false);
                    _youngTip.Find("GameOver").gameObject.SetActive(false);
                    //十点到第二天八点显示不提供游戏
                    _youngTip.Find("GameStop").gameObject.SetActive(true);
                    //强制退出游戏
                    _youngTip.Find("ConfirmBtn").GetComponent<Button>().onClick.AddListener(() =>
                    {
                        Application.Quit();
                    });
                    _youngTip.gameObject.SetActive(true);
                }
            }
            //一分钟判断一次
            yield return new WaitForSeconds(60f);
            //统计登陆时长
            youkeLimitTime++;
            if (youkeLimitTime >= 120 /*&& MicroHandler.LoginType == Login.LoginRequest.Types.LoginType.Youkelogin*/)
            {
                transform.Find("YouKeTipPanel/ConfirmBtn").GetComponent<Button>().onClick.AddListener(() =>
                {
                    Application.Quit();
                });
                transform.Find("YouKeTipPanel").gameObject.SetActive(true);
            }
        }
    }
}
