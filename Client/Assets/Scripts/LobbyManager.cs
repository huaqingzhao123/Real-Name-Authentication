 using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class LobbyManager : MonoBehaviour
{
    private void Awake()
    {
        //解析敏感词
        TextAsset asset = Resources.Load<TextAsset>("敏感词");
        IllegalWordDetection.Instance.ReadText(asset);
    }
    IEnumerator Start()
    {
        yield return new WaitUntil(() => Verification.Instance.Times != 1);
        if (Verification.Instance.Age > 0 && Verification.Instance.Age < 18)
        {
            AntiAddictionManager.Instance.ShowTips();
            //transform.Find("YoungText").gameObject.SetActive(true);
        }
        //未成年管理
        //YoungManger();
    }
}
