import { useEffect } from "react";

import { css } from "@emotion/react";
import { useForm, type SubmitHandler } from "react-hook-form";

export interface ProfileQuestion {
  questionId: number;
  question: string;
}
interface profileAnswer {
  answer: string;
}

const RegistProfileForm = ({
  currentQuestion,
  registFunc,
}: {
  currentQuestion: ProfileQuestion;
  registFunc: (question: ProfileQuestion, answer: string) => Promise<void>;
}) => {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitSuccessful },
  } = useForm<profileAnswer>();
  const onSubmit: SubmitHandler<profileAnswer> = async (data) => {
    await registFunc(currentQuestion, data.answer);
  };

  useEffect(() => {
    reset();
  }, [reset, isSubmitSuccessful]);

  return (
    <div css={containerStyle}>
      <form onSubmit={handleSubmit(onSubmit)} style={{ pointerEvents: "auto" }}>
        <h3 style={{ textAlign: "center" }}>{currentQuestion.question}</h3>
        <br />
        <label htmlFor="answer" style={{ textAlign: "left" }}>
          答え
        </label>
        <br />
        <input
          id="answer"
          type="textarea"
          {...register("answer", { required: true })}
        />
        {errors.answer && (
          <>
            <br />
            <span css={noticeStyle}>This field is required</span>
          </>
        )}
        <br />
        <span css={noticeStyle}>回答は簡潔に、一単語ないしは一言で</span>
        <br />
        <input type="submit" value="次へ" />
      </form>
    </div>
  );
};

const containerStyle = css`
  display: flex;
  width: 100%;
  justify-content: center;
  align-items: center;
  background-color: var(--sub-color-1-1-light);
`;

const noticeStyle = css`
  font-size: 0.4em;
`;

export default RegistProfileForm;
