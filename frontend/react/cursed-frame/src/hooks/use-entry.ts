import { useContext, useState } from "react";

import { UserStatusContext } from "@/context/user-status-context";

export interface entryInfo {
  accessToken: string;
  reconnectKey: string;
}

type entryFunc = (name: string) => Promise<entryInfo>;
type saveSecretFunc = (key: string) => void;

const useEntry = (entryFunc: entryFunc, saveSecretFunc: saveSecretFunc) => {
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const entry = async (userName: string) => {
    if (userStatus.token) {
      setError(null);
      setIsLoading(false);
      return;
    }
    setIsLoading(true);
    try {
      const { accessToken, reconnectKey } = await entryFunc(userName);
      setUserStatus({ token: accessToken });
      saveSecretFunc(reconnectKey);
      setError(null);
    } catch (e) {
      if (e instanceof Error) setError(e.message);
      else setError("unknown error: " + String(e));
    } finally {
      setIsLoading(false);
    }
  };
  return { entry, isLoading, error };
};

export default useEntry;
