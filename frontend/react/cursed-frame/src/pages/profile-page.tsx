import { useCallback, useContext, useState } from "react";

import { css } from "@emotion/react";

import RegistProfileForm from "@/components/regist-profile-form";
import TeamInfoPanel from "@/components/team-info-panel";
import { UserStatusContext } from "@/context/user-status-context";
import useRpcClient from "@/hooks/use-rpc-client";
import useStreamObserver from "@/hooks/use-stream-observer";

import type { ProfileQuestion } from "@/components/regist-profile-form";
import type { LobbyStatus } from "@/gen/lobby/v1/lobby_pb";

const firstQuestion = {
  questionId: 0,
  question: [
    "これから、幾つか質問に答えてもらいます。",
    "それらは、あなたが魔法使いと",
    "契約する為に行った交流の記録です。",
    "極めて個人的な内容ですが、",
    "可能な限り真摯にお答えください。",
    "準備ができた方は、ボタンを押してください",
  ].join("\n"),
};

const ProfilePage = ({ toNext }: { toNext: () => void }) => {
  const [currentQuestion, setCurrentQuestion] =
    useState<ProfileQuestion>(firstQuestion);
  const [isQuestionComplete, setIsQuestionComplete] = useState<boolean>(false);
  const [isTeamInfoShow, setIsTeamInfoShow] = useState<boolean>(false);
  const [members, setMembers] = useState<string[]>([]);
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const guestClient = useRpcClient("guest");
  const lobbyStatusStream = useCallback(
    (signal: AbortSignal) => guestClient.joinLobby({ signal: signal }),
    [guestClient],
  );
  const checkLobbyStatus = useCallback(
    (response: LobbyStatus) => response.isAllReady,
    [],
  );
  const isLobbyReady = useStreamObserver(lobbyStatusStream, checkLobbyStatus);
  const registFunc = async (question: ProfileQuestion, answer: string) => {
    const next = await guestClient.registProfile(question.questionId, answer);
    setCurrentQuestion({
      questionId: next.questionId,
      question: next.question,
    });
    if (next.isComplete) {
      await guestClient.isReady();
      setIsQuestionComplete(true);
    }
  };
  const getTeamInfo = async () => {
    const { members, ...teamInfo } = await guestClient.getTeamInfo();
    setUserStatus(teamInfo);
    setMembers(members);
    setIsTeamInfoShow(true);
  };

  return (
    <div css={baseStyle}>
      {!isQuestionComplete &&
        (currentQuestion.questionId === 0 ? (
          <button
            style={{ alignSelf: "center", pointerEvents: "auto" }}
            onClick={async () =>
              await registFunc(currentQuestion, "回答を始める")
            }
          >
            回答を始める
          </button>
        ) : (
          <RegistProfileForm
            currentQuestion={currentQuestion}
            registFunc={registFunc}
          />
        ))}
      {isQuestionComplete &&
        !isTeamInfoShow &&
        (isLobbyReady ? (
          <button style={{ pointerEvents: "auto" }} onClick={getTeamInfo}>
            チームを確認する
          </button>
        ) : (
          <div>{currentQuestion.question}</div>
        ))}
      {isTeamInfoShow && (
        <>
          <TeamInfoPanel
            teamId={userStatus.teamId}
            teamColor={userStatus.color}
            members={members}
          />
          <button onClick={toNext}>ゲーム開始</button>
        </>
      )}
    </div>
  );
};

const baseStyle = css`
  white-space: pre-wrap;
  background-color: var(--sub-color-1-1-light);

  & button {
    pointer-events: auto;
  }
`;

export default ProfilePage;
