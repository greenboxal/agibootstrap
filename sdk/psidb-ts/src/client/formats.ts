import {CachedNode, NodeFormatDefinition, NodeFormatWithCast} from "../../psidb-webui-sdk/data";
import {NodeFormat} from "../../psidb-webui-sdk/psi/nodes";

class NodeFormatImpl<T> implements NodeFormatDefinition<T> {
    get format(): string { return this.props.format }
    get id(): string { return this.props.id }
    get name(): string { return this.props.name }
    get view(): string { return this.props.view }

    constructor(
        private props: NodeFormat,
        private readonly loader?: (node: CachedNode, data: string) => Promise<T>,
        private readonly encoder?: (data: T) => Promise<string>,
    ) {}

    async load(cachedNode: CachedNode, data: string): Promise<T> {
        if (this.loader) {
            return this.loader(cachedNode, data)
        }

        return data as T
    }

    async encode(data: T): Promise<string> {
        if (this.encoder) {
            return this.encoder(data)
        }

        return data as unknown as string
    }
}

class JsonFormatImpl<T> extends NodeFormatImpl<T> {
    constructor(props: NodeFormat) {
        super(props, async (node, data) => {
            return JSON.parse(data)
        }, async (data) => {
            return JSON.stringify(data)
        })
    }

    as<T>() {
        return this as unknown as NodeFormatDefinition<T>
    }

    getter() {
        const getter = <T extends any>() => this.as<T>()

        Object.setPrototypeOf(getter, this)

        return getter as NodeFormatWithCast<T>
    }
}

export const NodeFormats = {
    Markdown: new NodeFormatImpl({
        id: "markdown",
        name: "Markdown",
        view: "",
        format: "text/markdown",
    }),
    HTML: new NodeFormatImpl({
        id: "html",
        name: "HTML",
        view: "",
        format: "text/html",
    }),
    Text: new NodeFormatImpl({
        id: "text",
        name: "Text",
        view: "",
        format: "text/plain",
    }),
    JSON: new JsonFormatImpl({
        id: "json",
        name: "JSON",
        view: "",
        format: "application/json",
    }).getter(),
}

export const GlobalNodeFormats: NodeFormatDefinition<any>[] = [
    NodeFormats.Markdown,
    NodeFormats.HTML,
    NodeFormats.Text,
    NodeFormats.JSON,
]
