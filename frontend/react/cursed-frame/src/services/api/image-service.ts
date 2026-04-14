import { pathReplaceRegex, waitExponentialBackoff } from "@/utils/util";

const baseUrl = window.location.pathname.replace(
  pathReplaceRegex,
  "/rest/images",
);
const fetchWithRetry = async (
  url: RequestInfo | URL,
  maxRetry: number,
  init?: RequestInit,
) => {
  for (let i = 0; i < maxRetry; i++) {
    try {
      const response = await fetch(url, init);
      if (!response.ok && Math.trunc(response.status / 100) === 5)
        throw new Error("Retry because Server Error");
      return response;
    } catch {
      await waitExponentialBackoff(i);
    }
  }
  return await fetch(url, init);
};

export const getImageService = (
  getToken: () => string,
  resetToken: () => Promise<void>,
  maxRetry: number = 3,
) => {
  let resetFlg = false; // resetToken実行後に再帰で無限ループに入らせない為のフラグ
  const imageService = {
    download: async (imageId: string): Promise<string> => {
      const token = getToken();
      const response = await fetchWithRetry(`${baseUrl}/${imageId}`, maxRetry, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) {
        if (response.status === 401 && !resetFlg) {
          resetFlg = true;
          try {
            await resetToken();
            return await imageService.download(imageId);
          } finally {
            resetFlg = false;
          }
        } else throw new Error("Download Failed");
      }
      resetFlg = false;
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    },
    upload: async (imageBlob: Blob): Promise<void> => {
      const formData = new FormData();
      formData.append("image", imageBlob, "profile-image.jpg");

      const token = getToken();
      const response = await fetchWithRetry(baseUrl, maxRetry, {
        method: "POST",
        body: formData,
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!response.ok) {
        if (response.status === 401 && !resetFlg) {
          resetFlg = true;
          try {
            await resetToken();
            return await imageService.upload(imageBlob);
          } finally {
            resetFlg = false;
          }
        } else throw new Error("Upload Failed");
      }
      resetFlg = false;
      return;
    },
  };
  return imageService;
};
