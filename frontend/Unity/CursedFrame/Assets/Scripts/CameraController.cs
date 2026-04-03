using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using DG.Tweening;

public class CameraController : MonoBehaviour
{
    public enum TweenTarget
    {
        Admin,
        Guest
    }
    public enum TweenType
    {
        Move,
        Rotate,
        Scale,
        LocalMove,
        LocalMoveX,
        LocalMoveY,
        LocalMoveZ
    }
    [System.Serializable]
    private class TweenParams
    {
        public Vector3 endState;
        public float duration;
        public TweenType type;
        public Ease ease = Ease.OutQuad;
        public Tweener buildTweener(Transform transform)
        {
            switch (type)
            {
                case TweenType.Move:
                    return transform.DOMove(endState, duration).SetEase(ease);
                case TweenType.Rotate:
                    return transform.DORotate(endState, duration).SetEase(ease);
                case TweenType.Scale:
                    return transform.DOScale(endState, duration).SetEase(ease);
                case TweenType.LocalMove:
                    return transform.DOLocalMove(endState, duration).SetEase(ease).SetRelative(true);
                case TweenType.LocalMoveX:
                    return transform.DOLocalMoveX(endState.x, duration).SetEase(ease).SetRelative(true);
                case TweenType.LocalMoveY:
                    return transform.DOLocalMoveY(endState.y, duration).SetEase(ease).SetRelative(true);
                case TweenType.LocalMoveZ:
                    return transform.DOLocalMoveZ(endState.z, duration).SetEase(ease).SetRelative(true);
                default:
                    throw new System.NotImplementedException("Unimplemented type");
            }
        }
    }

    [System.Serializable]
    private class TweenParamsList
    {
        public TweenParams[] paramList;
    }

    [SerializeField] private Camera camera;
    [SerializeField] private TweenParamsList[] adminTweens;
    [SerializeField] private TweenParamsList[] guestTweens;
    [SerializeField] private TweenParams[] attentionEnemyTweens;

    private FadePanelController fadeController = null;
    private Vector3 basePosition = Vector3.zero;
    private Quaternion baseRotation = Quaternion.identity;
    
    void Awake()
    {
        camera = GetComponent<Camera>();
    }

    // Start is called before the first frame update
    void Start()
    {
        fadeController = GameObject.FindWithTag("Canvas").GetComponent<FadePanelController>();
    }

    // Update is called once per frame
    void Update()
    {
        
    }

    /// <summary>
    /// クエスト開始時のアニメーションを再生する関数
    /// </summary>
    /// <param name="target">"Admin"（全員写す）か"Guest"（一人だけ移す）か</param>
    /// <param name="targetAngle">写したいtargetと原点との角度（Degree）</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    /// <exception cref="System.NotImplementedException"></exception>
    public void StartPrepareAnimation(TweenTarget target, float targetAngle = 0.0f, System.Action callback = null)
    {
        TweenParamsList[] targetTweens;
        switch (target)
        {
            case TweenTarget.Admin:
                targetTweens = adminTweens;
                break;
            case TweenTarget.Guest:
                targetTweens = guestTweens;
                break;
            default:
                throw new System.NotImplementedException("Unimplemented type");
        }
        // 写したいtargetと原点を結ぶ線の延長線上まで移動するための距離
        float offset = targetTweens[0].paramList[targetTweens[0].paramList.Length-1].endState.z * Mathf.Tan(targetAngle * Mathf.Deg2Rad);
        // 左右に離れるほどtargetとの距離が長くなるので、絵面を近づける為にその距離を補正する為の値　本来はtargetとの距離を掛ける必要があるが、一旦この簡易版で
        float adjuster = 1 / Mathf.Cos(targetAngle * Mathf.Deg2Rad) - 1f;
        targetTweens[0].paramList[targetTweens[0].paramList.Length-1].endState += Vector3.right * offset;
        targetTweens[1].paramList[targetTweens[1].paramList.Length-1].endState += Vector3.up * targetAngle;
        Sequence sequence = DOTween.Sequence();
        foreach (var tweenList in targetTweens)
        {
            Sequence subSequence = DOTween.Sequence();
            foreach (var tween in tweenList.paramList)
            {
                subSequence.Append(tween.buildTweener(transform));
            }
            sequence.Join(subSequence);
        }
        sequence.OnComplete(() => {
            transform.DOLocalMove(transform.forward * adjuster, 1f).SetEase(Ease.Linear).SetRelative(true).OnComplete(() => callback?.Invoke());
        });
    }

    /// <summary>
    /// Enemyに表示された画像をプレイヤーに見せるために
    /// Enemyに向かって近づいていくアニメーションを再生する関数
    /// </summary>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void StartAttentionAnimation(System.Action callback = null)
    {
        // Enemyの原点がちょっとずれてる影響か、x < 0の位置からEnemyの方向へ進むと左側に見切れてしまので、それを補正する為の係数
        float adjuster = transform.position.x < 0 ? 1.7f : 1.0f;
        Sequence sequence = DOTween.Sequence();
        foreach (var tween in attentionEnemyTweens)
        {
            sequence.Join(tween.buildTweener(transform));
        }
        sequence.Join(transform.DOLocalMoveX(-transform.position.x / 3 * adjuster, 3.0f).SetRelative(true));
        sequence.AppendInterval(0.5f);
        sequence.SetLoops(2, LoopType.Yoyo);
        if (callback != null)
        {
            sequence.OnComplete(() => callback());
        }
    }

    /// <summary>
    /// 囚われるアニメーションを再生する関数
    /// </summary>
    /// <param name="targetPosition">アニメーション終了時の移動後のワールド座標</param>
    /// <returns>アニメーションのSequence</returns>
    public Sequence CaptureAnimation(Vector3 targetPosition)
    {
        basePosition = transform.position;
        baseRotation = transform.rotation;
        return DOTween.Sequence()
            .Append(transform.DOShakeRotation(0.5f, new Vector3(2.0f, 2.0f, 10.0f), 10, 90.0f))
            .Join(fadeController.Fade(1.0f, 0.5f, () => {
                transform.position = targetPosition;
                transform.LookAt(basePosition);
                fadeController.Fade(0.6f, 3.0f);
            }));
    }

    /// <summary>
    /// 解放されるアニメーションを再生する関数
    /// 上のcaptureAnimationが実行される前に実行されるとおかしくなる可能性がある
    /// </summary>
    /// <returns>アニメーションのSequence</returns>
    public Sequence ReleaseAnimation()
    {
        return DOTween.Sequence()
            .Append(transform.DOShakeRotation(1.0f, new Vector3(2.0f, 2.0f, 10.0f), 10, 90.0f)).OnComplete(() =>
            {
                DOTween.Sequence()
                    .AppendInterval(0.5f)
                    .Append(transform.DOMove(basePosition, 0.5f).SetEase(Ease.InQuad))
                    .Join(transform.DORotateQuaternion(baseRotation, 0.5f))
                    .Join(fadeController.Fade(0.0f, 0.5f).SetEase(Ease.InQuad));
            });
    }
}
