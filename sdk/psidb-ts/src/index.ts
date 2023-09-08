import {ModuleRegistration, NodeActionContext, NodeInterface, NodeVTable, StaticModuleInterface, TypeRegistration} from "./api";

type IPerson = {
    sayHello(ctx: NodeActionContext, req: {pandas?: string}): {};
}

type PersonSchema = {
    name: string;
}

declare const _psidb_tx: any;
declare const _psidb_sp: any;

function test(ctx: any) {
    console.log("Hello from psidb-ts")
}

function testTypes(reg: ModuleRegistration, ip: NodeInterface<IPerson>) {

}

export const psidbModule = {
    test(person: IPerson): PersonSchema {
        return {name: ""}
    },

    lol() {

    },
}

class Module {
    constructor() {
    }

    test(person: IPerson): PersonSchema {
        return {name: ""}
    }

    lol() {

    }
}

export type Act<Name extends string, Req extends any, Res extends any> = {
    name: Name;
    request_type: Req;
    response_type: Res;
}

export interface ModuleInterface {
    actions: {
        test2: Act<"test2", IPerson, PersonSchema>;
        test: {
            name: "test";
            request_type: IPerson;
            response_type: PersonSchema;
        }
    }
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


