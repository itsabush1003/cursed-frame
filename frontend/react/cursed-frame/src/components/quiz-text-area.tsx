import { css } from "@emotion/react";

export const QuizTextArea = ({ text }: { text: string }) => {
  return <div css={TextAreaStyle}>{text}</div>;
};

const TextAreaStyle = css`
  width: 100%;
  height: 3.5em;
  text-align: center;
  align-content: center;
  border: 1px solid;
  border-radius: 4px;
  border-color: var(--sub-color-2-1);
  background-color: lightgray;
  color: black;
`;
