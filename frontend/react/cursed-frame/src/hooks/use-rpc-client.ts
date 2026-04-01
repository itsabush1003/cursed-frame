import { use, useContext, useMemo } from "react";

import {
  UserStatusContext,
  type UserType,
} from "@/context/user-status-context";

const cache = new Map<
  UserType,
  Promise<
    {
      admin: typeof import("@/services/rpc/admin-client.ts");
      guest: typeof import("@/services/rpc/guest-client.ts");
    }[UserType]
  >
>();

const useRpcClient = () => {
  const { userStatus } = useContext(UserStatusContext);
  const clientPromise = useMemo(() => {
    const cachedPromise = cache.get(userStatus.type);
    if (cachedPromise) return { type: userStatus.type, promise: cachedPromise };
    const promise =
      userStatus.type === "admin"
        ? import("@/services/rpc/admin-client.ts")
        : import("@/services/rpc/guest-client.ts");
    cache.set(userStatus.type, promise);
    return { type: userStatus.type, promise: promise };
  }, [userStatus.type]);
  const clientModule = use<
    | typeof import("@/services/rpc/admin-client.ts")
    | typeof import("@/services/rpc/guest-client.ts")
  >(clientPromise.promise);
  const client = useMemo(() => {
    return clientModule.default(() => userStatus.token);
  }, [clientModule, userStatus]);
  return client;
};

export default useRpcClient;
