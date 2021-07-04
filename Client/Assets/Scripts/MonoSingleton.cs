using System;
using UnityEngine;

/// <summary>
/// 需要使用Unity生命周期的单例模式
/// </summary>

public abstract class MonoSingleton<T> : MonoBehaviour where T : MonoSingleton<T>
{
    protected static T mInstance = null;

    public static T Instance
    {
        get
        {
            if (mInstance == null)
            {
                var instances = FindObjectsOfType<T>();
                if(instances.Length > 0)
                {
                    mInstance = instances[0];
                }
                if (instances.Length > 1)
                {
                    Debug.LogError("More than 1!");
                    return mInstance;
                }

                if (mInstance == null)
                {
                    string instanceName = typeof(T).Name;
                    Debug.Log("Instance Name: " + instanceName);
                    GameObject instanceGO = GameObject.Find(instanceName);

                    if (instanceGO == null)
                        instanceGO = new GameObject(instanceName);

                    mInstance = instanceGO.AddComponent<T>();
                    DontDestroyOnLoad(instanceGO);  //保证实例不会被释放
                    Debug.Log("Add New Singleton " + mInstance.name + " in Game!");
                }
            }

            return mInstance;
        }
    }


    protected virtual void OnDestroy()
    {
        mInstance = null;
    }
}

