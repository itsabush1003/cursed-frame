import { useContext } from "react";

import { css } from "@emotion/react";

import { UserStatusContext } from "@/context/user-status-context";

const GameStartButton = ({
  onStartButton,
}: {
  onStartButton: () => Promise<void>;
}) => {
  const { userStatus } = useContext(UserStatusContext);

  return (
    <div css={containerStyle}>
      {userStatus.type === "admin" && (
        <button disabled css={buttonStyle}>
          設定
        </button>
      )}
      <button css={buttonStyle} onClick={onStartButton}>
        {userStatus.type === "admin"
          ? "参加登録を受け付ける"
          : "ゲームに参加する"}
      </button>
    </div>
  );
};

const containerStyle = css`
  display: flex;
  width: 80%;
  gap: 5%;
  justify-content: space-evenly;
  align-items: center;
  margin: 0 auto;
`;

const buttonStyle = css`
  flex: 1;
  pointer-events: auto;
`;

export default GameStartButton;
