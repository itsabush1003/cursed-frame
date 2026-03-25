import { useCallback, useEffect, useRef, useState } from "react";

import { css } from "@emotion/react";
import { useForm, type SubmitHandler } from "react-hook-form";

import CropCanvas, { type CropRef } from "./crop-canvas";

type emptyObject = { [key: string]: never };
type getBlobFn = { func: ((blobType: string) => Promise<Blob | null>) | null };

export interface UserInfo {
  name: string;
  image: FileList;
}

const EntryForm = ({
  entry,
  imageUpload,
}: {
  entry: (name: string) => Promise<void>;
  imageUpload: (imageSource: Blob) => Promise<void>;
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<UserInfo>();
  const cropRef = useRef<CropRef>(null);
  const imageInputRef = useRef<HTMLInputElement>(null);
  const [imageSource, setImageSource] = useState<string>("");
  // react compilerのバグで、handleSubmitに渡す関数の中でrefの参照ができないので、workaroundとして関数をstate化
  const [getCroppedImage, setGetCroppedImage] = useState<getBlobFn>({
    func: null,
  });
  // 上記関数の都度更新のためのトリガー
  const [emptyTrigger, setEmptyTrigger] = useState<emptyObject>({});
  const trigger = useCallback(() => setEmptyTrigger({}), []);

  const onSubmit: SubmitHandler<UserInfo> = async (data) => {
    if (getCroppedImage.func !== null) {
      const blob = await getCroppedImage.func("image/jpeg");
      if (blob === null) return;
      await entry(data.name);
      await imageUpload(blob);
    }
  };
  const handleFileChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (!e.target.files || e.target.files.length <= 0) {
        setImageSource("");
        return;
      }
      const file = e.target.files[0];
      setImageSource((prev) => {
        const newURL = URL.createObjectURL(file);
        if (prev === newURL) return prev;
        if (prev !== "") URL.revokeObjectURL(prev);
        cropRef.current?.clearCroppedArea();
        return newURL;
      });
    },
    [],
  );
  const { ref, ...imageInputProps } = register("image", { required: true });

  useEffect(() => {
    // 無駄な再レンダリングを避けるため、常にprevを返す
    setGetCroppedImage((prev) => {
      if (cropRef.current?.getCroppedImage === undefined) return prev;
      else {
        prev.func = cropRef.current.getCroppedImage;
        return prev;
      }
    });
  }, [emptyTrigger]);

  return (
    <div css={containerStyle}>
      <form onSubmit={handleSubmit(onSubmit)} style={{ pointerEvents: "auto" }}>
        <h3 style={{ textAlign: "center" }}>画像と名前を登録してください</h3>
        <label htmlFor="image">画像</label>
        <div
          style={{
            width: "75vw",
            height: "100vw",
            borderStyle: "dashed",
            textAlign: "center",
            margin: "0 auto",
          }}
          onClick={() => {
            imageInputRef.current?.click();
          }}
        >
          {imageSource ? (
            <CropCanvas
              ref={cropRef}
              uploadedImage={imageSource}
              cropEventTrigger={trigger}
            />
          ) : (
            <>
              プレビュー
              <br />
              クリック／タップで画像選択
            </>
          )}
        </div>
        <br />
        <span css={noticeStyle}>
          ※他の参加者から識別できる画像にしてください
          <br />
          また、画像の著作権や写真の肖像権には注意し、
          <br />
          公序良俗に反しない画像を登録してください
        </span>
        <input
          id="image"
          ref={(e) => {
            ref(e);
            imageInputRef.current = e;
          }}
          {...imageInputProps}
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          style={{ display: "none" }}
        />
        {errors.image && (
          <>
            <br />
            <span css={noticeStyle}>This field is required</span>
          </>
        )}
        <br />
        <label htmlFor="name">名前</label>
        <br />
        <input id="name" {...register("name", { required: true })} />
        {errors.name && (
          <>
            <br />
            <span css={noticeStyle}>This field is required</span>
          </>
        )}
        <br />
        <span css={noticeStyle}>
          ※非本名可、
          <br />
          主催者や他の参加者から識別できる名前にしてください
        </span>
        <br />
        <input type="submit" value="登録する" />
      </form>
    </div>
  );
};

const containerStyle = css`
  display: flex;
  width: 100%;
  justify-content: center;
  align-items: center;
  background-color: var(--sub-color-1-1-light);
  overflow-y: scroll;
  scrollbar-gutter: stable;
  scrollbar-width: 0.5em;
  scrollbar-color: var(--sub-color-1-1-dark);
  ::-webkit-scrollbar {
    width: 0.5em;
    display: block;
  }
  ::-webkit-scrollbar-thumb {
    background-color: var(--sub-color-1-1-dark);
  }
`;

const noticeStyle = css`
  font-size: 0.4em;
`;

export default EntryForm;
