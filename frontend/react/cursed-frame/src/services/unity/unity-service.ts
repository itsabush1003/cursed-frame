import { useUnityContext } from "react-unity-webgl";

import { pathReplaceRegex } from "@/utils/util";

type unityContextTypes = ReturnType<typeof useUnityContext>;
export type eventTrigger = () => void;
export type projectionMode = "Landscape" | "Portlate";

const unityMessageReceiveObject = "MessageLiaison";
const getUnityService = (
  sendMessage: unityContextTypes["sendMessage"],
  addEventListener: unityContextTypes["addEventListener"],
  removeEventListener: unityContextTypes["removeEventListener"],
) => {
  const imageEndpoint = window.location.pathname.replace(
    pathReplaceRegex,
    "/rest/images/",
  );
  const registEvent = (event: string, callback: eventTrigger) => {
    addEventListener(event, callback);
    return () => removeEventListener(event, callback);
  };
  const unityService = {
    toUnity: {
      SetCameraProjection: (mode: projectionMode) =>
        sendMessage(unityMessageReceiveObject, "SetCameraProjection", mode),
      SetAudioEnable: (isEnable: boolean) =>
        sendMessage(
          unityMessageReceiveObject,
          "SetAudioEnable",
          Number(isEnable),
        ),
      SetAccessToken: (token: string) =>
        sendMessage(unityMessageReceiveObject, "SetAccessToken", token),
      StartQuest: (teamId: number, teamNum: number) =>
        sendMessage(
          unityMessageReceiveObject,
          "StartQuestScene",
          JSON.stringify({ teamId: teamId, teamNum: teamNum }),
        ),
      ChangeTargetTeam: (teamId: number) =>
        sendMessage(unityMessageReceiveObject, "ChangeTargetTeam", teamId),
      SetNextTarget: (targetImageId: string) =>
        sendMessage(
          unityMessageReceiveObject,
          "SetNextTexture",
          imageEndpoint + targetImageId,
        ),
      StartAttackAnimation: (isCorrect: boolean[]) =>
        sendMessage(
          unityMessageReceiveObject,
          "StartAttackAnimation",
          JSON.stringify({ boolList: isCorrect }),
        ),
    },
    fromUnity: {
      RegistSplashDone: (callback: eventTrigger) =>
        registEvent("SplashDone", callback),
      RegistQuestSceneReady: (callback: eventTrigger) =>
        registEvent("QuestSceneReady", callback),
      RegistTeamChanged: (callback: eventTrigger) =>
        registEvent("TargetTeamChanged", callback),
      RegistTargetChanged: (callback: eventTrigger) =>
        registEvent("TargetMemberChanged", callback),
      RegistAttackEnd: (callback: eventTrigger) =>
        registEvent("AttackAnimationEnd", callback),
    },
  };

  return unityService;
};

export default getUnityService;
