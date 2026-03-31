import { useCallback, useContext, useEffect, useRef, useState } from "react";

import { css, keyframes } from "@emotion/react";
import { CSSTransition } from "react-transition-group";

import { type ChoiceRef, Choices, type Choice } from "@/components/choices";
import { QuizTextArea } from "@/components/quiz-text-area";
import {
  TimePressureGage,
  MaxRemainTime,
} from "@/components/time-pressure-gage";
import { UserStatusContext } from "@/context/user-status-context";
import useApiCaller from "@/hooks/use-api-caller";
import useRpcClient from "@/hooks/use-rpc-client.ts";
import useStreamObserver from "@/hooks/use-stream-observer";

import type { StartQuestResponse as adminResponse } from "@/gen/admin/v1/admin_pb";
import type { StartQuestResponse as guestResponse } from "@/gen/quest/v1/quest_pb";

/*const QuizState = {
  FirstMount: "mounted",
  TeamSelectAnimation: "new team's quiz arrived from stream",
  ShowQuizUI: "end select animation or new same team's quiz from stream",
  Thinking: "until remained time from stream response become zero",
  Unanswerable: "remained time become zero",
  CheckAnswer: "response arrived from answer request",
  HideQuizUI: "after intervals from below",
  AttackAnimation: "",
  NextQuiz: "end attack animation",
  End: "stream end"
} as const;*/

interface quiz {
  questionId: number;
  quizText: string;
  choices: Choice[];
  imageId: string;
  teamId: number;
  canAnswer: boolean;
  isTarget: boolean;
}

const uiAnimateDuration: number = 500;

const QuizPage = ({ toNext }: { toNext: () => void }) => {
  const [remainTime, setRemainTime] = useState<number>(MaxRemainTime);
  const [isEnableAnswer, setIsEnableAnswer] = useState<boolean>(false);
  const [showQuiz, setShowQuiz] = useState<boolean>(false);
  const [currentQuiz, setCurrentQuiz] = useState<quiz>({
    questionId: 0,
    quizText: "",
    choices: [],
    imageId: "",
    teamId: 0,
    canAnswer: true,
    isTarget: false,
  });
  const [answerMap, setAnswerMap] = useState<Map<number, string[]> | undefined>(
    undefined,
  );
  const [results, setResults] = useState<boolean[]>([]);
  const { userStatus } = useContext(UserStatusContext);
  const nodeRef = useRef<HTMLDivElement | null>(null);
  const choiceRef = useRef<ChoiceRef | null>(null);
  const client = useRpcClient();
  const callAnswerApi = useCallback(async () => {
    if ("checkAnswers" in client) {
      // adminの場合
      const response = await client.checkAnswers({});
      const answers = response.answers.reduce(
        (map, answer) =>
          answer.answer !== undefined
            ? (map.has(answer.answer.choiceId)
                ? map
                    .get(answer.answer.choiceId)
                    ?.push(answer.teamColor)
                : map.set(answer.answer.choiceId, [answer.teamColor]),
              map)
            : map,
        new Map<number, string[]>(),
      );
      setAnswerMap(answers);
      setResults(response.answers.map((answer) => answer.isCorrect));
    } else if ("answer" in client) {
      // guestの場合
      let selectedChoice = choiceRef.current?.getSelected();
      if (!selectedChoice) {
        selectedChoice =
          currentQuiz.choices[
            Math.floor(Math.random() * currentQuiz.choices.length)
          ];
        choiceRef.current?.setSelected(selectedChoice);
      }
      const response = await client.answer(
        currentQuiz.questionId,
        selectedChoice.id,
        selectedChoice.text,
      );

      const answers = new Map<number, string[]>();
      // 回答可能な場合のみ結果を見る
      if (currentQuiz.canAnswer) {
        response.answerCount.forEach((cnt, idx) =>
          answers.set(idx + 1, Array(cnt).fill(userStatus.color)),
        );
        setResults([response.isCorrect]);
      }
      setAnswerMap(answers);
    } else {
      setAnswerMap(new Map<number, string[]>());
    }
  }, [client, currentQuiz, userStatus]);
  const { call: answer, isCalling, error } = useApiCaller(callAnswerApi);
  const streamFunc = useCallback(
    (signal: AbortSignal) => client.startQuest({}, { signal: signal }),
    [client],
  );
  const callBack = useCallback(
    (response: adminResponse | guestResponse) => {
      setCurrentQuiz((prev) => {
        if (prev.imageId === response.targetUserImageId) {
          return prev;
        } else {
          setAnswerMap(undefined);
          if (prev.teamId !== response.targetTeamId) {
            setTimeout(() => setShowQuiz(true), 1000);
          } else {
            setShowQuiz(true);
          }
          setIsEnableAnswer(true);
          return {
            questionId: response.questionId,
            quizText: response.question,
            choices: response.choices.map((value) => {
              return { id: value.choiceId, text: value.choiceText };
            }),
            imageId: response.targetUserImageId,
            teamId: response.targetTeamId,
            canAnswer: "canAnswer" in response ? response.canAnswer : true,
            isTarget: "isTarget" in response ? response.isTarget : false,
          };
        }
      });
      setRemainTime(response.lastTime);
      if (response.lastTime <= 0) {
        setIsEnableAnswer((prev) => {
          if (prev) answer();
          return false;
        });
      }
    },
    [answer],
  );
  const isQuizEnd = useStreamObserver(streamFunc, callBack);

  useEffect(() => {
    if (answerMap !== undefined) {
      setTimeout(() => {
        setShowQuiz(false);
        if ("nextQuiz" in client) {
          setTimeout(() => client.nextQuiz({}), 3000);
        }
      }, 3000);
    } else if (!error) {
      console.log(error);
    }
  }, [answerMap, error, client]);

  useEffect(() => {
    if (showQuiz) {
      const id = setInterval(() => {
        setRemainTime((prev) => prev - 0.1);
      }, 100);
      return () => clearInterval(id);
    }
  }, [showQuiz]);

  return (
    <>
      <CSSTransition
        nodeRef={nodeRef}
        classNames="answers"
        timeout={uiAnimateDuration}
        in={showQuiz}
      >
        <div ref={nodeRef} css={animationStyle}>
          <div
            css={containerStyle}
            style={{
              fontSize: userStatus.type === "admin" ? "x-large" : "medium",
            }}
          >
            <TimePressureGage remainTime={remainTime} />
            <QuizTextArea text={currentQuiz.quizText} />
            <Choices
              ref={choiceRef}
              choices={currentQuiz.choices}
              answers={answerMap}
            />
          </div>
          {!(isEnableAnswer && currentQuiz.canAnswer) && (
            <div css={maskStyle}>
              {isCalling ? (
                <p style={{ color: "white", alignSelf: "center" }}>
                  Connecting...
                </p>
              ) : !currentQuiz.canAnswer ? (
                <p style={{ color: "white", alignSelf: "center" }}>
                  囚われてしまったので答えられない！
                </p>
              ) : (
                results.length == 1 && (
                  <h2
                    style={{
                      alignSelf: "flex-start",
                      color: results[0]
                        ? "var(--main-color-2-light)"
                        : "var(--main-color-1-dark)",
                    }}
                    css={css`
                      animation: ${animation} 3s linear;
                    `}
                  >
                    {results[0] ? "正解" : "不正解"}
                  </h2>
                )
              )}
            </div>
          )}
        </div>
      </CSSTransition>
      {isQuizEnd && <button onClick={toNext}>結果を見る</button>}
    </>
  );
};

const animationStyle = css`
  position: relative;
  height: calc(90% - 1em);

  &.answers-enter {
    transform: "translateY(100%)";
  }
  &.answers-enter-active {
    transform: "translateY(0)";
    transition-duration: "${uiAnimateDuration}ms";
    transition-property: "transform";
  }
  &.answers-enter-done {
    transform: "translateY(0)";
  }
  &.answers-exit {
    transform: "translateY(0)";
  }
  &.answers-exit-active {
    transform: "translateY(100%)";
    transition-duration: "${uiAnimateDuration}ms";
    transition-property: "transform";
  }
  &.answers-exit-done {
    transform: "translateY(100%)";
  }
`;

const containerStyle = css`
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: calc(1em / 4);
  height: 100%;
`;

const maskStyle = css`
  display: flex;
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border: none;
  border-radius: 6px;
  background-color: rgba(64, 64, 64, 0.8);
  align-content: center;
  justify-content: center;
  z-index: 100;
  pointer-events: all;
`;

const animation = keyframes`
  0%, 50%, 100% {
    transform: scale(1);
  }
  25%, 75% {
    transform: scale(1.5);
  }
`;

export default QuizPage;
