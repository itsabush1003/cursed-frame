import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

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
          const delay = Math.pow(2, i);
          await new Promise((resolve) => setTimeout(() => resolve, delay));
        } else throw e;
      }
    }
    throw lastErr;
  };

export default retryInterceptor;
