import {createHttpFetchRequest, defaultJsonFetchClient, HttpFetchClient, HttpPromise,} from "simple-http-rest-client";
import {HttpMethod, HttpRequest} from "simple-http-request-builder";


const DefaultHttpEndpoint = "https://localhost:22440";

export class TypegenClient {
    private readonly baseUrl: string;

    constructor(baseUrl: string = DefaultHttpEndpoint) {
        this.baseUrl = baseUrl
    }

    public createHttpFetchRequest<T>(method: HttpMethod, path: string, client: HttpFetchClient): HttpRequest<HttpPromise<T>> {
        let req = createHttpFetchRequest<T>(this.baseUrl, method, path, client)
        return req
    }

    public async get<T>(url: string): Promise<T> {
        return this.createHttpFetchRequest<T>(
            "GET",
            url,
            defaultJsonFetchClient,
        ).execute()
    }

    public async getJsonSchema(): Promise<object> {
        return await this.get<object>("v1/openapi.json");
    }
}