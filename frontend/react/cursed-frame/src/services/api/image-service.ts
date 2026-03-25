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
    upload: async (imageBlob: Blob) => {
      const formData = new FormData();
      formData.append("image", imageBlob, "profile-image.jpg");

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
