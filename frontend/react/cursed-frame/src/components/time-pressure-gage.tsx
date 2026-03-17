import { css } from "@emotion/react";

export const MaxRemainTime = 15;

export const TimePressureGage = ({ remainTime }: { remainTime: number }) => {
  const proportion = Math.max(Math.min(remainTime / MaxRemainTime, 1), 0);

  return (
    <div css={baseStyle}>
      <div css={gageStyle} style={{ width: `${proportion * 100}%` }} />
      <div css={numeralStyle}>
        {remainTime > 0 ? remainTime.toFixed(1) : 0.0}
      </div>
    </div>
  );
};

const baseStyle = css`
  width: 100%;
  height: 1.5em;
  position: relative;
  font-size: max(medium, smaller);
  flex-shrink: 0;
  flex-grow: 0;
`;

const gageStyle = css`
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background-color: var(--main-color-2);
  z-index: 2;
`;

const numeralStyle = css`
  position: absolute;
  top: 0%;
  left: calc(50% - 1em);
  right: 50%;
  height: 100%;
  text-align: center;
  color: var(--comp-color-2);
  z-index: 3;
`;
