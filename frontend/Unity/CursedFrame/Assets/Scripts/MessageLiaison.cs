using System;
using System.Collections;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using UnityEngine;
using UnityEngine.Networking;
using UnityEngine.Rendering;
using UnityEngine.SceneManagement;

public class MessageLiaison : MonoBehaviour
{
    [DllImport("__Internal")]
    private static extern void SplashDone();
    [DllImport("__Internal")]
    private static extern void QuestSceneReady();
    [DllImport("__Internal")]
    private static extern void TargetMemberChanged();
    [DllImport("__Internal")]
    private static extern void TargetTeamChanged();
    [DllImport("__Internal")]
    private static extern void AttackAnimationEnd();
    private static MessageLiaison instance = null;

    private const float landscapeFov = 70.0f;
    private const float portlateFov = 55.0f;
    private const int maxRetry = 3;

    [Serializable]
    private class TeamInfo
    {
        public int teamId;
        public int teamNum;
    }

    [Serializable]
    private class BoolList
    {
        public bool[] boolList;
    }

    private GameObject mainCamera = null;
    private FadePanelController fadeController = null;
    private GameManager gameManager = null;

    private string cameraMode = "Landscape";
    private bool isAudioEnable = true;
    private int teamId = -1;
    private int teamNum = 0;
    private string accessToken = "";

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

    // Start is called before the first frame update
    void Start()
    {
# if !UNITY_EDITOR && UNITY_WEBGL
        WebGLInput.captureAllKeyboardInput = false;
        StartCoroutine(WaitSplashDone());
# endif

        mainCamera = GameObject.FindWithTag("MainCamera");
        fadeController = GameObject.FindWithTag("Canvas").GetComponent<FadePanelController>();
    }

    void OnEnable()
    {
        SceneManager.sceneLoaded += OnSceneLoaded;
    }

    void OnDisable()
    {
        SceneManager.sceneLoaded -= OnSceneLoaded;
    }

    // Update is called once per frame
    void Update()
    {
# if UNITY_EDITOR
        if(Input.GetMouseButtonDown(0))
        {
            if (gameManager == null)
            {
                StartQuestScene(JsonUtility.ToJson(new TeamInfo { teamId = 4, teamNum = 6 }));
            }
        }
# endif
    }

    public void SetCameraProjection(string mode)
    {
        Camera camera = mainCamera.GetComponent<Camera>();
        switch (mode)
        {
            case "Landscape":
                camera.fieldOfView = landscapeFov;
                break;
            case "Portlate":
                camera.fieldOfView = portlateFov;
                break;
            default:
                return;
        }
        cameraMode = mode;
    }

    public void SetAudioEnable(int isEnable)
    {
        _SetAudioEnable(isEnable == 1);
    }
    private void _SetAudioEnable(bool isEnable)
    {
        AudioListener audioListener = mainCamera.GetComponent<AudioListener>();
        if (audioListener != null)
        {
            audioListener.enabled = isEnable;
        }
    }

    private void SetTeamInfo(string teamInfo)
    {
        TeamInfo _teamInfo = JsonUtility.FromJson<TeamInfo>(teamInfo);
        _SetTeamInfo(_teamInfo.teamId, _teamInfo.teamNum);
    }
    private void _SetTeamInfo(int _teamId, int _teamNum)
    {
        teamId = _teamId;
        teamNum = _teamNum;
    }
    public int GetTeamId()
    {
        return teamId;
    }
    public int GetTeamNum()
    {
        return teamNum;
    }

    public void SetAccessToken(string token)
    {
        accessToken = token;
    }

    public void StartQuestScene(string teamInfoJson)
    {
        SetTeamInfo(teamInfoJson);
        fadeController.Fade(1.0f, 1.0f, () => {
            SceneManager.LoadScene("QuestScene");
        });
    }

    public void ChangeTargetTeam(int teamId)
    {
        gameManager.ChangeTargetTeam(teamId, TargetTeamChanged);
    }

    public void SetNextTexture(string textureURL)
    {
        StartCoroutine(GetTexture(textureURL, tex => gameManager.SetNextTarget(tex, TargetMemberChanged) ));
    }

    public void StartAttackAnimation(string resultJson)
    {
        BoolList results = JsonUtility.FromJson<BoolList>(resultJson);
        _StartAttackAnimation(results.boolList);
    }

    private void _StartAttackAnimation(bool[] isCorrects)
    {
        gameManager.StartAttackAnimation(isCorrects, AttackAnimationEnd);
    }

    private IEnumerator WaitSplashDone()
    {
        while (!SplashScreen.isFinished)
        {
            yield return null;
        }
        SplashDone();
    }

    private void OnSceneLoaded(Scene scene, LoadSceneMode mode)
    {
        mainCamera = GameObject.FindWithTag("MainCamera");
        if (mainCamera != null)
        {
            SetCameraProjection(cameraMode);
            _SetAudioEnable(isAudioEnable);
        }
        if (scene.name == "QuestScene")
        {
            gameManager = GameObject.FindWithTag("GameController").GetComponent<GameManager>();
            fadeController.Fade(0.5f, 0.5f, () => {
                if (gameManager != null)
                {
                    gameManager.RegistMessageLiaison(this);
                    gameManager.StartPrepareAnimation(QuestSceneReady);
                }
                fadeController.Fade(0.0f, 0.5f);
            });
        }
    }

    private IEnumerator GetTexture(string textureURL, Action<Texture2D> callback)
    {
        for (int i = 0; i < maxRetry; i++)
        {
            using (UnityWebRequest uwr = UnityWebRequestTexture.GetTexture(textureURL))
            {
                uwr.SetRequestHeader("Authorization", $"Bearer {accessToken}");
                yield return uwr.SendWebRequest();

                if (uwr.result == UnityWebRequest.Result.Success)
                {
                    Texture2D texture = DownloadHandlerTexture.GetContent(uwr);
                    texture.wrapMode = TextureWrapMode.Clamp;
                    callback?.Invoke(texture);
                    yield break;
                }
                else
                {
                    Debug.Log($"Texture download failed at {i} time: {uwr.error}");
                    float waitSec = Mathf.Pow(2, i);
                    yield return new WaitForSeconds(waitSec);
                }
            }
        }
    }
}
