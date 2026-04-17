using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using DG.Tweening;

public class MagiciansController : MonoBehaviour
{
    public const int MAX_CHARA_NUM = 9;

    [SerializeField] private GameObject[] magicians = new GameObject[MAX_CHARA_NUM];
    [SerializeField] private GameObject[] bullets = new GameObject[MAX_CHARA_NUM];
    [SerializeField] private Animator[] animators = new Animator[MAX_CHARA_NUM];
    [SerializeField] private ParticleSystem particleSystem;

    private MaterialPropertyBlock propertyBlock = null;
    private readonly Vector3 positionOffset = new Vector3(-0.5f, 0.0f, 0.0f);
    private readonly Vector3 basePosition = new Vector3(0.0f, 0.0f, 2.0f);
    private readonly Color[] colors = new Color[MAX_CHARA_NUM]
    {
        Color.red,
        Color.blue,
        Color.green,
        Color.yellow,
        Color.cyan,
        Color.magenta,
        Color.white,
        Color.gray,
        Color.black
    };

    void Awake()
    {
        propertyBlock  = new MaterialPropertyBlock();
        particleSystem = GetComponent<ParticleSystem>();
        for (var i = 0; i < magicians.Length; i++)
        {
            bullets[i] = magicians[i].transform.Find("Bullet").gameObject;
            animators[i] = magicians[i].GetComponent<Animator>();
        }
    }

    /// <summary>
    /// キャラクタのGameObjectを取得する関数
    /// </summary>
    /// <param name="teamId">何番目のキャラクタか</param>
    /// <returns>該当するGameObject</returns>
    public GameObject GetMagician(int teamId)
    {
        return magicians[teamId-1];
    }

    /// <summary>
    /// クエスト開始時のアニメーションを再生する関数
    /// </summary>
    /// <param name="teamId">何番目のキャラクタを表示するか　複数体表示したい場合は0を入れる</param>
    /// <param name="teamNum">teamId==0のとき、全部で何体のキャラを表示するか</param>
    /// <param name="target">キャラクタが視線を向ける対象</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void StartPrepareAnimation(int teamId, int teamNum, Transform target, System.Action callback = null)
    {
        if (teamId == 0)
        {
            for (int i = 0; i < magicians.Length && i < teamNum; i++)
            {
                if (teamNum % 2 == 0) magicians[i].transform.localPosition += positionOffset;
                SetupMagician(magicians[i], target, i < colors.Length ? colors[i] : Color.clear);
                animators[i].SetBool("isRunning", true);
            }
            MoveToBase().OnComplete(() => {
                for (int i = 0; i < magicians.Length && i < teamNum; i++)
                {
                    animators[i].SetBool("isRunning", false);
                }
                particleSystem.Stop();
                callback?.Invoke();
            });
        }
        else if (0 < teamId && teamId <= magicians.Length)
        {
            magicians[teamId-1].transform.localPosition /= 2;
            var sh = particleSystem.shape;
            sh.position += magicians[teamId-1].transform.localPosition;
            sh.scale = Vector3.one;
            float angle = Mathf.Atan2(magicians[teamId-1].transform.localPosition.x, basePosition.z) * Mathf.Rad2Deg;
            SetupMagician(magicians[teamId-1], target, colors[teamId-1]);
            animators[teamId-1].SetBool("isRunning", true);
            DOTween.Sequence()
                .Append(MoveToBase())
                .Join(magicians[teamId-1].transform.DOLookAt(target.position, 2.0f).SetDelay(2.0f))
                .OnComplete(() => {
                    animators[teamId-1].SetBool("isRunning", false);
                    particleSystem.Stop();
                    callback?.Invoke();
                });
        }
    }

    /// <summary>
    /// 攻撃用アニメーションを再生する関数
    /// </summary>
    /// <param name="teamId">アニメーションを適用するチームのID</param>
    /// <param name="unAttackTeamId">攻撃できないチームのID</param>
    /// <param name="targetPosition">攻撃を当てるターゲットのワールド座標</param>
    /// <param name="isSuccesses">攻撃が成功するかどうかが入った配列　teamId==0の時はチーム数分、1<=teamId<=teamNumの時は一つだけ</param>
    /// <param name="createHitMotion">攻撃が当たった事を表現するアニメーションを生成する関数</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void StartAttackAnimation(int teamId, int unAttackTeamId, Vector3 targetPosition, bool[] isSuccesses, System.Func<Tweener> createHitMotion, System.Action callback = null)
    {
        Sequence sequence = DOTween.Sequence();
        if (teamId == 0)
        {
            for (var i = 0; i < bullets.Length; i++)
            {
                if (i+1 == unAttackTeamId) continue;
                sequence.Join(SetupBulletAnimation(bullets[i], isSuccesses[i] ? targetPosition : transform.position));
            }
            sequence.Insert(3.0f, createHitMotion());
        }
        else if (0 < teamId && teamId <= MAX_CHARA_NUM && teamId != unAttackTeamId)
        {
            sequence.Append(SetupBulletAnimation(bullets[teamId-1], isSuccesses[0] ? targetPosition : transform.position).Join(createHitMotion()));
        }
        else
        {
            // 攻撃できないチームも他のチームとcallbackの実行タイミングを合わせる
            sequence.AppendInterval(4.0f);
        }
        sequence.OnComplete(() => callback?.Invoke());
    }

    /// <summary>
    /// キャラクタの初期設定をする関数
    /// </summary>
    /// <param name="magician">初期設定するキャラクタのGameObject</param>
    /// <param name="target">GameObjectが向くべき対象</param>
    /// <param name="color">モデルの色</param>
    private void SetupMagician(GameObject magician, Transform target, Color color)
    {
        magician.SetActive(true);
        magician.transform.LookAt(target);
        Renderer[] renderers = magician.GetComponentsInChildren<Renderer>();
        if (renderers != null)
        {
            foreach (Renderer renderer in renderers)
            {  
                // 弾は最初ActiveでないとGetComponentsに引っかからないので、取得後に非表示にする
                if (renderer.gameObject.name == "Energy") renderer.transform.parent.gameObject.SetActive(false);
                renderer.GetPropertyBlock(propertyBlock);
                propertyBlock.SetColor("_Color", color);
                // 完全な黒だと弾が見えないので、少しだけ色を付ける
                // それ以外の色の場合は、透明度を半分にする
                propertyBlock.SetColor("_TintColor", color == Color.black ? new Color(0.1f, 0.1f, 0.1f, 0.5f) : color - new Color(0.0f, 0.0f, 0.0f, 0.5f));
                renderer.SetPropertyBlock(propertyBlock);
            }
        }
    }

    /// <summary>
    /// bulletのアニメーションを再生する関数
    /// </summary>
    /// <param name="bullet">bulletのGameObject</param>
    /// <param name="target">bulletを当てるワールド座標</param>
    /// <returns>bulletのアニメーションのSequence</returns>
    private Sequence SetupBulletAnimation(GameObject bullet, Vector3 target)
    {
        Vector3 originalPosition = bullet.transform.position;
        Vector3 originalScale = bullet.transform.localScale;
        Renderer energyRenderer = bullet.transform.Find("Energy").gameObject.GetComponent<Renderer>();
        Color originalColor = GetColorWithPB(energyRenderer);
        bullet.SetActive(true);
        return DOTween.Sequence()
            .AppendInterval(1.0f)
            .Append(bullet.transform.DOMove(target, 2.0f))
            .Append(DOTween.To(() => originalColor.a, x => SetColorWithPB(energyRenderer, new Color(originalColor.r, originalColor.g, originalColor.b, x)), 0.0f, 1.0f).SetEase(Ease.OutExpo))
            // .Append(energyRenderer.materal.DOFade(0.0f, "_TintColor", 1.0f).SetEase(Ease.OutExpo))
            .Join(bullet.transform.DOScale(new Vector3(2.0f, 2.0f, 1.0f), 1.0f).SetEase(Ease.OutExpo))
            .OnComplete(() =>
            {
                bullet.SetActive(false);
                bullet.transform.position = originalPosition;
                bullet.transform.localScale = originalScale;
                SetColorWithPB(energyRenderer, originalColor);
            });
    }

    /// <summary>
    /// MaterialPropertyBlock経由で色を取得する関数
    /// </summary>
    /// <param name="renderer">色を取得したいマテリアルを持つRenderer</param>
    /// <param name="property">プロパティ名</param>
    /// <returns>プロパティに設定されているColor</returns>
    private Color GetColorWithPB(Renderer renderer, string property = "_TintColor")
    {
        renderer.GetPropertyBlock(propertyBlock);
        Color color = propertyBlock.GetColor(property);
        renderer.SetPropertyBlock(propertyBlock);
        return color;
    }

    /// <summary>
    /// MaterialPropertyBlock経由で色を設定する関数
    /// </summary>
    /// <param name="renderer">色を設定したいマテリアルを持つRenderer</param>
    /// <param name="color">設定したい色</param>
    /// <param name="property">プロパティ名</param>
    private void SetColorWithPB(Renderer renderer, Color color, string property = "_TintColor")
    {
        renderer.GetPropertyBlock(propertyBlock);
        propertyBlock.SetColor(property, color);
        renderer.SetPropertyBlock(propertyBlock);
    }

    /// <summary>
    /// キャラクタを初期位置に移動するアニメーションを返す（再生する）関数
    /// </summary>
    /// <returns>Tweenアニメーション</returns>
    private Tweener MoveToBase()
    {
        return transform.DOMove(basePosition, 4.0f);
    }
}
