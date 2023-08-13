import React from "react"

export type Type = {
    name: string;
}

export type NodeActionHandler = () => Promise<void>;

export type NodeActionDefinition = {
    name: string;
    requestType: Type;
    responseType: Type;
}

export type NodeInterface = {
    name: string;
    actions: Record<string, NodeActionDefinition>
}

export type NodeVTable = {
    interface: NodeInterface;
    actions: Record<string, NodeActionHandler>;
}

export type NodeType = {
    name: string;
    schema: Type;
    interfaces: Record<string, NodeVTable>;
}

export type DispatchRequest = {
    node: any,
    interface: string,
    action: string,
    request: any,
}

export function dispatch(req: DispatchRequest) {
    React.createElement("a")

    return req.request
}

export function register() {
    return React.createElement("a")
}
