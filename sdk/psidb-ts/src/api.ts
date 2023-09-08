export type NodeActionContext = {

}

export type NodeActionHandler<TSelf, TReq, TRes> = (ctx: NodeActionContext, self: TSelf, req: TReq) => TRes | Promise<TRes>;

export type NodeActionDefinition<TReq, TRes> = {
    name: string;
    request_type: TypeRegistration<TReq>["name"];
    response_type: TypeRegistration<TRes>["name"];
}

export type _NodeInterface = {
    [k: string]: NodeActionHandler<any, any, any>
}

export type NodeInterface<TInterface extends _NodeInterface = {}> = {
    name: string;
    actions: {
        [K in keyof TInterface]: NodeActionDefinition<Parameters<TInterface[K]>[1], ReturnType<TInterface[K]>>;
    }
}

export type NodeVTable<TSelf = {}, TInterface extends _NodeInterface = {}> = {
    interface: NodeInterface<TInterface>;
    actions: {
        [K in keyof TInterface]: NodeActionHandler<TSelf, Parameters<TInterface[K]>[1], ReturnType<TInterface[K]>>;
    }
}

export type FieldDefinition = {
    name: string;
    type: string;
    optional?: boolean;
}

export interface TypeRegistration<TFields = {}, TInterfaces extends { [k: string]: _NodeInterface } = {}, TName extends string = ""> {
    name: string;
    fields: { [K in keyof TFields]: FieldDefinition; };
    interfaces: { [K in keyof TInterfaces]: NodeInterface<TInterfaces[K]>; };
}

export type ModuleRegistration = {
    types: TypeRegistration[];
}

type IPerson = {
    sayHello(ctx: NodeActionContext, req: {pandas?: string}): {};
}

type PersonSchema = {
    name: string;
}

declare const _psidb_tx: any;
declare const _psidb_sp: any;

export function test(ctx: any) {
    console.log("Hello from psidb-ts")
}

export function testTypes(reg: ModuleRegistration, ip: NodeInterface<IPerson>) {

}

export type ModuleInterfaceDefinition = {
    [k: string]: (...args: any) => any;
}

//type UnwrapReturnType<T extends (...args: any) => any> = ReturnType<T> extends Promise<infer R> ? R : ReturnType<T>;

export type StaticModuleInterface<Def extends ModuleInterfaceDefinition> = {
    [K in keyof Def]: {
        name: K;
        requestType: {
            0: Parameters<Def[K]>[0];
            1: Parameters<Def[K]>[1];
        };
        returnType: ReturnType<Def[K]>;
    }
}
