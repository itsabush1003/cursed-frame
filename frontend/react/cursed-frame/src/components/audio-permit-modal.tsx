import { useRef } from "react";

import { css } from "@emotion/react";
import { CSSTransition } from "react-transition-group";

type boolStateSetter = (bool: boolean) => void;

const AudioPermitModal = ({
  isVisible,
  setIsVisible,
  setAudioEnable,
}: {
  isVisible: boolean;
  setIsVisible: boolStateSetter;
  setAudioEnable: boolStateSetter;
}) => {
  const handleEnableAudio = () => {
    setAudioEnable(true);
    setIsVisible(false);
  };

  const handleDisableAudio = () => {
    setAudioEnable(false);
    setIsVisible(false);
  };

  const nodeRef = useRef<HTMLDivElement | null>(null);

  return (
    <CSSTransition
      nodeRef={nodeRef}
      classNames="modal"
      in={isVisible}
      timeout={1000}
      unmountOnExit
    >
      <div ref={nodeRef} css={overlayStyle}>
        <div css={modalStyle}>
          <h2>SOUND NOTICE</h2>
          <p css={css({ "@media(max-width: 340px)": { fontSize: "0.8em" } })}>
            このゲームではBGM・音声が流れます。
            <br />
            音声を有効にしますか？
          </p>

          <div css={buttonContainer}>
            <button onClick={handleEnableAudio} css={buttonStyle}>
              <img src="sound_on.svg"></img>はい
              <br />
              （音声を有効にする）
            </button>
            <button onClick={handleDisableAudio} css={buttonStyle}>
              <img src="sound_off.svg"></img>いいえ
              <br />
              （音声を無効にする）
            </button>
          </div>
        </div>
      </div>
    </CSSTransition>
  );
};

const overlayStyle = css({
  position: "fixed",
  top: 0,
  left: 0,
  width: "100vw",
  height: "100vh",
  backgroundColor: "rgba(0,0,0,0.8)",
  display: "flex",
  justifyContent: "center",
  alignItems: "center",
  zIndex: 1000,
  "&.modal-enter > div": {
    opacity: 0,
    transform: "translateY(-50%)",
  },
  "&.modal-enter-active > div": {
    opacity: 1,
    transform: "translateY(0)",
    transitionDuration: "200ms",
    transitionProperty: "opacity, transform",
  },
  "&.modal-exit > div": {
    opacity: 1,
    transform: "translateY(0)",
  },
  "&.modal-exit-active > div": {
    opacity: 0,
    transform: "translateY(-50%)",
    transitionDuration: "200ms",
    transitionProperty: "opacity, transform",
  },
});

const modalStyle = css({
  padding: "20px",
  textAlign: "center",
  borderRadius: "8px",
  minWidth: "75%",
  maxHeight: "80%",
  color: "white",
});

const buttonContainer = css({
  marginTop: "20px",
  display: "flex",
  gap: "10px",
  alignItems: "center",
  justifyContent: "space-evenly",
});

const buttonStyle = css({
  flex: "1",
  padding: "10px 20px",
  cursor: "pointer",
  backgroundColor: "#fff",
  color: "var(--main-color-2)",
  maxInlineSize: "max-content",
  "@media(max-width: 480px)": { fontSize: "9px" },
});

export default AudioPermitModal;
