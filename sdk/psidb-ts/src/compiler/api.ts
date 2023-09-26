import {KeyOf} from "@loopback/repository";
import {JSONSchema} from "@apidevtools/json-schema-ref-parser/dist/lib/types";

export type PrimitiveKind = KeyOf<typeof PrimitiveKindMap>

export const PrimitiveKindMap: Record<string, number> = {
    "Boolean": 1,
    "Bytes": 2,
    "String": 3,
    "Int": 4,
    "UnsignedInt": 5,
    "Float":  6,
    "List": 7,
    "Map": 8,
    "Struct": 9,
    "Interface": 10,
    "Link": 11,
}

export type TypeOrRef<T = any> = BasicType<T> | string;

export interface BasicType<T = any> {
    name: string;
    primitive_kind: PrimitiveKind;
}

export interface PrimitiveType<T = any> extends BasicType<T> {
    primitive_kind: "Boolean" | "Bytes" | "String" | "Int" | "UnsignedInt" | "Float";
}

export type ListType<T = any> = BasicType<Array<T>> & {
    primitive_kind: "List";
    element_type: TypeOrRef<T>;
}

export type MapType<K extends string | number | symbol = any, V = any> = BasicType<Record<K, V>> & {
    primitive_kind: "Map";
    key_type: TypeOrRef<K>;
    value_type: TypeOrRef<V>;
}

export interface StructMember<T = any> {
    name: string;
    type: TypeOrRef<T>;
    required: boolean;
    nullable: boolean;
}

export type StructType<T = any> = BasicType<T> & {
    primitive_kind: "Struct";
    members: Record<string, StructMember>;
}

export interface InterfaceType<T = any> extends BasicType<T> {
    primitive_kind: "Interface";
}

export interface LinkType<T = any> extends BasicType<T> {
    primitive_kind: "Link";
}

export type Type = (BasicType | ListType | MapType | StructType | InterfaceType | LinkType);

export type Schema = JSONSchema
export interface CustomSchema {
    name: string;
    types: Record<string, Type>;
}

export class SchemaBuilder {
    private validatedTypes = new Set<string>()

    private schema: CustomSchema = {
        name: "",
        types: {},
    }

    constructor(name: string) {
        this.schema.name = name
    }

    addType(type: Type) {
        this.schema.types[type.name] = type
    }

    resolveType(type: TypeOrRef): Type {
        if (typeof type === "string") {
            return this.schema.types[type]
        }

        return type
    }

    validate() {
        const {name, types} = this.schema

        if (!name) {
            throw new Error("Schema name is required")
        }

        if (Object.keys(types).length === 0) {
            throw new Error("Schema must contain at least one type")
        }

        for (const type of Object.values(types)) {
            this.validateType(type)
        }
    }

    build() {
        return this.schema
    }

    private validateType(type: Type) {
        if (this.validatedTypes.has(type.name)) {
            return
        }

        switch (type.primitive_kind) {
            case "List":
                const lt = type as ListType

                this.validateType(this.resolveType(lt.element_type))

                break;
            case "Map":
                const mt = type as MapType

                this.validateType(this.resolveType(mt.key_type))
                this.validateType(this.resolveType(mt.value_type))

                break;
            case "Struct":
                const st = type as StructType

                for (const member of Object.values(st.members)) {
                    this.validateType(this.resolveType(member.type))
                }

                break
        }

        this.validatedTypes.add(type.name)
    }
}