import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { EntryService } from "@/gen/entry/v1/entry_pb";
import retryInterceptor from "@/services/rpc/retry-interceptor";
import { pathReplaceRegex } from "@/utils/util";

const connectNoAuthTransport = createConnectTransport({
  baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
  interceptors: [retryInterceptor(3)],
});

const rawEntryClient = createClient(EntryService, connectNoAuthTransport);

export const entryClient = {
  entry: async (userName: string) => {
    const entryResponse = await rawEntryClient.entry({ userName: userName });
    return {
      accessToken: entryResponse.accessToken,
      reconnectKey: entryResponse.reconnectKey,
    };
  },
  reconnect: async (reconnectKey: string) => {
    const reconnectResponse = await rawEntryClient.reconnect({
      reconnectKey: reconnectKey,
    });
    return reconnectResponse.accessToken;
  },
};
