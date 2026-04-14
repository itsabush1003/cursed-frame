import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { AdminService } from "@/gen/admin/v1/admin_pb";
import { entryClient } from "@/services/rpc/entry-client";
import reauthInterceptor from "@/services/rpc/reauth-interceptor";
import retryInterceptor from "@/services/rpc/retry-interceptor";
import getSetAccessTokenInterceptor from "@/services/rpc/set-access-token-interceptor";
import { pathReplaceRegex } from "@/utils/util";

const getAdminClient = (
  getToken: () => string,
  resetToken: (
    refreshTokenFunc: (key: string) => Promise<string>,
  ) => Promise<void>,
) => {
  const transport = createConnectTransport({
    baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
    interceptors: [
      retryInterceptor(3),
      reauthInterceptor(() => resetToken(entryClient.reconnect)),
      getSetAccessTokenInterceptor(getToken),
    ],
  });

  const adminClient = createClient(AdminService, transport);
  const client = {
    entry: async () => {
      return await entryClient.entry(
        "admin_" + (Math.random() * 100000).toFixed(0),
      );
    },
    reconnect: entryClient.reconnect,
  };
  return Object.assign(client, adminClient);
};

export default getAdminClient;
