using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
using UnityEngine;
using UnityEngine.UI;
/// <summary>
/// 实名认证相关页面
/// </summary>
public class VerificationID : MonoBehaviour
{
    Text _Output; //成功失败的回调
    InputField _InputNum;  //用户输入的身份证号码
    InputField _InputName;  //用户输入的姓名
    Transform _accessText;
    //用户输入的身份证号
    private string _id;
    //用户输入的姓名
    string _name;
    Button _verifiCationBtn;
    StringBuilder _stringBuilder;
    private void Awake()
    {
        _stringBuilder = new StringBuilder();
        //if (/*BDLauncher.GamePlatform == BDLauncher.PlatformType.ChinaTW*/)
        //{
        //    return;
        //}
        _Output = transform.GetChild(0).Find("BG/OutPutText").GetComponent<Text>();
        _InputNum = transform.GetChild(0).Find("BG/NumInputField").GetComponent<InputField>();
        _InputName = transform.GetChild(0).Find("BG/NameInputField").GetComponent<InputField>();
        _accessText = transform.GetChild(0).Find("BG/FinshTip");
        _verifiCationBtn = transform.GetChild(0).Find("BG/VerificationBtn").GetComponent<Button>();
        _verifiCationBtn.onClick.AddListener(() => VerificationCard());
        _InputName.onValueChanged.AddListener((text) =>
        {
            if (!string.IsNullOrEmpty(text))
            {
                //Debug.LogError($"是否包含该敏感词:{_InputName.text},{IllegalWordDetection.Instance.strings.Contains(_InputName.text)}");
                int length = _InputName.text.Length;
                text = /*IllegalWordDetection.Filter(_InputName.text);*/IllegalWordDetection.Filter(_InputName.text);
                if (text.Contains("*"))
                {
                    _stringBuilder.Clear();
                    for (int i = 0; i < length; i++)
                    {
                        _stringBuilder.Append("*");
                    }
                    _InputName.text = _stringBuilder.ToString();
                }
                _InputName.textComponent.text = _InputName.text;
                //Debug.Log($"输入框文本{ _InputName.text},显示文本:{_InputName.textComponent.text}");
                Canvas.ForceUpdateCanvases();
            }
        });
        //需要认证
        Verification.Instance.NeedCollectRealNameHandler += () =>
        {
            transform.GetChild(0).gameObject.SetActive(true);
        };
        Verification.Instance.VerifyMesSuccessHandler += () =>
         {
             //_Output.text = "恭喜验证通过";
             _Output.text = "";
             _accessText.gameObject.SetActive(true);
             transform.GetChild(0).Find("BG/Title").gameObject.SetActive(false);
             transform.GetChild(0).GetComponent<Button>().interactable = true;
             if (Verification.Instance.Age > 0 && Verification.Instance.Age < 18)
             {
                 Debug.Log($"Verification.Instance.Age:{Verification.Instance.Age}");
                 AntiAddictionManager.Instance.YoungManger(true);
             }
             transform.GetChild(0).Find("BG/VerificationBtn").GetComponent<Button>().interactable = false;
         };
        Verification.Instance.NeedCollectAgainHandler += () =>
        {
            _Output.text = "此身份信息已经被注册！";
        };
        #region 幽灵中的使用,其它游戏可根据情况在合适时机调用
        //初始化防沉迷游戏信息
        Verification.Instance.SetGameBaseMes("ghost", "123");
        //在登陆时请求一次
        if (Verification.Instance.Times == 1)
        {
            Verification.Instance.RequestVerifyAntiAddiction();
        }
        #endregion
    }
    private void OnDestroy()
    {
        Verification.Instance.NeedCollectRealNameHandler = null;
        Verification.Instance.VerifyMesSuccessHandler = null;
        Verification.Instance.NeedCollectAgainHandler = null;
    }
    //点击按钮之后验证
    public void VerificationCard()
    {
        if (string.IsNullOrEmpty(_InputName.text))
        {
            _Output.text = "请输入姓名！";
            return;
        }
        //将用户输入的身份证号码赋值到string类型的值上去
        _id = _InputNum.text;
        _name = _InputName.text;
        if (_id.Length == 18)
        {
            //基本验证位数18位已经通过执行拆分方法
            Resolution();
        }
        else
        {
            //基本验证没有通过，位数多或者少，不合法
            Debug.Log(_id);
            _Output.text = "身份证验证不通过，原因：身份证位数不够！";
        }
    }
    //基本位数验证通过之后开始拆分计算
    void Resolution()
    {
        //将所输入的身份证号码进行拆分，拆分为17位，最后一位留着待用
        int Num0, Num1, Num2, Num3, Num4, Num5, Num6, Num7, Num8, Num9, Num10,
            Num11, Num12, Num13, Num14, Num15, Num16;
        int numTemp;
        Num0 = JudgePerChar(_id, 0, 7);
        if (Num0 < 0) return;
        Num1 = JudgePerChar(_id, 1, 9);
        if (Num1 < 0) return;
        Num2 = JudgePerChar(_id, 2, 10);
        if (Num2 < 0) return;
        Num3 = JudgePerChar(_id, 3, 5);
        if (Num3 < 0) return;
        Num4 = JudgePerChar(_id, 4, 8);
        if (Num4 < 0) return;
        Num5 = JudgePerChar(_id, 5, 4);
        if (Num5 < 0) return;
        Num6 = JudgePerChar(_id, 6, 2);
        if (Num6 < 0) return;
        Num7 = JudgePerChar(_id, 7, 1);
        if (Num7 < 0) return;
        Num8 = JudgePerChar(_id, 8, 6);
        if (Num8 < 0) return;
        Num9 = JudgePerChar(_id, 9, 3);
        if (Num9 < 0) return;
        Num10 = JudgePerChar(_id, 10, 7);
        if (Num10 < 0) return;
        Num11 = JudgePerChar(_id, 11, 9);
        if (Num11 < 0) return;
        Num12 = JudgePerChar(_id, 12, 10);
        if (Num12 < 0) return;
        Num13 = JudgePerChar(_id, 13, 5);
        if (Num13 < 0) return;
        Num14 = JudgePerChar(_id, 14, 8);
        if (Num14 < 0) return;
        Num15 = JudgePerChar(_id, 15, 4);
        if (Num15 < 0) return;
        Num16 = JudgePerChar(_id, 16, 2);
        if (Num16 < 0) return;
        //int Num0 = int.Parse(_id.Substring(0, 1)) * 7;
        //int Num1 = int.Parse(_id.Substring(1, 1)) * 9;
        //int Num2 = int.Parse(_id.Substring(2, 1)) * 10;
        //int Num3 = int.Parse(_id.Substring(3, 1)) * 5;
        //int Num4 = int.Parse(_id.Substring(4, 1)) * 8;
        //int Num5 = int.Parse(_id.Substring(5, 1)) * 4;
        //int Num6 = int.Parse(_id.Substring(6, 1)) * 2;
        //int Num7 = int.Parse(_id.Substring(7, 1)) * 1;
        //int Num8 = int.Parse(_id.Substring(8, 1)) * 6;
        //int Num9 = int.Parse(_id.Substring(9, 1)) * 3;
        //int Num10 = int.Parse(_id.Substring(10, 1)) * 7;
        //int Num11 = int.Parse(_id.Substring(11, 1)) * 9;
        //int Num12 = int.Parse(_id.Substring(12, 1)) * 10;
        //int Num13 = int.Parse(_id.Substring(13, 1)) * 5;
        //int Num14 = int.Parse(_id.Substring(14, 1)) * 8;
        //int Num15 = int.Parse(_id.Substring(15, 1)) * 4;
        //int Num16 = int.Parse(_id.Substring(16, 1)) * 2;
        int allNum = Num0 + Num1 + Num2 + Num3 + Num4 + Num5 + Num6 + Num7 + Num8 + Num9 + Num10 + Num11 + Num12 + Num13 + Num14 + Num15 + Num16;
        Judge(allNum, _id.Substring(17, 1));
    }

    int JudgePerChar(string id, int index, int Mutli)
    {
        int num = -1;
        if (int.TryParse(_id.Substring(index, 1), out int numTemp))
        {
            num = numTemp * Mutli;
        }
        else
        {
            _Output.text = "错误的身份证格式";
        }
        return num;
    }
    /// <summary>
    /// 计算身份证号码是否合法
    /// </summary>
    /// <param 前17位相加之和="allNum"></param>
    /// <param 身份证号码最后一位="LastNum"></param>
    void Judge(int allNum, string LastNum)
    {
        int Remainder = allNum % 11;
        //如果最后一位数不是X的时候将最后一位数转换为int
        //共有以下十种情况
        JudgeQualified(Remainder, 0, LastNum, 1);
        JudgeQualified(Remainder, 1, LastNum, 0);
        JudgeQualified(Remainder, 2, LastNum, "x".GetHashCode(), true);
        JudgeQualified(Remainder, 3, LastNum, 9);
        JudgeQualified(Remainder, 4, LastNum, 8);
        JudgeQualified(Remainder, 5, LastNum, 7);
        JudgeQualified(Remainder, 6, LastNum, 6);
        JudgeQualified(Remainder, 7, LastNum, 5);
        JudgeQualified(Remainder, 8, LastNum, 4);
        JudgeQualified(Remainder, 9, LastNum, 3);
        JudgeQualified(Remainder, 10, LastNum, 2);
        #region 重构前
        //if (Remainder == 0)
        //{
        //    if (int.Parse(LastNum) == 1)
        //    {
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 1)
        //{
        //    if (int.Parse(LastNum) == 0)
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 2)
        //{
        //    Debug.Log("最后一位数是X");
        //    if (LastNum != "x" && LastNum != "X")
        //    {
        //        Debug.Log(LastNum);
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //}
        //if (Remainder == 3)
        //{
        //    if (int.Parse(LastNum) == 9)
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 4)
        //{
        //    if (int.Parse(LastNum) == 8)
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 5)
        //{
        //    if (int.Parse(LastNum) == 7)
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 6)
        //{
        //    if (int.Parse(LastNum) == 6)
        //    {
        //        Debug.Log("身份证合法，已经通过验证");
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Debug.Log("身份证填写错误");
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 7)
        //{
        //    if (int.Parse(LastNum) == 5)
        //    {
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 8)
        //{
        //    if (int.Parse(LastNum) == 4)
        //    {
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 9)
        //{
        //    if (int.Parse(LastNum) == 3)
        //    {
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        //if (Remainder == 10)
        //{
        //    if (int.Parse(LastNum) == 2)
        //    {
        //        Output.text = "恭喜验证通过";
        //    }
        //    else
        //    {
        //        Output.text = "您的身份证号不码合法";
        //    }
        //}
        #endregion
    }
    /// <summary>
    /// 检查身份证是否合法
    /// </summary>
    void JudgeQualified(int remind, int condition, string lastNum, int contrast, bool isX = false)
    {
        if (remind == condition)
        {
            if (isX)
            {
                if (lastNum == "x" || lastNum == "X")
                {
                    //暂时发送认证成功的消息到服务器
                    //TODO接入实名认证系统进行验证身份
                    //发消息到服务器验证是否重复,不重复就认证成功，收集信息
                    Verification.Instance.RequestVerifyStatus(_id);
                }
                else
                {
                    _Output.text = $"您的身份证号码不合法";
#if WINGJOY_TEST
                    _Output.text = $"您的身份证号码不合法,最后一位应该是:X";
#endif
                }
            }
            else
            {
                //最后一位和正常的对比
                if (int.Parse(lastNum) == contrast)
                {
                    //暂时发送认证成功的消息到服务器
                    //TODO接入实名认证系统进行验证身份
                    //发消息到服务器验证是否重复
                    Verification.Instance.RequestVerifyStatus(_id);
                }
                else
                {
                    _Output.text = $"您的身份证号码不合法";
#if WINGJOY_TEST
                    _Output.text = $"您的身份证号码不合法,最后一位应该是:{contrast}";
#endif
                }
            }
        }
    }

    /// <summary>
    /// 判断年龄是否是未成年,暂时在客户端做，后面在服务器做
    /// </summary>
    int JudgeAge(int year, int month)
    {
        var nowY = DateTime.Now.Year;
        var nowMonth = DateTime.Now.Month;
        var offY = nowY - year;
        var offM = nowMonth - month;
        if (offM < 0)
        {
            offM = nowMonth + 12 - month;
            offY = offY - 1;
        }
        if (offY >= 18)
        {
            _Output.text += "\n您已成年";
            return 2;
        }
        return 1;
    }


    void CheckFont() {
        if (!string.IsNullOrEmpty(_InputName.text))
        {
            //Debug.LogError($"是否包含该敏感词:{_InputName.text},{IllegalWordDetection.Instance.strings.Contains(_InputName.text)}");
            int length = _InputName.text.Length;
            _InputName.text = /*IllegalWordDetection.Filter(_InputName.text);*/IllegalWordDetection.Filter(_InputName.text);
            if (_InputName.text.Contains("*"))
            {
                _stringBuilder.Clear();
                for (int i = 0; i < length; i++)
                {
                    _stringBuilder.Append("*");
                }
                _InputName.text = _stringBuilder.ToString();
            }
            _InputName.textComponent.text = _InputName.text;
            //Debug.Log($"输入框文本{ _InputName.text},显示文本:{_InputName.textComponent.text}");
            Canvas.ForceUpdateCanvases();
        }
    }
}
