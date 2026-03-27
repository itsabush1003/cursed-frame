using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using DG.Tweening;

public class MagiciansController : MonoBehaviour
{
    public const int MAX_CHARA_NUM = 9;

    [SerializeField] private GameObject[] magicians = new GameObject[MAX_CHARA_NUM];
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
    }

    // Start is called before the first frame update
    void Start()
    {
        
    }

    // Update is called once per frame
    void Update()
    {
        
    }

    /// <summary>
    /// キャラクタのGameObjectを取得する関数
    /// </summary>
    /// <param name="teamId">何番目のキャラクタか</param>
    /// <returns></returns>
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
    /// <param name="callback">コールバック関数</param>
    public void StartPrepareAnimation(int teamId, int teamNum, Transform target, System.Action callback = null)
    {
        if (teamId == 0)
        {
            for (int i = 0; i < magicians.Length && i < teamNum; i++)
            {
                if (teamNum % 2 == 0) magicians[i].transform.localPosition += positionOffset;
                SetupMagician(magicians[i], target, i < colors.Length ? colors[i] : Color.clear);
            }
            moveToBase().OnComplete(() => {
                particleSystem.Stop();
                if (callback != null) callback();
            });
        }
        else if (0 < teamId && teamId <= magicians.Length)
        {
            magicians[teamId-1].transform.localPosition /= 2;
            var sh = particleSystem.shape;
            sh.position += magicians[teamId-1].transform.localPosition;
            sh.scale = Vector3.right;
            float angle = Mathf.Atan2(magicians[teamId-1].transform.localPosition.x, basePosition.z) * Mathf.Rad2Deg;
            SetupMagician(magicians[teamId-1], target, colors[teamId-1]);
            DOTween.Sequence()
                .Append(moveToBase())
                .Join(magicians[teamId-1].transform.DOLookAt(target.position, 2.0f).SetDelay(2.0f))
                .OnComplete(() => {
                    particleSystem.Stop();
                    if (callback != null) callback();
                });
        }
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
                renderer.GetPropertyBlock(propertyBlock);
                propertyBlock.SetColor("_Color", color);
                renderer.SetPropertyBlock(propertyBlock);
            }
        }
    }

    /// <summary>
    /// キャラクタを初期位置に移動するアニメーションを返す（再生する）関数
    /// </summary>
    /// <returns>Tweenアニメーション</returns>
    private Tweener moveToBase()
    {
        return transform.DOMove(basePosition, 4.0f);
    }
}
