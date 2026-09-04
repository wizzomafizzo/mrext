import { WebSocketHook } from "react-use-websocket/dist/lib/types";
import useWebSocket from "react-use-websocket";
import { ServerState, useServerStateStore } from "../lib/store";
import { getWsEndpoint } from "../lib/api";

function handleMessage(event: MessageEvent, serverState: ServerState) {
  const msg = event.data;
  const idx = msg.indexOf(":");
  const cmd = msg.substring(0, idx);
  const data = msg.substring(idx + 1);

  switch (cmd) {
    case "indexStatus":
      const args = data.split(",");
      if (args.length === 5) {
        const [ready, indexing, totalSteps, currentStep, currentDesc] = args;
        serverState.setSearch({
          ready: ready === "y",
          indexing: indexing === "y",
          totalSteps: parseInt(totalSteps),
          currentStep: parseInt(currentStep),
          currentDesc: currentDesc,
        });
      }
      break;
    case "gameRunning":
      serverState.setActiveGame(data);
      break;
    case "coreRunning":
      serverState.setActiveCore(data);
      break;
  }
}

export default function useWs(): WebSocketHook {
  const serverState = useServerStateStore();

  return useWebSocket(getWsEndpoint(), {
    onMessage: (event) => handleMessage(event, serverState),
    shouldReconnect: () => true,
    share: true,
  });
}
