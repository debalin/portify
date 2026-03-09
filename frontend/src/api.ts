import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { ConverterService } from "./gen/converter/v1/service_connect";

// The transport defines how the client communicates with the server.
// Our Go server is running on port 8080 by default.
const transport = createConnectTransport({
  baseUrl: "/", // Requests are now proxied by Vite
});

// The client provides a type-safe interface matching the protobuf definition.
export const apiClient = createPromiseClient(ConverterService, transport);
