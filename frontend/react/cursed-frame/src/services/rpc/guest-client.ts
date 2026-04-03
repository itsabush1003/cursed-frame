import { createClient, type CallOptions } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { EntryService } from "@/gen/entry/v1/entry_pb";
import { LobbyService } from "@/gen/lobby/v1/lobby_pb";
import { QuestService } from "@/gen/quest/v1/quest_pb";
import getSetAccessTokenInterceptor from "@/services/rpc/set-access-token-interceptor";
import { pathReplaceRegex } from "@/utils/util";

const getGuestClient = (getToken: () => string) => {
  const connectTransport = createConnectTransport({
    baseUrl: window.location.pathname.replace(pathReplaceRegex, "/rpc"),
    interceptors: [getSetAccessTokenInterceptor(getToken)],
  });

  const entryClient = createClient(EntryService, connectTransport);
  const lobbyClient = createClient(LobbyService, connectTransport);
  const questClient = createClient(QuestService, connectTransport);

  const guestClient = {
    entry: async (userName: string) => {
      const entryResponse = await entryClient.entry({ userName: userName });
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
