import { useCallback, useContext, useMemo } from "react";

import { css } from "@emotion/react";

import GameStartButton from "@/components/game-start-button";
import { UserStatusContext } from "@/context/user-status-context";

const TitlePage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus } = useContext(UserStatusContext);
  const onStartButton = useCallback(() => {
    toNext();
  }, [toNext]);

  const layoutStyle = useMemo(
    () => css`
      align-self: self-end;
      margin-top: ${userStatus.type === "admin" ? "10%" : "25%"};
      margin-bottom: 1em;
    `,
    [userStatus.type],
  );

  return (
    <div css={layoutStyle}>
      <GameStartButton onStartButton={onStartButton} />
    </div>
  );
};

export default TitlePage;
