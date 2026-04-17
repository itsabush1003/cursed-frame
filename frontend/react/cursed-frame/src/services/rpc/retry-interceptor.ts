import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

import { waitExponentialBackoff } from "@/utils/util";

const retryInterceptor: (maxRetry: number) => Interceptor =
  (maxRetry) => (next) => async (req) => {
    let lastErr: unknown;
    for (let i = 0; i < maxRetry; i++) {
      try {
        return await next(req);
      } catch (e) {
        lastErr = e;
        if (
          e instanceof ConnectError &&
          e.code in [Code.Internal, Code.Unavailable, Code.Unknown]
        ) {
          await waitExponentialBackoff(i);
        } else throw e;
      }
    }
    throw lastErr;
  };

export default retryInterceptor;
