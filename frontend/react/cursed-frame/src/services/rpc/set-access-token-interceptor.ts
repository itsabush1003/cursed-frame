import type { Interceptor } from "@connectrpc/connect";

const getSetAccessTokenInterceptor: (getToken: () => string) => Interceptor =
  (getToken) => (next) => async (req) => {
    if (!req.url.includes("/entry")) {
      const token = getToken();
      req.header.set("Authorization", `Bearer ${token}`);
    }
    return await next(req);
  };

export default getSetAccessTokenInterceptor;
