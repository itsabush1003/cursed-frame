import { css } from "@emotion/react";

export interface PersonalStats {
  name: string;
  correctRate: number;
  order: number;
}
export interface TeamStats {
  teamName: string;
  correctRate: number;
  memberStats: PersonalStats[];
}

const TeamStatsPanel = ({ teamStats }: { teamStats: TeamStats[] }) => {
  const maxMemberNum = teamStats.reduce((prevMax, ts) => {
    if (ts.memberStats.length > prevMax) return ts.memberStats.length;
    else return prevMax;
  }, 0);
  return (
    <div style={{ overflow: "scroll" }}>
      <table css={tableStyle}>
        <colgroup span={1}></colgroup>
        <colgroup span={maxMemberNum}></colgroup>
        <thead>
          <tr key="first">
            <th rowSpan={2}>チーム順位</th>
            <th colSpan={Math.floor(maxMemberNum / 2)}>チーム名</th>
            <th colSpan={Math.ceil(maxMemberNum / 2)}>チーム正解率</th>
          </tr>
          <tr key="last">
            <th colSpan={maxMemberNum}>
              各チームメンバーの結果（名前、個人正解率）
            </th>
          </tr>
        </thead>
        <tbody>
          {teamStats
            .sort((a, b) => b.correctRate - a.correctRate)
            .map((ts, idx) => {
              return (
                <>
                  <tr key={ts.teamName}>
                    <th rowSpan={2}>{idx}</th>
                    <td colSpan={Math.floor(maxMemberNum / 2)}>
                      {ts.teamName}
                    </td>
                    <td colSpan={Math.ceil(maxMemberNum / 2)}>
                      {Math.round(ts.correctRate * 1000) / 10}%
                    </td>
                  </tr>
                  <tr>
                    {ts.memberStats
                      .sort((a, b) => a.order - b.order)
                      .map((ps) => {
                        return (
                          <td>
                            <h6>
                              {ps.name}
                              {ps.order <= 3 && (
                                <span
                                  style={{
                                    color: ["gold", "silver", "bronze"][
                                      ps.order
                                    ],
                                  }}
                                >
                                  ★
                                </span>
                              )}
                            </h6>
                            <p>{Math.round(ps.correctRate * 1000) / 10}%</p>
                          </td>
                        );
                      })}
                  </tr>
                </>
              );
            })}
        </tbody>
      </table>
    </div>
  );
};

const tableStyle = css`
  border-color: var(--main-color-2);
  border-collapse: collapse;
  width: max-content;
  height: max-content;

  thead th {
    position: sticky;
    top: 0;
    z-index: 10;
  }
  thead tr:last-child th {
    top: 1em;
  }
  th,
  td {
    text-align: center;
    padding: 0 auto;
  }
  colgroup:first-child {
    position: sticky;
    left: 0;
    z-index: 5;
  }
`;

export default TeamStatsPanel;
