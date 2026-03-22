import { useCallback, useContext, useMemo, useState } from "react";

import RegistProfileForm from "@/components/regist-profile-form";
import TeamInfoPanel from "@/components/team-info-panel";
import { UserStatusContext } from "@/context/user-status-context";
import useStreamObserver from "@/hooks/use-stream-observer";
import getGuestClient from "@/services/rpc/guest-client";

import type { ProfileQuestion } from "@/components/regist-profile-form";
import type { LobbyStatus } from "@/gen/lobby/v1/lobby_pb";

const firstQuestion = {
  questionId: 0,
  question: "次に、いくつかの質問に答えてもらいます。よろしいですね？",
};

const ProfilePage = ({ toNext }: { toNext: () => void }) => {
  const [currentQuestion, setCurrentQuestion] =
    useState<ProfileQuestion>(firstQuestion);
  const [isQuestionComplete, setIsQuestionComplete] = useState<boolean>(false);
  const [isTeamInfoShow, setIsTeamInfoShow] = useState<boolean>(false);
  const [members, setMembers] = useState<string[]>([]);
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const guestClient = useMemo<ReturnType<typeof getGuestClient>>(
    () => getGuestClient(() => userStatus.token),
    [userStatus.token],
  );
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
    <div>
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

export default ProfilePage;
