import { css } from "@emotion/react";

export interface PersonalStats {
  name: string;
  correctRate: number;
  order: number;
}
export interface TeamStats {
  teamColor: string;
  teamOrder: number;
  correctRate: number;
  memberStats: PersonalStats[];
}

const toPercentString = (rate: number) => (rate * 100).toFixed(1) + "%";

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
            <th colSpan={Math.floor(maxMemberNum / 2)}>チーム</th>
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
            .sort((a, b) => a.teamOrder - b.teamOrder)
            .map((ts) => {
              return (
                <>
                  <tr key={ts.teamColor}>
                    <th rowSpan={2}>{ts.teamOrder}</th>
                    <td colSpan={Math.floor(maxMemberNum / 2)}>
                      {ts.teamColor}
                    </td>
                    <td colSpan={Math.ceil(maxMemberNum / 2)}>
                      {toPercentString(ts.correctRate)}
                    </td>
                  </tr>
                  <tr>
                    {ts.memberStats
                      .sort((a, b) => a.order - b.order)
                      .map((ps) => {
                        return (
                          <td>
                            <h4>
                              {ps.name}
                              {ps.order <= 3 && (
                                <span
                                  style={{
                                    color: ["gold", "silver", "bronze"][
                                      ps.order - 1
                                    ],
                                  }}
                                >
                                  ★
                                </span>
                              )}
                            </h4>
                            <p>{toPercentString(ps.correctRate)}</p>
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
  border-color: var(--sub-color-2-1-light);
  border-style: solid;
  border-collapse: collapse;
  border-radius: 6px;
  width: max-content;
  height: max-content;
  margin: 0 auto;

  thead th {
    position: sticky;
    top: 0;
    z-index: 10;
  }
  thead tr:last-child th {
    top: 1.5em;
  }
  thead tr {
    background-color: var(--sub-color-2-1-dark);
  }
  thead tr:last-child {
    border-bottom-style: solid;
  }
  tbody tr:nth-child(even) {
    border-bottom-style: dashed;
  }
  th,
  td {
    text-align: center;
    padding: 0 auto;
  }
  colgroup:first-child {
    position: sticky;
    left: 0;
    border-right-style: solid;
    z-index: 5;
  }
`;

export default TeamStatsPanel;
