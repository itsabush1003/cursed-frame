import { useCallback, useState } from "react";

export type ApiCaller<T, S> = (args?: S) => Promise<T>;
const useApiCaller = <T, S>(caller: ApiCaller<T, S>) => {
  const [isCalling, setIsCalling] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const call = useCallback(
    async (args?: S) => {
      setIsCalling(true);
      try {
        return await caller(args);
      } catch (e) {
        if (e instanceof Error) setError(e.message);
        else setError("unknown error: " + String(e));
      } finally {
        setIsCalling(false);
      }
    },
    [caller],
  );
  return { call, isCalling, error };
};

export default useApiCaller;
