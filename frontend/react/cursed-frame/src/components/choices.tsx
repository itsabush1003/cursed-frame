import {
  useContext,
  useEffect,
  useImperativeHandle,
  useState,
  type RefObject,
} from "react";

import { css } from "@emotion/react";

import { UserStatusContext } from "@/context/user-status-context";

export interface Choice {
  id: number;
  text: string;
}
export interface ChoiceRef {
  getSelected: () => Choice | null;
  setSelected: (choice: Choice) => void;
}

export const Choices = ({
  ref,
  choices,
  answers,
}: {
  ref: RefObject<ChoiceRef | null>;
  choices: Choice[];
  answers?: Map<number, string[]>;
}) => {
  const { userStatus } = useContext(UserStatusContext);
  const [selected, setSelected] = useState<Choice | null>(null);

  useImperativeHandle(ref, () => ({
    getSelected: () => selected,
    setSelected: (choice: Choice) =>
      !selected ? setSelected(choice) : undefined,
  }));

  useEffect(() => {
    return () => setSelected(null);
  }, [choices]);

  const labelStyle = css`
    display: inline-block;
    box-sizing: border-box;
    border: 1px solid;
    border-radius: 4px;
    border-color: var(--comp-color-1);
    background-color: var(--comp-color-1-light);
    color: black;
    cursor: pointer;
    height: ${userStatus.type === "admin" ? "100%" : "20%"};
    width: 100%;
    align-content: center;
    pointer-events: auto;

    &:has(input[type="radio"]:checked) {
      border-width: 3px;
      border-color: ${userStatus.color.toLowerCase()};
    }

    ${userStatus.type === "admin" &&
    `&:nth-child(3):last-child { /* 選択肢が３つの時に配置を整える */
            grid-column: 1 / span 2; /* 1列目から2列分（全幅）を占有 */
            justify-self: center;
            width: 50%; /* 100%のままだと全幅まで伸びてしまうので */
        }`}
  `;

  return (
    <div
      css={
        userStatus.type === "admin" ? adminContainerStyle : guestContainerStyle
      }
    >
      {choices.map((choice) => (
        <>
          <label key={choice.id} css={labelStyle}>
            <input
              type="radio"
              value={choice.text}
              checked={selected?.id === choice.id}
              onChange={() => setSelected(choice)}
              style={{ display: "none" }}
            />
            {choice.text}
            {answers !== undefined &&
              answers
                .get(choice.id)
                ?.map((col) => <span style={{ color: col }}>・</span>)}
          </label>
        </>
      ))}
    </div>
  );
};

const guestContainerStyle = css`
  display: flex;
  flex-direction: column; /* 縦に並べる */
  justify-content: center;
  align-items: center;
  height: 90%;
  width: 100%;
  gap: 2px;
  flex-shrink: 0;
  flex-grow: 0;
`;

const adminContainerStyle = css`
  display: grid; /* 2×2に並べる */
  grid-template-rows: repeat(2, 50%);
  grid-template-columns: repeat(2, 50%);
  align-self: self-end;
  height: 4em;
  width: 100%;
  gap: 5px;
  flex-shrink: 0;
  flex-grow: 0;
`;
