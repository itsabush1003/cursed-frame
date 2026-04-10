import { useCallback, useContext, useMemo } from "react";

import { css } from "@emotion/react";

import GameStartButton from "@/components/game-start-button";
import { UserStatusContext } from "@/context/user-status-context";
import LocalStorageRepository from "@/services/repository/localstorage-repository";
import { entryClient } from "@/services/rpc/entry-client";

const TitlePage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const onStartButton = useCallback(async () => {
    const secretKey = LocalStorageRepository.getSecret();
    if (secretKey != null) {
      try {
        const token = await entryClient.reconnect(secretKey);
        setUserStatus({ token: token });
      } catch (e) {
        if (
          e instanceof Error &&
          e.message.includes("invalid UUID") // もし別の時のsecretKeyが同じドメインに残っていた場合に返ってくるエラーメッセージ（の一部）
        ) {
          LocalStorageRepository.removeSecret();
          // returnが無いとtoNextが２度呼ばれてしまう
          return onStartButton();
        } else {
          console.error(e);
          throw e;
        }
      }
    } else if (userStatus.type === "admin") {
      // adminのnameの自動生成がadminClientにあるので、entryClientではなくadminClientを使う
      const { default: getAdminClient } =
        await import("@/services/rpc/admin-client");
      const client = getAdminClient(() => userStatus.token);
      const response = await client.entry();
      setUserStatus({ token: response.accessToken });
      LocalStorageRepository.saveSecret(response.reconnectKey);
    }
    toNext();
  }, [toNext, userStatus, setUserStatus]);

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
