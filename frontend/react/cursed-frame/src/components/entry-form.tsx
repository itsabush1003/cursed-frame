import { useRef, useState } from "react";

import { css } from "@emotion/react";
import { useForm, type SubmitHandler } from "react-hook-form";

export interface UserInfo {
  name: string;
  image: FileList;
}

const EntryForm = ({
  entry,
  imageUpload,
}: {
  entry: (name: string) => Promise<void>;
  imageUpload: (imageSource: string) => Promise<void>;
}) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<UserInfo>();
  const onSubmit: SubmitHandler<UserInfo> = async (data) => {
    await entry(data.name);
    await imageUpload(imageSource);
  };
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const [imageSource, setImageSource] = useState<string>("");
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length <= 0) {
      setImageSource("");
      return;
    }
    const file = e.target.files[0];
    const fileReader = new FileReader();
    fileReader.onload = () => {
      setImageSource(fileReader.result as string);
    };
    fileReader.readAsDataURL(file);
  };
  const { ref, ...imageInputProps } = register("image", { required: true });

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
            <img
              src={imageSource}
              width={"100%"}
              height={"100%"}
              style={{ maxWidth: "100%", height: "auto", objectFit: "cover" }}
            />
          ) : (
            <>
              プレビュー
              <br />
              クリックで画像選択
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
`;

const noticeStyle = css`
  font-size: 0.4em;
`;

export default EntryForm;
