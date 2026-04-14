import {
  lazy,
  Suspense,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";

import { css } from "@emotion/react";

import LoadingMark from "@/components/loding-mark";
import { UserStatusContext } from "@/context/user-status-context";
import { Result } from "@/gen/common/v1/common_pb";
import useApiCaller from "@/hooks/use-api-caller.ts";
import useRpcClient from "@/hooks/use-rpc-client.ts";

import type { TeamStats } from "@/components/team-stats-panel";

const TeamStatsPage = lazy(() => import("@/components/team-stats-panel"));
const PersonalStatsPage = lazy(
  () => import("@/components/personal-stats-panel"),
);

const ResultPage = () => {
  const { userStatus } = useContext(UserStatusContext);
  const [result, setResult] = useState<Result>(Result.UNSPECIFIED);
  const [stats, setStats] = useState<TeamStats[]>([]);
  const client = useRpcClient(userStatus.type);
  const callResultApi = useCallback(async () => {
    if ("endQuest" in client) {
      // adminの場合
      const response = await client.endQuest({});
      setResult(response.result);
      setStats(
        response.stats
          .sort((a, b) => a.teamOrder - b.teamOrder)
          .map((ts) => {
            return {
              teamColor: ts.teamColor,
              teamOrder: ts.teamOrder,
              correctRate: ts.teamCorrectRate,
              memberStats: ts.membersStats.map((ps) => ({
                name: ps.userName,
                correctRate: ps.correctRate,
                order: ps.personalOrder,
              })),
            };
          }),
      );
    } else if ("getResult" in client) {
      // guestの場合
      const response = await client.getResult();
      setResult(response.result);
      setStats([
        {
          teamColor: userStatus.color,
          teamOrder: response.teamOrder,
          correctRate: 0,
          memberStats: [
            {
              name: "self",
              correctRate: response.personalRate,
              order: response.personalOrder,
            },
          ],
        },
      ]);
    }
  }, [userStatus.color, client]);
  const { call: getResult, isCalling, error } = useApiCaller(callResultApi);

  useEffect(() => {
    getResult();
  }, [getResult]);

  return (
    <div css={backgroundStyle}>
      <h2>RESULT: {Result[result]}</h2>
      <Suspense fallback={<LoadingMark />}>
        {!isCalling &&
          (userStatus.type === "admin" ? (
            <TeamStatsPage teamStats={stats} />
          ) : (
            <PersonalStatsPage
              teamOrder={stats[0]?.teamOrder || 0}
              personalOrder={stats[0]?.memberStats[0]?.order || 0}
              correctRate={stats[0]?.memberStats[0]?.correctRate || 0}
            />
          ))}
      </Suspense>
      {error !== null && <p style={{ color: "red" }}>{error}</p>}
    </div>
  );
};

const backgroundStyle = css`
  background-color: var(--sub-color-2-1);
`;

export default ResultPage;
