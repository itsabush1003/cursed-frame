import {
  useCallback,
  useImperativeHandle,
  useRef,
  useState,
  type RefObject,
} from "react";

import { css } from "@emotion/react";
import Cropped, {
  type Area,
  type MediaSize,
  type Point,
} from "react-easy-crop";

export interface CropRef {
  getCroppedImage: (blobType: string) => Promise<Blob | null>;
  clearCroppedArea: () => void;
}

const cropHeight = 512;
const cropWidth = 384; // 512 * 3 / 4

const CropCanvas = ({
  ref,
  uploadedImage,
  cropEventTrigger,
}: {
  ref: RefObject<CropRef | null>;
  uploadedImage: string;
  cropEventTrigger: () => void;
}) => {
  const [crop, setCrop] = useState<Point>({ x: 0, y: 0 });
  const [zoom, setZoom] = useState<number>(1);
  const [minZoom, setMinZoom] = useState<number>(1);
  const [croppedAreaPixels, setCroppedAreaPixels] = useState<Area>({
    height: cropHeight,
    width: cropWidth,
    x: 0,
    y: 0,
  });
  const imageRef = useRef<HTMLImageElement>(null);

  useImperativeHandle(ref, () => {
    return {
      getCroppedImage: async (blobType: string) => {
        if (imageRef.current === null) return null;
        const canvas = document.createElement("canvas");
        const ctx = canvas.getContext("2d");
        // 最終的な出力を2のべき乗の正方形にする
        canvas.width = cropHeight;
        canvas.height = cropHeight;
        ctx?.drawImage(
          imageRef.current,
          croppedAreaPixels.x,
          croppedAreaPixels.y,
          croppedAreaPixels.width,
          croppedAreaPixels.height,
          (cropHeight - cropWidth) / 2, // 左右に余白を作る
          0,
          cropWidth,
          cropHeight,
        );
        return new Promise((resolve, reject) =>
          canvas.toBlob(
            (blob) => (blob ? resolve(blob) : reject(null)),
            blobType,
          ),
        );
      },
      clearCroppedArea: () => {
        setCrop({ x: 0, y: 0 });
        setZoom(1);
        setMinZoom(1);
      },
    };
  }, [croppedAreaPixels]);

  const onCropComplete = useCallback(
    (_croppedArea: Area, croppedAreaPixels: Area) => {
      setCroppedAreaPixels(croppedAreaPixels);
      cropEventTrigger();
    },
    [cropEventTrigger],
  );

  const onMediaLoaded = useCallback((mediaSize: MediaSize) => {
    if (
      mediaSize.naturalWidth < cropWidth ||
      mediaSize.naturalHeight < cropHeight
    ) {
      const needyZoom = Math.max(
        cropWidth / mediaSize.naturalWidth,
        cropHeight / mediaSize.naturalHeight,
      );
      setMinZoom(needyZoom);
      setZoom(needyZoom);
    } else {
      setMinZoom(
        Math.min(
          Math.max(
            cropWidth / mediaSize.naturalWidth,
            cropHeight / mediaSize.naturalHeight,
          ),
          1,
        ),
      );
    }
  }, []);

  return (
    <>
      <div css={cropContainerStyle}>
        <Cropped
          image={uploadedImage}
          crop={crop}
          zoom={zoom}
          aspect={3 / 4}
          minZoom={minZoom}
          cropShape="rect"
          showGrid={false}
          onCropChange={setCrop}
          onCropComplete={onCropComplete}
          onZoomChange={setZoom}
          onMediaLoaded={onMediaLoaded}
          setImageRef={(ref: RefObject<HTMLImageElement>) => {
            imageRef.current = ref.current;
          }}
        />
      </div>
    </>
  );
};

const cropContainerStyle = css`
  position: relative;
  width: 100%;
  height: 100%;
`;

export default CropCanvas;
