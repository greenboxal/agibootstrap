
export const TypeSymbol = "@type"
export const PathSymbol = Symbol("PsiPath")
export const SchemaSymbol = Symbol("PsiSchema")

export interface BasicType {
    [TypeSymbol]: "psidb.typing.Type";

    name: string;
}

export interface TypeBase<TName extends string = string> extends Omit<BasicType, "name"> {
    name: TName;
}

export interface ListType<TItem extends Type<T>, T extends any> extends Type<T[]> {
    itemType: TItem,
}

export interface MapType<K extends Type<KT>, V extends Type<VT>, KT extends any, VT extends any> extends Type<{
    [key: string]: VT;
}> {
    keyType: K,
    valueType: V,
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
    new(props?: Partial<SchemaInstanceFields<T>>): SchemaInstanceFields<T>

    schema: T;
    defaultValue: SchemaInstanceFields<T>;
}

export type PartialSchemaInstance<T extends SchemaDefinition> = {
    [TypeSymbol]: T[typeof TypeSymbol];
} & Partial<SchemaInstanceFields<T>>

function unwrapForwardRef<T>(ref: Type<T> | ForwardRef<Type<T>, T>): Type<T> {
    return typeof ref === "function" ? ref() : ref
}

export function makeSchema<TName extends string, TFields extends FieldListDefinition>(name: TName, definition: TFields) {
    type T = {
        [k in keyof TFields]: TFields[k]
    } & {
        [TypeSymbol]: TName;
    };

    const cls = class SchemaInstanceClass {
        static [TypeSymbol]: "psidb.typing.Type" = "psidb.typing.Type"

        static schema: T = {
            ...definition,
            [TypeSymbol]: name,
        }

        static defaultValue: SchemaInstanceClass = {} as any

        get [TypeSymbol]() { return name }
        get [SchemaSymbol]() { return cls }

        constructor(props?: Partial<SchemaInstanceFields<T>>) {
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
        value: name,
    })

    cls.defaultValue = new cls()

    return cls as unknown as SchemaClass<any, T>
}

const makePrimitiveType = <T, Name extends string>(name: Name, defaultValue?: T): NamedType<T, Name> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: name,
    defaultValue: defaultValue,
})

const StringType = makePrimitiveType("string", "")
const BooleanType = makePrimitiveType("bool", false)
const IntType = makePrimitiveType("int", 0)
const UnsignedIntType = makePrimitiveType("uint", 0)
const Float32Type = makePrimitiveType("float32", 0)
const Float64Type = makePrimitiveType("float64", 0)

export function ForwardRef<T>(resolve: () => T): T {
    return resolve as T
}

export function indirect<TType extends Type<T>, T extends any>(v: ForwardRef<TType, T>): Type<T> {
    if (typeof v === "function") {
        return v()
    }

    return v
}

declare type ForwardRef<TType extends Type<T>, T extends any> = TType | (() => TType)

export type ResolveForwardRef<T> = T extends ForwardRef<infer U, any> ? U : T

export type ArgumentList<Types extends Type[]> = {
    [k in keyof Types]: Types[k] extends Type<infer U> ? U : never
}

export type FunctionStub<TIn extends Type[], TOut extends Type[]> = (...args: ArgumentList<TIn>) => ArgumentList<TOut>

export interface FunctionType<TIn extends Type[], TOut extends Type[]> extends Type<FunctionStub<TIn, TOut>> {
    inArguments: TIn;
    outArguments: TOut;
}

export function makeFunction<TIn extends Type[], TOut extends Type[]>(inTypes: TIn, outTypes: TOut): FunctionType<TIn, TOut> {
    return {
        [TypeSymbol]: "psidb.typing.Type",
        name: `func(${inTypes.map(t => t.name).join(";")})(${outTypes.map(t => t.name).join(";")})})`,
        inArguments: inTypes,
        outArguments: outTypes,
    }
}

export const PrimitiveTypes = {
    Any: makePrimitiveType<any, "any">("any"),

    String: StringType,
    Boolean: BooleanType,
    Int: IntType,
    Integer: IntType,
    UnsignedInt: UnsignedIntType,
    UnsignedInteger: UnsignedIntType,
    Float32: Float32Type,
    Float64: Float64Type,

    Pointer: <T extends any>(itemType: Type<T> | (() => any)): Type<T> => indirect(itemType),

    Array: <T extends any>(itemType: Type<T>) => makeListType(itemType),

    Map: <K extends any, V extends any>(keyType: Type<K>, valueType: Type<V>) => makeMapType(keyType, valueType),

    Func: <TIn extends Type[]>(...args: TIn) => {
        return <TOut extends Type[]>(...outArgs: TOut) => {
            return makeFunction(args, outArgs)
        }
    },
}

export const makeListType = <T extends any>(itemType: Type<T>): ListType<Type<T>, T> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `slice(${itemType.name})`,
    itemType: itemType,
})

export const makeMapType = <KT extends any, VT extends any>(keyType: Type<KT>, valueType: Type<VT>): MapType<Type<KT>, Type<VT>, KT, VT> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `map(${keyType.name};${valueType.name})`,
    keyType: keyType,
    valueType: valueType,
})

export const makeTuple = <T extends Type>(...itemTypes: T[]): Type<{ [key: string]: T }> => ({
    [TypeSymbol]: "psidb.typing.Type",

    name: `tuple(${itemTypes.map(t => t.name).join(";")})`,
})

export const ArrayOf = makeListType
export const MapOf = makeMapType
export const SequenceOf = <T extends Array<Type>>(...itemType: T) => makeTuple(...itemType)
