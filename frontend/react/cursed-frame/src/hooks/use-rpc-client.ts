import { use, useContext, useMemo } from "react";

import {
  UserStatusContext,
  type UserType,
} from "@/context/user-status-context";
import LocalStorageRepository from "@/services/repository/localstorage-repository";

type adminClientModule = typeof import("@/services/rpc/admin-client.ts");
type guestClientModule = typeof import("@/services/rpc/guest-client.ts");
type clientModule = adminClientModule | guestClientModule;
type clientPromise =
  | {
      type: "admin";
      promise: Promise<adminClientModule>;
    }
  | {
      type: "guest";
      promise: Promise<guestClientModule>;
    };

const cache = new Map<UserType, clientPromise>();

const useClientPromise = <T extends UserType>(userType: T) => {
  const clientPromise = useMemo(() => {
    const cachedPromise = cache.get(userType);
    if (cachedPromise && cachedPromise.type === userType) return cachedPromise;
    const promise: clientPromise =
      userType === "admin"
        ? { type: userType, promise: import("@/services/rpc/admin-client.ts") }
        : { type: userType, promise: import("@/services/rpc/guest-client.ts") };
    cache.set(userType, promise);
    return promise;
  }, [userType]);
  return clientPromise as Extract<clientPromise, { type: T }>;
};

export const resetToken = async (
  getSecretKey: () => string | null,
  refreshToken: (key: string) => Promise<string>,
  setToken: (newToken: string) => void,
) => {
  const secretKey = getSecretKey();
  if (secretKey === null) throw new Error("Secret key is not found.");
  const newToken = await refreshToken(secretKey);
  setToken(newToken);
};

const useClientConstruct = <T extends clientModule>(clientModule: T) => {
  const { userStatus, setUserStatus } = useContext(UserStatusContext);
  const client = useMemo(() => {
    return clientModule.default(
      () => userStatus.token,
      (refreshFunc: (key: string) => Promise<string>) =>
        resetToken(
          LocalStorageRepository.getSecret,
          refreshFunc,
          (token: string) => setUserStatus({ token: token }),
        ),
    );
  }, [clientModule, userStatus.token, setUserStatus]);
  return client as ReturnType<T["default"]>;
};

const useRpcClient = <T extends UserType>(userType: T) => {
  const clientPromise = useClientPromise(userType);
  const clientModule = use<clientModule>(clientPromise.promise) as Awaited<
    Extract<clientPromise, { type: T }>["promise"]
  >;
  return useClientConstruct(clientModule);
};

export default useRpcClient;
