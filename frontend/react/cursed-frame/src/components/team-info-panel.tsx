import { css } from "@emotion/react";

const TeamInfoPanel = ({
  teamId,
  teamColor,
  members,
}: {
  teamId: number;
  teamColor: string;
  members: string[];
}) => {
  return (
    <div style={{ alignContent: "center" }}>
      <p>あなたは</p>
      <h3 style={{ color: teamColor.toLowerCase() }}>
        #{teamId}: {teamColor}
      </h3>
      <p>の魔法使いと契約しました</p>
      <br />
      <p>他に</p>
      <div css={membersStyle}>
        {members.map((name) => (
          <p css={memberNameStyle}>{name}</p>
        ))}
      </div>
      <p>も契約しています</p>
    </div>
  );
};

const membersStyle = css`
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  width: 100%;
`;

const memberNameStyle = css`
  text-align: center;
  flex-grow: 1;
`;

export default TeamInfoPanel;
