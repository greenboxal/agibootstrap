import {
    BaseType,
    Config,
    createFormatter,
    createParser,
    createProgram,
    FunctionType,
    NodeParser,
    SchemaGenerator,
    SubTypeFormatter,
    TypeFormatter,
} from "ts-json-schema-generator";
import JsonRefParser from "@apidevtools/json-schema-ref-parser";

import * as TJS from "typescript-json-schema";
import {JSONSchema7 as JSONSchema} from "json-schema";
import {Schema, SchemaBuilder, StructMember, Type} from "./api";

class FunctionTypeFormatter implements SubTypeFormatter {
    supportsType(type: BaseType) {
        return type instanceof FunctionType;
    }

    getDefinition(_type: BaseType) {
        return {}
    }

    getChildren(_type: BaseType) {
        return []
    }
}

export class SchemaCompiler {
    private program: ReturnType<typeof createProgram>;
    private parser: NodeParser;
    private generator: SchemaGenerator;
    private formatter: TypeFormatter;

    constructor(
        protected basePath: string,

        protected config: Config
    ) {
        this.program = createProgram(config);
        this.parser = createParser(this.program, this.config)

        this.formatter = createFormatter(config, (fmt, _circularReferenceTypeFormatter) => {
            fmt.addTypeFormatter(new FunctionTypeFormatter());
        });

        this.generator = new SchemaGenerator(this.program, this.parser, this.formatter, config);
    }

    compileSchema2(): JSONSchema {
        const settings = {
            required: true,
        };

        const compilerOptions = require(this.config.tsconfig!).compilerOptions;

        const program = TJS.getProgramFromFiles(
            [this.config.path!],
            compilerOptions,
            this.basePath
        );

        const generator = TJS.buildGenerator(program, settings)!;

        const schema = generator.getSchemaForSymbol("ModuleInterface");

        schema.$id = this.config.schemaId

        return schema as JSONSchema
    }

    async compileSchema(): Promise<Schema> {
        const schema = this.generator.createSchema("ModuleInterface")

        const schemaString = JSON.stringify(schema)
        const parsedSchema = JSON.parse(schemaString) as JSONSchema
        const fixedSchema = await this.fixSchema(parsedSchema, this.config.schemaId)

        return fixedSchema
    }

    private async fixSchema(schema: JSONSchema, nameHint?: string): Promise<JSONSchema> {
        const parser = new JsonRefParser()
        const parsed = await parser.parse(schema)
        const bundled = await parser.bundle(this.config.schemaId!, parsed, {})
        const rebased = this.rebaseSchema(bundled as JSONSchema)

        schema = {...rebased} as JSONSchema

        if (!schema.$id) {
            schema.$id = nameHint
        }

        if (!schema.$schema) {
            schema.$schema = "https://json-schema.org/draft/2020-12/schema"
        }

        return schema;
    }

    private rebaseSchema(schema: JSONSchema): JSONSchema {
        if (schema.$ref) {
            return {
                $ref: this.rebaseRef(schema, schema.$ref)
            }
        }

        schema = {...schema}

        switch (schema.type) {
            case "object":
                if (schema.properties) {
                    const newProps: Record<string, JSONSchema> = {}

                    Object
                        .entries(schema.properties)
                        .forEach(async ([key, value]) => {
                            newProps[key] = this.rebaseSchema(value as JSONSchema)
                        })

                    schema.properties = newProps
                }
                break;
            case "array":
                if (schema.items) {
                    schema.items = this.rebaseSchema(schema.items as JSONSchema)
                }
                break;
        }

        if (schema.definitions || schema.$defs) {
            schema.$defs = {
                ...schema.definitions,
                ...schema.$defs,
            }

            schema.definitions = undefined
        }

        if (schema.$defs) {
            const newProps: Record<string, JSONSchema> = {}

            Object
                .entries(schema.$defs)
                .forEach(async ([key, value]) => {
                    newProps[key] = this.rebaseSchema(value as JSONSchema)
                })

            schema.$defs = newProps
        }

        return schema
    }

    private rebaseRef(parsed: JSONSchema, ref: string) {
        const [base, path] = ref.split("#", 2)

        if (base && base != this.config.schemaId) {
            return ref
        }

        const decodedPath = decodeURIComponent(path)

        const name =
            decodedPath.startsWith(DEFINITIONS_PATH) ? decodedPath.substring(DEFINITIONS_PATH_LEN) :
                decodedPath.startsWith(DEFS_PATH) ? decodedPath.substring(DEFS_PATH_LEN) :
                    decodedPath

        return `#/$defs/${name}`;
    }

    private async convertSchema(fixedSchema: JSONSchema): Promise<Schema> {
        const builder = new SchemaBuilder(fixedSchema.$id!)
        const resolvedRefs = await JsonRefParser.resolve(fixedSchema)

        const convertType = (name: string, schema: JSONSchema): Type => {
            switch (schema.type) {
                case "object":
                    if (schema.properties) {
                        const members: Record<string, StructMember> = {}

                        Object
                            .entries(schema.properties)
                            .forEach(([key, value]) => {
                                const memberTypeName = `${name}_${key}`
                                const typ = processType("", value as JSONSchema)

                                members[key] = {
                                    name: key,
                                    type: typ.name,
                                    required: schema.required?.includes(key) ?? false,
                                    nullable: true,
                                }
                            })

                        return {
                            name: name,
                            primitive_kind: "Struct",
                            members: members,
                        }
                    } else {
                        return {
                            name: name,
                            primitive_kind: "Struct",
                            members: {},
                        }
                    }


                case "array":
                    const element = processType(name, schema.items as JSONSchema)

                    return {
                        name: name,
                        primitive_kind: "List",
                        element_type: element.name,
                    }

                case "string":
                    return {
                        name: name || "_rt_.string",
                        primitive_kind: "String"
                    }

                case "number":
                    return {
                        name: name,
                        primitive_kind: "Number"
                    }

                case "boolean":
                    return {
                        name: name,
                        primitive_kind: "Boolean"
                    }

                case "integer":
                    return {
                        name: name,
                        primitive_kind: "Integer"
                    }

                case "null":
                    return {
                        name: name,
                        primitive_kind: "Null"
                    }

                default:
                    throw new Error(`Unsupported type: ${schema.type}`)
            }
        }

        const processType = (name: string, schema: JSONSchema): Type => {
            if (schema.$ref) {
                const resolved = resolvedRefs.get(schema.$ref)

                if (!resolved) {
                    throw new Error(`Failed to resolve ${schema.$ref}`)
                }

                const [_, path] = schema.$ref.split("#", 2)
                const lastSlash = path.lastIndexOf("/")
                const refName = path.substring(lastSlash + 1)

                return processType(refName, resolved as JSONSchema)
            }

            if (name == "" && schema.$id) {
                name = schema.$id
            }

            const existing = builder.resolveType(name)

            if (existing) {
                return existing
            }

            const t = convertType(name, schema)

            builder.addType(t)

            return t
        }

        for (const key of Object.keys(fixedSchema.$defs!)) {
            const def = fixedSchema.$defs![key]

            processType(key, def as JSONSchema)
        }

        return builder.build()
    }
}

const DEFINITIONS_PATH = "/definitions/"
const DEFINITIONS_PATH_LEN = DEFINITIONS_PATH.length
const DEFS_PATH = "/$defs/"
const DEFS_PATH_LEN = DEFS_PATH.length
