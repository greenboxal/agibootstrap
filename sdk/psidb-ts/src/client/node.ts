import {makeSchema, NamedType, PrimitiveTypes, SchemaClass, SchemaDefinition, SchemaInstance, TypeSymbol} from "./schema";
import {InterfaceImplementationClass, makeInterface} from "./iface";

export type NodeInterfaceMap = {
    [ifaceName: string]: InterfaceImplementationClass<any>
}

export type NodeClass<
    TSchema extends SchemaDefinition,
    TIFaces extends NodeInterfaceMap = {}
> = NamedType<NodeInstance<TSchema>, TSchema[typeof TypeSymbol]> & {
    new(props?: Partial<TSchema>): NodeInstance<TSchema, TIFaces>
}

export interface BasicNode {
    getPsiNodeType(): string
    getCanonicalPath(): string
}

export type NodeInstance<
    TSchema extends SchemaDefinition,
    TIFaces extends NodeInterfaceMap = {}
> = BasicNode & SchemaInstance<TSchema> & TIFaces[keyof TIFaces]

export function makeNodeType<TSchema extends SchemaDefinition, TIFaces extends NodeInterfaceMap = {}>(
    schema: SchemaClass<TSchema>,
    ifaces: TIFaces = {} as TIFaces
): NodeClass<TSchema, TIFaces> {
    const prototype: any = schema

    const cls = class extends prototype {

    } as unknown as NodeClass<TSchema>

    Object.defineProperty(cls, TypeSymbol, {
        enumerable: true,
        configurable: false,
        writable: false,
        value: "psidb.typing.Type",
    })

    Object.defineProperty(cls, "name", {
        enumerable: true,
        configurable: false,
        writable: false,
        value: schema.name,
    })

    Object.defineProperty(cls, "schema", {
        enumerable: true,
        configurable: false,
        writable: false,
        value: schema.schema,
    })

    return cls
}

const PlayerInterface = makeInterface({
    name: "IPlayer",
    methods: {
        play: (): Promise<void> => Promise.resolve(),
        pause: (): Promise<void> => Promise.resolve(),
        stop: (): Promise<void> => Promise.resolve(),
    },
})

const SongSchema = makeSchema({
    [TypeSymbol]: "psidb.jukebox.Song",
    name: PrimitiveTypes.String,
})