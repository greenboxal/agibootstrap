import {
    createHttpFetchRequest, defaultJsonErrorMapper,
} from "simple-http-rest-client";
import {HttpMethod, HttpRequest} from "simple-http-request-builder";
import {HttpPromise} from "simple-http-rest-client/build/esm/lib/promise/HttpPromise";
import {fetchClient, HttpFetchClient} from "simple-http-rest-client/build/esm/lib/client/FetchClient";
import {TypedObject} from "../../minstrel/data/types";
import {validateBasicStatusCodes} from "simple-http-rest-client/build/esm/lib/client/FetchStatusValidators";
import {HttpResponse} from "simple-http-rest-client/build/esm/lib/client/HttpResponse";

import {NodeFormats} from "./formats";
import {CreateNodeResult} from "./types";
import {NodeFormatDefinition, NodeFormatWithCast} from "../../psidb-webui-sdk/data";
import {PartialSchemaInstance, SchemaInstance, TypeSymbol} from "./schema";

export const DefaultHttpEndpoint = "https://localhost:22440";
export const DefaultWsEndpoint = "wss://localhost:22440/ws/v1";

export type Link = {
    "/": string
}

export type EdgeInfo = {
    key: string;

    to_index: number;
    to_path: string;
    to_link: Link | null;
}

export type NodeInfo = {
    id: number;
    path: string;
    type: string;
    link: Link | null;
    edges: [EdgeInfo];
}

export type GraphInfo = {
    nodes: { [key: string]: NodeInfo }
}

export type SearchHit = {
    indexedItem: {
        chunkIndex: number;
        embeddings?: number[];
        index: number;
        path: string;
    };

    node: any;

    score: number;
}

export type SearchResponse = {
    results: SearchHit[]
}

export interface ModelOptions {
    model?: string;
    temperature?: number;
    top_p?: number;
    stop_sequence?: string;
    max_tokens?: number;
    presence_penalty?: number;
    frequency_penalty?: number;
    logit_bias?: { [key: string]: number };

    force_function_call?: string;
}

export type ChatMessageKind = 'emit' | 'merge' | 'join' | 'error'

export type ChatMessage = {
    kind: ChatMessageKind

    from: {
        id: string;
        name: string;
        role: string;
    };

    timestamp: string;
    text: string;

    attachments?: string[];

    function_call?: {
        name: string;
        arguments: string;
    };

    model_options?: ModelOptions;
}

export type SerializedNode = {
    index: number;
    parent: number;
    version: number;
    path: string;
    flags: number;
    type: string;
    data?: string;
    link?: Link;
}

export type SerializedEdge = {
    index: number;
    version: number;

    flags: number;
    key: string;

    toIndex: number;
    toPath?: string;
    data?: string;
}

export type SerializedNotification = {
    notifier: string;
    notified: string;
    interface: string;
    action: string;
    params: string;
}

export type JournalOp = "JournalOpInvalid" | "JournalOpBegin" | "JournalOpCommit"
    | "JournalOpRollback" | "JournalOpWrite" | "JournalOpSetEdge"
    | "JournalOpRemoveEdge" | "JournalOpNotify" | "JournalOpConfirm"
    | "JournalOpWait" | "JournalOpSignal";

export const JournalOpValues: Record<number, JournalOp> = {
    0: "JournalOpInvalid",
    1: "JournalOpBegin",
    2: "JournalOpCommit",
    3: "JournalOpRollback",
    4: "JournalOpWrite",
    5: "JournalOpSetEdge",
    6: "JournalOpRemoveEdge",
    7: "JournalOpNotify",
    8: "JournalOpConfirm",
    9: "JournalOpWait",
    10: "JournalOpSignal",
}

export function JournalOperationName(op: string | number): string {
    const value = JournalOpValues[op as number] || op

    return value.substring(9)
}

export type SerializedConfirmation = {
    rid: number;
    xid: number;
    nonce: number;
    ok: boolean;
}

export type SerializedPromise = {
    xid: number;
    nonce: number;
    count: number;
}

export type JournalEntry = {
    ts: number;
    op: number;
    rid: number;
    xid: number;
    inode: number;
    path?: string;
    node?: SerializedNode;
    edge?: SerializedEdge;
    notification?: SerializedNotification;
    confirmation?: SerializedConfirmation;
    promises?: SerializedPromise[];
}

export type RpcResponse<T> = {
    jsonrpc: "2.0";
    id: number;
    result: T;
    error?: any;
}

export type GetMessagesRequest = {
    from?: string,
    to?: string,
    skip_base_history?: boolean,
}

export default class Client {
    private readonly baseUrl: string;
    private readonly wsEndpoint: string;

    private sessionId: string | null = null;

    setSessionId(sessionId: string | null) {
        this.sessionId = sessionId
    }

    constructor(baseUrl: string = DefaultHttpEndpoint, wsEndpoint: string = DefaultWsEndpoint) {
        this.baseUrl = baseUrl
        this.wsEndpoint = wsEndpoint
    }

    public createWebSocket(): WebSocket {
        return new WebSocket(this.wsEndpoint)
    }

    private buildClientForFormat<T>(format: NodeFormatDefinition<T>): HttpFetchClient {
        return <T>(httpRequest: HttpRequest<unknown>) => fetchClient<T>(
            httpRequest,
            validateBasicStatusCodes,
            async (response): Promise<HttpResponse<T>> => {
                return response.text().then(async (body) => {
                    if (!response.ok) {
                        return defaultJsonErrorMapper(response, body) as any;
                    }

                    const loaded = format.load
                        ? (await format.load(null as any, body?.toString() || ""))
                        : body;

                    return {
                        response: loaded as unknown as T,
                    }
                })
            },
        )
    }

    public createHttpFetchRequest<T>(method: HttpMethod, path: string, format: NodeFormatDefinition<T> = NodeFormats.JSON(), httpClient?: HttpFetchClient): HttpRequest<HttpPromise<T>> {
        const client: HttpFetchClient = httpClient  || this.buildClientForFormat(format);

        let req = createHttpFetchRequest<T>(this.baseUrl, method, path, client)

        if (this.sessionId) {
            req = req.headers({
                'X-PsiDB-Session-ID': this.sessionId,
                'Accept': format.format,
            })
        }

        return req
    }

    public async get<T>(url: string, format: NodeFormatDefinition<T> = NodeFormats.JSON()): Promise<T> {
        return this.createHttpFetchRequest<T>(
            "GET",
            url,
            format,
        ).execute()
    }

    public async post<T>(url: string, data: any, format: NodeFormatDefinition<T> = NodeFormats.JSON()): Promise<T> {
        return this.createHttpFetchRequest<T>(
            "POST",
            url,
            format,
        )
            .headers({ 'Content-Type': format.format, })
            .body(format.encode ? await format.encode(data) : data)
            .execute();
    }

    public async rpc<T>(method: string, params: any): Promise<T> {
        const result = await this.post<RpcResponse<T>>(`rpc/v1`, {
            jsonrpc: "2.0",
            id: 1,
            method: method,
            params: params,
        });

        if (result.error) {
            throw result.error;
        }

        return result.result;
    }

    public async getJsonSchema(): Promise<object> {
        return await this.get<object>("v1/openapi.json");
    }

    public async getJournalHead(): Promise<number> {
        const result = await this.rpc<{
            last_record_id: number
        }>("Journal.GetHead", {})

        return result.last_record_id
    }

    public async getJournalEntries(from: number, count: number): Promise<JournalEntry[]> {
        const result = await this.rpc<{
            entries: JournalEntry[]
        }>("Journal.GetEntryRange", {
            from: from,
            count: count,
        })

        return result.entries
    }

    public async getNodeSnapshot(path: string): Promise<NodeInfo> {
        const sanitizedPath = path.split("/").map(encodeURIComponent).join("/")

        return await this.get<NodeInfo>(`v1/psi/${sanitizedPath}?format=json&view=psi-snapshot&depth=0&nested=true`);
    }

    public async getGraphSnapshot(path: string, depth: number = 2): Promise<GraphInfo> {
        const sanitizedPath = path.split("/").map(encodeURIComponent).join("/")

        return await this.get<GraphInfo>(`v1/psi/${sanitizedPath}?format=json&view=psi-snapshot&depth=${depth}`);
    }

    public async getNodeView(path: string, view: string, format: string): Promise<string> {
        const sanitizedPath = path.split("/").map(encodeURIComponent).join("/")
        const url = `v1/psi/${sanitizedPath}?format=${format}&view=${view}&depth=0&nested=true`;

        const res = await fetch(
            `${this.baseUrl}/${url}`,
            {
                method: "GET",
            }
        )

        return res.text()
    }

    public async createNode<T extends SchemaInstance<any>>(path: string, data: T): Promise<CreateNodeResult<T>> {
        return this.post<CreateNodeResult<T>>(`v1/psi/${path}?type=${data[TypeSymbol]}`, data)
    }


    public async createNodeWithType<T>(
        path: string,
        type: string,
        data: Partial<T> = {}
    ): Promise<CreateNodeResult<T>> {
        return this.post<CreateNodeResult<T>>(`v1/psi/${path}?type=${type}`, data)
    }

    public async search(scope: string, query: string, limit: number = 10): Promise<SearchResponse> {
        query = encodeURIComponent(query);
        scope = encodeURIComponent(scope);

        return await this.get<SearchResponse>(`v1/search/?format=json&limit=${limit}&scope=${scope}&query=${query}`);
    }

    public async sendMessage(topic: string, message: ChatMessage): Promise<void> {
        return this.callNodeAction(topic, "IChat", "SendMessage", message)
    }

    public async getMessages(topic: string, request: GetMessagesRequest = {}): Promise<ChatMessage[]> {
        return this.callNodeAction(topic, "IChat", "GetMessages", request)
    }

    public async callNodeAction<T>(
        path: string,
        iface: string,
        action: string,
        params: any,
    ): Promise<T> {
        const result = await this.rpc<{result: T}>("NodeService.CallNodeAction", {
            path: path,
            interface: iface,
            action: action,
            args: params,
        })

        return result.result
    }
};
