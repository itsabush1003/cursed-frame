import {
  lazy,
  Suspense,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import { css } from "@emotion/react";
import { Unity, useUnityContext } from "react-unity-webgl";

import AudioPermitModal from "@/components/audio-permit-modal.tsx";
import LoadingMark from "@/components/loding-mark.tsx";
import TitleLogo from "@/components/title-logo.tsx";
import { UserStatusContext } from "@/context/user-status-context.ts";
import QuizPage from "@/pages/quiz-page.tsx";
import ResultPage from "@/pages/result-page.tsx";
import TitlePage from "@/pages/title-page.tsx";
import getUnityService from "@/services/unity/unity-service";

const EntryPage = lazy(() => import("@/pages/entry-page.tsx"));
const LobbyViewerPage = lazy(() => import("@/pages/lobby-viewer-page.tsx"));
const ProfilePage = lazy(() => import("@/pages/profile-page.tsx"));

const GameState = {
  Prepare: "PREPARE",
  Title: "TITLE",
  Entry: "ENTRY",
  Ready: "READY",
  Quiz: "QUIZ",
  Result: "RESULT",
} as const;

type GameState = (typeof GameState)[keyof typeof GameState];

const unityWebglFileName = import.meta.env.VITE_UNITY_WEBGL_NAME;

function App() {
  const {
    unityProvider,
    loadingProgression,
    isLoaded,
    sendMessage,
    addEventListener,
    removeEventListener,
  } = useUnityContext({
    loaderUrl: "webgl/" + unityWebglFileName + ".loader.js",
    dataUrl: "webgl/" + unityWebglFileName + ".data.gz",
    frameworkUrl: "webgl/" + unityWebglFileName + ".framework.js.gz",
    codeUrl: "webgl/" + unityWebglFileName + ".wasm.gz",
  });

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isPermitAudioModalVisible, setIsPermitAudioModalVisible] =
    useState<boolean>(true);
  const [isAudioEnable, setAudioEnable] = useState<boolean>(false);
  const [isSplashEnd, setIsSplashEnd] = useState<boolean>(false);
  const [isSceneReady, setIsSceneReady] = useState<boolean>(false);
  const [currentState, setCurrentState] = useState<GameState>(
    GameState.Prepare,
  );
  const { userStatus } = useContext(UserStatusContext);
  const unityService = useMemo(
    () => getUnityService(sendMessage, addEventListener, removeEventListener),
    [sendMessage, addEventListener, removeEventListener],
  );

  if (
    isLoaded &&
    isSplashEnd &&
    !isPermitAudioModalVisible &&
    currentState === GameState.Prepare
  ) {
    unityService.toUnity.SetAudioEnable(isAudioEnable);
    unityService.toUnity.SetCameraProjection(
      userStatus.type === "admin" ? "Landscape" : "Portlate",
    );
    setCurrentState(GameState.Title);
  }
  const calcCanvasSize = useCallback(() => {
    const w = window.innerWidth;
    const h = window.innerHeight;
    // adminの場合は横長16:9で固定
    if (userStatus.type === "admin")
      return { w: w, h: (w * 9) / 16, cw: "100vw", ch: "56.25vw" };
    // guestで横長の画面の場合、その画面に収まる縦長9:16に
    if (w / h > 4 / 3)
      return { w: (h * 9) / 16, h: h, cw: "56.25vh", ch: "100vh" };
    return { w, h, cw: "100vw", ch: "100vh" };
  }, [userStatus.type]);
  const targetCanvasSize = calcCanvasSize();

  useEffect(() => {
    const adjustCanvas = () => {
      if (canvasRef.current) {
        const { w, h } = calcCanvasSize();
        canvasRef.current.width = w;
        canvasRef.current.height = h;
      }
    };
    adjustCanvas();
    window.addEventListener("resize", adjustCanvas);
    const cleanSplashEvent = unityService.fromUnity.RegistSplashDone(() =>
      setIsSplashEnd(true),
    );
    const cleanSceneReadyEvent = unityService.fromUnity.RegistQuestSceneReady(
      () => setIsSceneReady(true),
    );
    return () => {
      window.removeEventListener("resize", adjustCanvas);
      cleanSplashEvent();
      cleanSceneReadyEvent();
    };
  }, [isLoaded, calcCanvasSize, unityService]);

  return (
    <>
      <div>
        <AudioPermitModal
          isVisible={isPermitAudioModalVisible}
          setIsVisible={setIsPermitAudioModalVisible}
          setAudioEnable={setAudioEnable}
        />
      </div>
      <UserStatusContext value={useContext(UserStatusContext)}>
        <div id="unity-webgl-container" style={{ position: "relative" }}>
          {!isLoaded && (
            <p>
              Loading Application... {Math.round(loadingProgression * 100)}%
            </p>
          )}
          <Unity
            unityProvider={unityProvider}
            matchWebGLToCanvasSize={false}
            style={{
              visibility:
                !isPermitAudioModalVisible && isLoaded ? "visible" : "hidden",
              width: targetCanvasSize.cw,
              height: targetCanvasSize.ch,
            }}
            ref={canvasRef}
          />
          {(currentState === GameState.Title ||
            currentState === GameState.Entry) && <TitleLogo />}
          <div
            id="ui-overlay-layer"
            css={uiLayerStyle}
            style={{
              height:
                currentState === GameState.Title ||
                currentState === GameState.Quiz
                  ? "50%"
                  : "90%",
            }}
          >
            {currentState === GameState.Title && (
              <TitlePage toNext={() => setCurrentState(GameState.Entry)} />
            )}
            {currentState === GameState.Entry && (
              <Suspense fallback={<LoadingMark />}>
                {userStatus.type === "admin" ? (
                  <LobbyViewerPage
                    toNext={(teamNum: number) => {
                      unityService.toUnity.SetAccessToken(userStatus.token);
                      unityService.toUnity.StartQuest(
                        userStatus.teamId,
                        teamNum,
                      );
                      setCurrentState(GameState.Quiz);
                    }}
                  />
                ) : (
                  <EntryPage
                    toNext={() => {
                      unityService.toUnity.SetAccessToken(userStatus.token);
                      setCurrentState(GameState.Ready);
                    }}
                  />
                )}
              </Suspense>
            )}
            {currentState === GameState.Ready &&
              (userStatus.type === "admin" ? (
                <></>
              ) : (
                <Suspense fallback={<LoadingMark />}>
                  <ProfilePage
                    toNext={() => {
                      unityService.toUnity.StartQuest(userStatus.teamId, 9);
                      setCurrentState(GameState.Quiz);
                    }}
                  />
                </Suspense>
              ))}
            {currentState === GameState.Quiz && isSceneReady && (
              <Suspense fallback={<LoadingMark />}>
                <QuizPage
                  toNext={() => setCurrentState(GameState.Result)}
                  eventController={{
                    eventSender: unityService.toUnity,
                    registEventReceiver: unityService.fromUnity,
                  }}
                />
              </Suspense>
            )}
            {currentState === GameState.Result && <ResultPage />}
          </div>
        </div>
      </UserStatusContext>
    </>
  );
}

const uiLayerStyle = css`
  position: absolute;
  left: 0;
  right: 0;
  margin: 0 auto;
  width: 90%;
  // height: 50%;
  bottom: 1em;
  z-index: 10;
  align-content: center;
  pointer-events: none;
  overflow-x: hidden;
  overflow-y: visible;
`;

export default App;
