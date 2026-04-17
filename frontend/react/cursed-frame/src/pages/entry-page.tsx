import { Activity, useContext } from "react";

import EntryForm from "@/components/entry-form";
import { UserStatusContext } from "@/context/user-status-context";
import useEntry from "@/hooks/use-entry";
import useImageService from "@/hooks/use-image-service";
import { resetToken } from "@/hooks/use-rpc-client";
import { getImageService } from "@/services/api/image-service";
import LocalStorageRepository from "@/services/repository/localstorage-repository";
import { entryClient } from "@/services/rpc/entry-client";

const EntryPage = ({ toNext }: { toNext: () => void }) => {
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const { entry, isLoading, error } = useEntry(
    entryClient.entry,
    LocalStorageRepository.saveSecret,
  );
  const { uploader, isUploading } = useImageService(
    getImageService(
      () => userStatus.token,
      () =>
        resetToken(
          LocalStorageRepository.getSecret,
          entryClient.reconnect,
          (token: string) => setUserStatus({ token: token }),
        ),
    ).upload,
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
