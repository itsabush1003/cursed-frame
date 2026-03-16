import { pathReplaceRegex } from "@/utils/util";

const baseUrl = window.location.pathname.replace(
  pathReplaceRegex,
  "/rest/images",
);

export const getImageService = (getToken: () => string) => {
  const imageService = {
    download: async (imageId: string) => {
      const token = getToken();
      const response = await fetch(`${baseUrl}/${imageId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) throw new Error("Download Failed");
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    },
    upload: async (imageSource: string) => {
      const formData = new FormData();
      const imageData = await fetch(imageSource);
      const imageBlob = await imageData.blob();
      formData.append("image", imageBlob, "picture.jpg");

      const token = getToken();
      const response = await fetch(baseUrl, {
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
