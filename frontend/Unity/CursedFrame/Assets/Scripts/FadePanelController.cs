using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

using DG.Tweening;

public class FadePanelController : MonoBehaviour
{

    private static FadePanelController instance;

   [SerializeField] private Image fadeImage = null;

    void Awake()
    {
        if (instance != null)
        {
            Destroy(gameObject);
            return;
        }

        instance = this;
        DontDestroyOnLoad(gameObject);
    }

    /// <summary>
    /// Fadeを実行する関数
    /// </summary>
    /// <param name="targetAlpha">到達したい透明度　0fなら透明、1fなら真っ黒</param>
    /// <param name="duration">Fadeにかける時間</param>
    /// <param name="callback">Fade終了後に呼ばれるコールバック関数</param>
    /// <returns>FadeアニメーションのTweener</returns>
    public Tweener Fade(float targetAlpha, float duration, System.Action callback = null)
    {
        return fadeImage.DOFade(targetAlpha, duration).OnComplete(() => callback?.Invoke());
    }
}
