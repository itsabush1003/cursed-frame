using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;


public class GameManager : MonoBehaviour
{
    [SerializeField] private GameObject enemyObject = null;
    [SerializeField] private Material profileImageMaterial = null;
    [SerializeField] private CameraController cameraController;
    [SerializeField] private MagiciansController magiciansController;
    [SerializeField] private int debugTeamId = 0;
    private MessageLiaison liaison = null;

    // Start is called before the first frame update
    void Start()
    {
        liaison = GameObject.FindWithTag("MessageLiaison")?.GetComponent<MessageLiaison>();
        Transform target = enemyObject.transform.Find("EnemyGroundPosition");
        int teamId = liaison?.GetTeamId() ?? debugTeamId;
        int teamNum = liaison?.GetTeamNum() ?? MagiciansController.MAX_CHARA_NUM;
        magiciansController.StartPrepareAnimation(teamId, teamNum, target, () =>
        {
            if (0 < teamId && teamId <= MagiciansController.MAX_CHARA_NUM)
            {
                enemyObject.transform.LookAt(magiciansController.GetMagician(teamId).transform.position + Vector3.up * enemyObject.transform.position.y);
            }
        });
        if (teamId == 0)
        {
            cameraController.StartPrepareAnimation(CameraController.TweenTarget.Admin);
        } else if (0 < teamId && teamId <= MagiciansController.MAX_CHARA_NUM)
        {
            float angle = Mathf.Atan2(magiciansController.GetMagician(teamId).transform.localPosition.x, 2.0f) * Mathf.Rad2Deg;
            cameraController.StartPrepareAnimation(CameraController.TweenTarget.Guest, angle);
        }
    }

    // Update is called once per frame
    void Update()
    {
# if UNITY_EDITOR
        if(Input.GetMouseButtonDown(0))
        {
            cameraController.StartAttentionAnimation();
        }
# endif
    }

    /// <summary>
    /// 次のターゲットを表示・強調するアニメーションを再生する関数
    /// </summary>
    /// <param name="texture">Enemyに読み込ませる画像用テクスチャ</param>
    /// <param name="callback">コールバック関数</param>
    public void SetNextTarget(Texture2D texture, Action callback)
    {
        ChangeEnemyTexture(texture);
        cameraController.StartAttentionAnimation(callback);
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