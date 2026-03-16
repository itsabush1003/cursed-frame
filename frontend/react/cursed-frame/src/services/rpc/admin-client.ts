import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { AdminService } from "@/gen/admin/v1/admin_pb";
import { pathReplaceRegex } from "@/utils/util";

const transport = createConnectTransport({
  baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
});
const adminClient = createClient(AdminService, transport);

export default adminClient;
