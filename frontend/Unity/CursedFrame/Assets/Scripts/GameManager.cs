using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using DG.Tweening;

public class GameManager : MonoBehaviour
{
    [SerializeField] private GameObject enemyObject = null;
    [SerializeField] private Material profileImageMaterial = null;
    [SerializeField] private CameraController cameraController;
    [SerializeField] private MagiciansController magiciansController;
    [SerializeField] private int debugTeamId = 0;
    private MessageLiaison liaison = null;
    private int currentTargetTeamId = -1;

    // Start is called before the first frame update
    void Start()
    {
        bool alreadyPrepared = liaison != null;
        liaison = GameObject.FindWithTag("MessageLiaison")?.GetComponent<MessageLiaison>();
        enemyObject.transform.DOMove(Vector3.up * 0.3f, 2.0f).SetRelative(true).SetEase(Ease.InOutSine).SetLoops(-1, LoopType.Yoyo);
# if UNITY_EDITOR
        if (!alreadyPrepared) StartPrepareAnimation();
# endif
    }

    // Update is called once per frame
    void Update()
    {
# if UNITY_EDITOR
        if(Input.GetMouseButtonDown(0))
        {
            cameraController.StartAttentionAnimation();
        }
        if(Input.GetMouseButtonDown(1))
        {
            ChangeTargetTeam(4);
        }
        if(Input.GetMouseButtonDown(2))
        {
            StartAttackAnimation(new bool[] {false, false, true, false, true, true, true, false, true});
        }
# endif
    }

    /// <summary>
    /// MessageLiaisonが自身をGameManagerに登録するための関数
    /// OnSceneLoadedがStartより前に実行されてしまうので、その対策
    /// </summary>
    /// <param name="messageLiaison">MessageLiaisonのインスタンス</param>
    public void RegistMessageLiaison(MessageLiaison messageLiaison)
    {
        liaison = messageLiaison;
    }

    /// <summary>
    /// 開始準備のアニメーションを再生する関数
    /// </summary>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void StartPrepareAnimation(Action callback = null)
    {
        enemyObject.transform.DOPause();
        Transform target = enemyObject.transform.Find("EnemyGroundPosition");
        int teamId = liaison?.GetTeamId() ?? debugTeamId;
        int teamNum = liaison?.GetTeamNum() ?? MagiciansController.MAX_CHARA_NUM;
        magiciansController.StartPrepareAnimation(teamId, teamNum, target, () =>
        {
            if (0 < teamId && teamId <= MagiciansController.MAX_CHARA_NUM)
            {
                enemyObject.transform.LookAt(magiciansController.GetMagician(teamId).transform.position + Vector3.up * enemyObject.transform.position.y);
            }
            enemyObject.transform.DOPlay();
        });
        if (teamId == 0)
        {
            cameraController.StartPrepareAnimation(CameraController.TweenTarget.Admin, 0.0f, callback);
        }
        else if (0 < teamId && teamId <= MagiciansController.MAX_CHARA_NUM)
        {
            float angle = Mathf.Atan2(magiciansController.GetMagician(teamId).transform.localPosition.x, 2.0f) * Mathf.Rad2Deg;
            cameraController.StartPrepareAnimation(CameraController.TweenTarget.Guest, angle, callback);
        }
    }

    /// <summary>
    /// 次のターゲットを表示・強調するアニメーションを再生する関数
    /// </summary>
    /// <param name="texture">Enemyに読み込ませる画像用テクスチャ</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void SetNextTarget(Texture2D texture, Action callback)
    {
        enemyObject.transform.DOPause();
        ChangeEnemyTexture(texture);
        cameraController.StartAttentionAnimation(() => {
            enemyObject.transform.DOPlay();
            callback?.Invoke();
        });
    }

    /// <summary>
    /// ターゲットとなるチームが変わった事を示すアニメーションを再生する関数
    /// </summary>
    /// <param name="teamId">次のターゲットになるチームのID</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void ChangeTargetTeam(int teamId, Action callback = null)
    {
        int myTeamId = liaison?.GetTeamId() ?? debugTeamId;
        GameObject target = magiciansController.GetMagician(teamId);
        enemyObject.transform.DOPause();
        Transform originalPosition = enemyObject.transform;
        Sequence sequence = DOTween.Sequence()
            .Append(enemyObject.transform.DOShakeRotation(1.0f, new Vector3(5.0f, 5.0f, 30.0f), 10, 90.0f));
        if (currentTargetTeamId == myTeamId)
        {
            sequence.Join(cameraController.ReleaseAnimation());
        }
        sequence.AppendInterval(1.0f)
            .Append(enemyObject.transform.DOLookAt(target.transform.position + Vector3.up * enemyObject.transform.position.y, 1.0f))
            .Append(enemyObject.transform.DOMove(target.transform.position + Vector3.right * enemyObject.transform.localScale.x / 2 * Mathf.Sign(target.transform.localPosition.x) * -1 + Vector3.up + enemyObject.transform.forward, 3.0f).SetEase(Ease.InOutBack, 2.0f))
            .AppendInterval(0.5f);
        if (teamId == myTeamId)
        {
            sequence.Join(cameraController.CaptureAnimation(originalPosition.position));
        }
        sequence.Append(enemyObject.transform.DOMove(originalPosition.position, 3.0f))
            .Join(enemyObject.transform.DORotateQuaternion(originalPosition.rotation, 3.0f))
            .OnComplete(() => {
                currentTargetTeamId = teamId;
                enemyObject.transform.DOPlay();
                callback?.Invoke();
            });
    }

    /// <summary>
    /// 攻撃アニメーションを再生する関数
    /// </summary>
    /// <param name="isCorrects">各チームの正解・不正解が入った配列　個人用の場合はそのチームの要素一つだけ</param>
    /// <param name="callback">アニメーション終了時に呼ばれるコールバック関数</param>
    public void StartAttackAnimation(bool[] isCorrects, Action callback = null)
    {
        magiciansController.StartAttackAnimation(liaison?.GetTeamId() ?? debugTeamId, currentTargetTeamId, enemyObject.transform.position + Vector3.right, isCorrects, callback);
    }

    /// <summary>
    /// Enemyの画像テクスチャを指定したテクスチャに変更する関数
    /// </summary>
    /// <param name="tex">読み込む画像テクスチャ</param>
    private void ChangeEnemyTexture(Texture2D tex)
    {
        profileImageMaterial.mainTexture = tex;
        profileImageMaterial.color = Color.white;
    }
}