import { Code, ConnectError, type Interceptor } from "@connectrpc/connect";

const reauthInterceptor: (resetToken: () => Promise<void>) => Interceptor = (
  resetToken,
) => {
  let resetPromise: Promise<void> | null = null;
  return (next) => async (req) => {
    try {
      return await next(req);
    } catch (e) {
      if (e instanceof ConnectError && e.code === Code.Unauthenticated) {
        try {
          if (resetPromise === null) resetPromise = resetToken();
          await resetPromise;
        } finally {
          resetPromise = null;
        }
        return await next(req);
      } else throw e;
    }
  };
};

export default reauthInterceptor;
