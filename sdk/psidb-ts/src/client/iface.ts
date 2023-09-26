import Case from "case";
import Client from "./client"
import {TypeSymbol, FunctionType} from "@psidb/sdk/client/schema";

export type MethodDefinition<TReq, TRes> = {
    requestType: TReq;
    responseType: TRes;
}

export type MethodStub<TReq, TRes> = (request: TReq) => Promise<TRes>

export type InterfaceDefinition = {
    name: string;
    methods: {
        [methodName: string]: MethodDefinition<any, any> | MethodStub<any, any> | FunctionType<any, any>;
    }
}

export type MethodDefinitionFromStub<T extends MethodStub<any, any>> = {
    requestType: Parameters<T>[0];
    responseType: ReturnType<T> extends Promise<infer TRes> ? TRes : ReturnType<T>;
}

export type InterfaceMethod<Def extends InterfaceDefinition, Name extends keyof Def["methods"]>
    = Def["methods"][Name] extends MethodDefinition<any, any>
    ? Def["methods"][Name]
    : (
        Def["methods"][Name] extends MethodStub<any, any>
        ? MethodDefinitionFromStub<Def["methods"][Name]>
            : never
        );

export type StubFromMethodDefinition<Def extends MethodDefinition<any, any>> =
    Def extends MethodDefinition<infer TReq, infer TRes>
    ? (
            TReq extends void
                ?(path: string) => Promise<TRes>
                :(path: string, request: TReq) => Promise<TRes>
            )
        : never;

export type InterfaceImplementation<T extends InterfaceDefinition> = {
    [methodName in keyof T["methods"]]: StubFromMethodDefinition<InterfaceMethod<T, methodName>>;
}

export type StaticStubFromMethodDefinition<Def extends MethodDefinition<any, any>> =
    Def extends MethodDefinition<infer TReq, infer TRes>
        ? (
            TReq extends void
                ?(client: Client, path: string) => Promise<TRes>
                :(client: Client, path: string, request: TReq) => Promise<TRes>
            )
        : never;

export type InterfaceStaticImplementation<T extends InterfaceDefinition> = {
    [methodName in keyof T["methods"]]: StaticStubFromMethodDefinition<InterfaceMethod<T, methodName>>;
}

export type InterfaceImplementationClass<T extends InterfaceDefinition> = {
    [TypeSymbol]: "psidb.typing.Type",

    new(client: Client): InterfaceImplementation<T>
} & InterfaceStaticImplementation<T>

export function makeInterface<T extends InterfaceDefinition>(definition: T): InterfaceImplementationClass<T> {
    const methods = class {
        constructor(public client: Client) {}
    }

    for (const methodName in definition.methods) {
        const actualName = Case.capital(methodName, "")

        Object.defineProperty(methods.prototype, methodName, {
            enumerable: true,
            configurable: false,

            value: function (path: string, request: any) {
                return this.client.callNodeAction(
                    path,
                    definition.name,
                    actualName,
                    request,
                )
            },
        })
    }

    return class extends methods {
        constructor(public client: Client) {
            super(client)
        }
    } as any
}
