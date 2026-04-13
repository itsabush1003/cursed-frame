import { pathReplaceRegex, waitExponentialBackoff } from "@/utils/util";

const baseUrl = window.location.pathname.replace(
  pathReplaceRegex,
  "/rest/images",
);
const fetchWithRetry = async (
  url: URL | string,
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
  maxRetry: number = 3,
) => {
  const imageService = {
    download: async (imageId: string) => {
      const token = getToken();
      const response = await fetchWithRetry(`${baseUrl}/${imageId}`, maxRetry, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) throw new Error("Download Failed");
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    },
    upload: async (imageBlob: Blob) => {
      const formData = new FormData();
      formData.append("image", imageBlob, "profile-image.jpg");

      const token = getToken();
      const response = await fetchWithRetry(baseUrl, maxRetry, {
        method: "POST",
        body: formData,
        headers: { Authorization: `Bearer ${token}` },
      });

      if (!response.ok) throw new Error("Upload Failed");
      return;
    },
  };
  return imageService;
};
