import {Type} from "@loopback/repository";

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

export function register(): ModuleRegistration {
    const personInterface: NodeInterface<IPerson> = {
        name: "Person",
        actions: {
            sayHello: {
                name: "sayHello",
                request_type: "",
                response_type: "",
            }
        }
    }

    const personVTable: NodeVTable<PersonSchema, IPerson> = {
        interface: personInterface,
        actions: {
            sayHello: async (ctx, self, req) => {
                return req
            }
        }
    }

    const personType: TypeRegistration<PersonSchema> = {
        name: "vmtest.lol.Person",
        fields: {
            name: {
                name: "name",
                type: "_rt_.string",
            },
        },
        interfaces: {
            IPerson: personVTable,
        }
    }

    return {
        types: [
            personType,
        ]
    }
}
