import { useState } from "react";

type uploadFunc = (imageSource: Blob) => Promise<void>;
const useImageService = (uploadFunc: uploadFunc) => {
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const uploader = async (imageSource: Blob) => {
    setIsUploading(true);
    await uploadFunc(imageSource);
    setIsUploading(false);
  };
  return { uploader, isUploading };
};

export default useImageService;
