import { Activity, useContext } from "react";

import EntryForm from "@/components/entry-form";
import { UserStatusContext } from "@/context/user-status-context";
import useEntry from "@/hooks/use-entry";
import useImageService from "@/hooks/use-image-service";
import { getImageService } from "@/services/api/image-service";
import LocalStorageRepository from "@/services/repository/localstorage-repository";
import guestClient from "@/services/rpc/guest-client";

const EntryPage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus } = useContext(UserStatusContext);
  const { entry, isLoading, error } = useEntry(
    guestClient(() => userStatus.token).entry,
    LocalStorageRepository.saveSecret,
  );
  const { uploader, isUploading } = useImageService(
    getImageService(() => {
      return userStatus.token;
    }).upload,
  );
  if (userStatus.token && !isUploading) {
    toNext();
    return <></>;
  }

  return (
    <>
      <Activity
        mode={
          (isLoading || isUploading) && error === null ? "hidden" : "visible"
        }
      >
        <EntryForm entry={entry} imageUpload={uploader} />
      </Activity>
    </>
  );
};

export default EntryPage;
