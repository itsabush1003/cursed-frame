import { createClient, type CallOptions } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { LobbyService } from "@/gen/lobby/v1/lobby_pb";
import { QuestService } from "@/gen/quest/v1/quest_pb";
import { entryClient } from "@/services/rpc/entry-client";
import reauthInterceptor from "@/services/rpc/reauth-interceptor";
import retryInterceptor from "@/services/rpc/retry-interceptor";
import getSetAccessTokenInterceptor from "@/services/rpc/set-access-token-interceptor";
import { pathReplaceRegex } from "@/utils/util";

const getGuestClient = (
  getToken: () => string,
  resetToken: (
    refreshTokenFunc: (key: string) => Promise<string>,
  ) => Promise<void>,
) => {
  const connectWithAuthTransport = createConnectTransport({
    baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
    interceptors: [
      retryInterceptor(3),
      reauthInterceptor(() => resetToken(entryClient.reconnect)),
      getSetAccessTokenInterceptor(getToken),
    ],
  });

  const lobbyClient = createClient(LobbyService, connectWithAuthTransport);
  const questClient = createClient(QuestService, connectWithAuthTransport);

  const guestClient = {
    entry: entryClient.entry,
    reconnect: entryClient.reconnect,
    joinLobby: (options?: CallOptions) => lobbyClient.joinLobby({}, options),
    registProfile: async (profileId: number, answer: string) => {
      const registProfileResponse = await lobbyClient.registProfile({
        questionId: profileId,
        answer: answer,
      });
      return {
        questionId: registProfileResponse.nextQuestionId,
        question: registProfileResponse.nextQuestionText,
        isComplete: registProfileResponse.noMoreAnswer,
      };
    },
    isReady: async () => {
      lobbyClient.isReady({});
    },
    getTeamInfo: async () => {
      const getTeamInfoResponse = await lobbyClient.getTeamInfo({});
      return {
        teamId: getTeamInfoResponse.teamId,
        color: getTeamInfoResponse.teamColor,
        members: getTeamInfoResponse.members,
      };
    },
    startQuest: questClient.startQuest,
    answer: async (
      questionId: number,
      choiceId: number,
      choiceText: string,
    ) => {
      return questClient.answer({
        questionId: questionId,
        answer: { choiceId: choiceId, choiceText: choiceText },
      });
    },
    getResult: async () => {
      return questClient.getResult({});
    },
  };

  return guestClient;
};

export default getGuestClient;
