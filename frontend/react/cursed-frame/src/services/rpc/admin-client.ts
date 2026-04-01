import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { AdminService } from "@/gen/admin/v1/admin_pb";
import { EntryService } from "@/gen/entry/v1/entry_pb";
import { pathReplaceRegex } from "@/utils/util";

import getSetAccessTokenInterceptor from "./set-access-token-interceptor";

const getAdminClient = (getToken: () => string) => {
  const transport = createConnectTransport({
    baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
    interceptors: [getSetAccessTokenInterceptor(getToken)],
  });
  const adminClient = createClient(AdminService, transport);
  const entryClient = createClient(EntryService, transport);
  const client = {
    entry: async () => {
      const entryResponse = await entryClient.entry({
        userName: "admin_" + (Math.random() * 100000).toFixed(0),
      });
      return {
        accessToken: entryResponse.accessToken,
        reconnectKey: entryResponse.reconnectKey,
      };
    },
    reconnect: async (reconnectKey: string) => {
      const reconnectResponse = await entryClient.reconnect({
        reconnectKey: reconnectKey,
      });
      return reconnectResponse.accessToken;
    },
  };
  return Object.assign(client, adminClient);
};

export default getAdminClient;
