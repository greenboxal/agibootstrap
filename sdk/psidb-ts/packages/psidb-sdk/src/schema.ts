export const TypeSymbol = Symbol("PsiType")
export const PathSymbol = Symbol("PsiPath")
export const SchemaSymbol = Symbol("PsiSchema")

export interface BasicType {
    [TypeSymbol]: "psidb.typing.Type";

    name: string;
}

export interface TypeBase<TName extends string = string> extends Omit<BasicType, "name"> {
    name: TName;
}

export interface ListType<T extends Type> extends TypeBase<`_rt_.slice_QZQZ_${T["name"]}`> {
}

export interface MapType<T extends Type> extends TypeBase<`_rt_.map_QZQZ__rt_.string_QZQZ_${T["name"]}`> {
}

export type NamedBasicType<Name extends string> = Omit<BasicType, "name"> & {
    name: Name;
}

export type Type<T = any> = BasicType & {
    defaultValue?: T,
}

export type NamedType<T = any, Name extends string = string> = Type<T> & NamedBasicType<Name>

export type PsiTypeOf<T> = Type<T>

export type UnwrapType<T extends Type> = T extends Type<infer U> ? U : never

export type FieldListDefinition = {
    [fieldName: string]: Type | ForwardRef<Type, any>;
}

export type SchemaDefinition<T extends string = string> = {
    [TypeSymbol]: T;
} & FieldListDefinition

export type SchemaDefinitionFields<T extends SchemaDefinition, Name extends keyof T> =
    ResolveForwardRef<T[Name]> extends Type<infer U> ? U : never

export type SchemaInstanceFields<T extends SchemaDefinition> = {
    [fieldName in keyof T]: SchemaDefinitionFields<T, fieldName>
}

export interface ISchemaInstance<TSchema extends SchemaClass<TSchema, T>, T extends SchemaDefinition>{
    [TypeSymbol]: T[typeof TypeSymbol];
    [SchemaSymbol]: TSchema;
}

export type SchemaInstance<T extends SchemaDefinition> = {
    [TypeSymbol]: T[typeof TypeSymbol];
} & SchemaInstanceFields<T>

export interface SchemaClass<TSchema extends SchemaClass<TSchema, T>, T extends SchemaDefinition> extends TypeBase {
    new(props?: Partial<SchemaInstance<T>>): ISchemaInstance<TSchema, T>

    schema: T;
}

export type PartialSchemaInstance<T extends SchemaDefinition> = {
    [TypeSymbol]: T[typeof TypeSymbol];
} & Partial<SchemaInstanceFields<T>>

function unwrapForwardRef<T>(ref: Type<T> | ForwardRef<Type<T>, T>): Type<T> {
    return typeof ref === "function" ? ref() : ref
}

export function makeSchema<TName extends string, T extends SchemaDefinition<TName>>(definition: T) {
    const cls = class {
        static [TypeSymbol]: "psidb.typing.Type" = "psidb.typing.Type"
        static schema = definition

        get [TypeSymbol](): T[typeof TypeSymbol] { return definition[TypeSymbol] }
        get [SchemaSymbol](): SchemaClass<typeof cls, T> { return cls }

        constructor(props?: Partial<SchemaInstance<T>>) {

            if (props) {
                Object.assign(this, props)
            }
        }
    }

    for (const fieldName in definition) {
        const def = unwrapForwardRef(definition[fieldName])

        if (typeof def === "object" && def[TypeSymbol] === "psidb.typing.Type") {
            Object.defineProperty(cls.prototype, fieldName, {
                enumerable: true,
                configurable: true,
                writable: true,
                value: def.defaultValue,
            })
        }
    }

    Object.defineProperty(cls, "name", {
        enumerable: true,
        configurable: false,
        writable: false,
        value: definition[TypeSymbol],
    })

    return cls
}

const makePrimitiveType = <T, Name extends string>(name: Name, defaultValue?: T): NamedType<T, Name> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: name,
    defaultValue: defaultValue,
})

const StringType = makePrimitiveType("_rt_.string", "")
const BooleanType = makePrimitiveType("_rt_.bool", false)
const IntType = makePrimitiveType("_rt_.int", 0)
const UnsignedIntType = makePrimitiveType("_rt_.uint", 0)
const Float32Type = makePrimitiveType("_rt_.float32", 0)
const Float64Type = makePrimitiveType("_rt_.float64", 0)

export function ForwardRef<T>(resolve: () => T): T {
    return resolve as T
}

// eslint-disable-next-line @typescript-eslint/no-redeclare
declare type ForwardRef<TType extends Type<T>, T> = () => TType

export type ResolveForwardRef<T> = T extends ForwardRef<infer U, any> ? U : T

export const PrimitiveTypes = {
    String: StringType,
    Boolean: BooleanType,
    Int: IntType,
    Integer: IntType,
    UnsignedInt: UnsignedIntType,
    UnsignedInteger: UnsignedIntType,
    Float32: Float32Type,
    Float64: Float64Type,
}

export const makeListType = <T extends Type>(itemType: T): Type<T[]> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `_rt_.slice_QZQZ_${itemType.name}`,
})

export const makeMapType = <T extends Type>(itemType: T): Type<{ [key: string]: T }> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `_rt_.map_QZQZ__rt_.string_QZQZ_${itemType.name}`,
})

export const makeTuple = <T extends Type>(...itemTypes: T[]): Type<{ [key: string]: T }> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `_rt_.tuple_QZQZ__${itemTypes.map(t => t.name).join("_QZQZ_")}`,
})

export const ArrayOf = <T extends any>(itemType: Type<T>) => makeListType(itemType)
export const MapOf = <T extends any>(itemType: Type<T>) => makeMapType(itemType)
export const SequenceOf = <T extends Array<Type>>(...itemType: T) => makeTuple(...itemType)
