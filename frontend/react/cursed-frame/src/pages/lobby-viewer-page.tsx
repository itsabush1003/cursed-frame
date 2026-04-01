import { useCallback, useContext, useMemo, useState } from "react";

import { css } from "@emotion/react";

import UsersTable, { type rowData } from "@/components/users-table";
import { UserStatusContext } from "@/context/user-status-context";
import useStreamObserver from "@/hooks/use-stream-observer";
import getAdminClient from "@/services/rpc/admin-client";

import type { OpenEntryResponse } from "@/gen/admin/v1/admin_pb";

const LobbyViewerPage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus } = useContext(UserStatusContext);
  const [entryUsersData, setEntryUsersData] = useState<rowData[]>([]);
  const [isExpectedNumEntered, setIsExpectedNumEntered] =
    useState<boolean>(false);
  const adminClient = useMemo(
    () => getAdminClient(() => userStatus.token),
    [userStatus.token],
  );
  const lobbyStatusStream = useCallback(
    (signal: AbortSignal) => adminClient.openEntry({}, { signal: signal }),
    [adminClient],
  );
  const reflectLobbyStatus = useCallback(
    (res: OpenEntryResponse) => {
    const rowDataList: rowData[] = [];
    for (const userData of res.enteredUsers) {
      const rowData = {
        userName: userData.userName,
        teamId: userData.teamId,
        isReady: userData.isReady,
        reject: () => {
          adminClient.rejectUser({ userId: userData.userId });
        },
      };
      rowDataList.push(rowData);
    }
    setEntryUsersData((prev) => {
      if (prev.length !== rowDataList.length) return rowDataList;
      const checkProps: (keyof rowData)[] = ["userName", "teamId", "isReady"];
      if (
        prev.every((data, i) =>
          checkProps.every((prop) => data[prop] === rowDataList[i][prop]),
        )
      )
        return prev;
      else return rowDataList;
    });
    if (rowDataList.length >= res.expectedUserNum)
      setIsExpectedNumEntered(true);
    },
    [adminClient],
  );
  const isLobbyReady = useStreamObserver(lobbyStatusStream, reflectLobbyStatus);
  const onCloseButton = async () => {
    if (!isExpectedNumEntered) {
      const result = window.confirm(
        "参加者数が予定参加者数に達していません。一度締め切ると参加者を増やすことができませんが、参加を締め切って良いですか？",
      );
      if (!result) return;
    }
    await adminClient.closeEntry({});
  };

  return (
    <div css={containerStyle}>
      <UsersTable data={entryUsersData} />
      {!isLobbyReady ? (
        <button
          onClick={onCloseButton}
          style={{ pointerEvents: "auto", width: "fit-content" }}
        >
          参加を締め切る
        </button>
      ) : (
        <div>
          <button onClick={toNext}>ゲームを始める</button>
        </div>
      )}
    </div>
  );
};

const containerStyle = css`
  position: relative;
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
  overflow-y: scroll;
  align-items: center;
  align-self: self-end;
  max-height: 100%;
  top: calc((100% - auto) / 2);
  background-color: var(--sub-color-1-1-light);
  pointer-events: auto;
`;

export default LobbyViewerPage;
