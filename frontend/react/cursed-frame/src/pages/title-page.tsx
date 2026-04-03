import { useCallback, useContext, useMemo } from "react";

import { css } from "@emotion/react";

import GameStartButton from "@/components/game-start-button";
import { UserStatusContext } from "@/context/user-status-context";
import LocalStorageRepository from "@/services/repository/localstorage-repository";
import getAdminClient from "@/services/rpc/admin-client";

const TitlePage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const onStartButton = useCallback(async () => {
    if (userStatus.type === "admin") {
      const client = getAdminClient(() => "");
      const response = await client.entry();
      setUserStatus({ token: response.accessToken });
      LocalStorageRepository.saveSecret(response.reconnectKey);
    }
    toNext();
  }, [toNext, userStatus.type, setUserStatus]);

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
