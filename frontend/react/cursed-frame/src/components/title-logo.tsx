import { css } from "@emotion/react";

const TitleLogo = () => {
  return (
    <>
      <h1 css={logoStyle}>
        Purify
        <br />
        the
        <br />
        Cursed Frame
      </h1>
    </>
  );
};

const logoStyle = css`
  position: absolute;
  top: 0px;
  align-self: center;
  z-index: 5;
  width: 100%;
  max-height: 50%;
  margin: 0;
  margin-top: 1rem;

  @media (max-width: 360px) {
    font-size: 2.5rem;
  }

  background: linear-gradient(var(--main-color-2), var(--main-color-1));
  background-clip: text;
  color: transparent;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  -webkit-text-stroke: 1pt gray;
  paint-order: stroke;
`;

export default TitleLogo;
