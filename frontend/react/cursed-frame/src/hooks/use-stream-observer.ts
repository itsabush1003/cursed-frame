import { useEffect, useEffectEvent, useState } from "react";

export type StreamFunc<T> = (signal: AbortSignal) => AsyncIterable<T>;
export type Callback<T> = (response: T) => void | boolean;
const useStreamObserver = <T>(
  streamFunc: StreamFunc<T>,
  callback: Callback<T>,
) => {
  const [isStreamEnd, setIsStreamEnd] = useState<boolean>(false);
  const callbackEvent = useEffectEvent(callback);

  useEffect(() => {
    const controller = new AbortController();
    const streamObserve = async () => {
      try {
        const stream = streamFunc(controller.signal);
        for await (const response of stream) {
          if (callbackEvent(response)) break;
        }
        setIsStreamEnd(true);
      } catch (e) {
        if (e instanceof Error && e.name === "AbortError") {
          setIsStreamEnd(true);
        } else {
          console.error(e);
          throw e;
        }
      }
    };
    streamObserve();
    return () => controller.abort();
  }, [streamFunc]);

  return isStreamEnd;
};

export default useStreamObserver;
