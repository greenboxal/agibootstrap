import {TypegenClient} from "./client";
import {JSONSchema7, JSONSchema7Definition} from "json-schema";
import {CodeWriter, DefinedType, FileGenerator, PackageGenerator, PackageGeneratorOptions} from "./codewriter";
import {TypeName, parseMangledName, mangleName} from "../typesystem/TypeName";
import * as path from "path";

type JSONSchemaType = "string" | "number" | "integer" | "boolean" | "object" | "array" | "null" | "function" | "interface"
type JSONSchema = Omit<JSONSchema7, "type"> & {
    type: JSONSchemaType,
    prefixItems?: JSONSchema[],
}

const NAME_MAP: Record<string, string> = {
    "string": "PrimitiveTypes.String",
    "number": "PrimitiveTypes.Number",
    "boolean": "PrimitiveTypes.Boolean",
    "object": "PrimitiveTypes.Object",
    "array": "PrimitiveTypes.Array",
    "slice": "PrimitiveTypes.Array",
    "null": "PrimitiveTypes.Null",
    "int": "PrimitiveTypes.Integer",
    "uint": "PrimitiveTypes.UnsignedInteger",
    "int64": "PrimitiveTypes.Integer",
    "uint64": "PrimitiveTypes.UnsignedInteger",
    "float32": "PrimitiveTypes.Float32",
    "float64": "PrimitiveTypes.Float64",
    "func": "PrimitiveTypes.Func",
    "interface": "PrimitiveTypes.Any",
    "ptr": "PrimitiveTypes.Pointer",
}
const fetchSpec = async () => {
    const client = new TypegenClient()

    return await client.getJsonSchema() as JSONSchema
}

type GenericTypeTracker = {
    definedType: DefinedType,
    genericName: TypeName,
    instances: Map<TypeName, JSONSchema>,
}

class TypeGenerator extends PackageGenerator {
    private genericInstances: Record<string, GenericTypeTracker> = {}

    constructor(public schema: JSONSchema, options: PackageGeneratorOptions) {
        super(options)

        this.defineTypeImport([
            "makeSchema",
            "defineFunction",
            "Func",
            "Type",
            "ArrayOf",
            "MapOf",
            "SequenceOf",
            "PrimitiveTypes",
            "ForwardRef",
            "ResolveForwardRef",
            "TypeSymbol",
        ], "@psidb/psidb-sdk/client/schema")

        this.defineTypeImport([
            "makeInterface",
        ], "@psidb/psidb-sdk/client/iface")
    }

    addSchema(schema: JSONSchema) {
        for (const [name, type] of Object.entries(schema.$defs ?? {})) {
            if (typeof type === "boolean") {
                continue
            }

            this.defineType(name, type as JSONSchema)
        }
    }

    defineType(name: string, type: JSONSchema) {
        if (this.types[name]) {
            return
        }

        if (!type.$id) {
            type.$id = name
        }

        const parsedName = parseMangledName(name)
        const genericName = parsedName.generic().toString()

        if (NAME_MAP[genericName]) {
            return
        }

        const def = new DefinedType(
            name,
            parsedName.name,
            this.options.importPrefix + "/" + parsedName.pkg + "/" + parsedName.name,
            parsedName.pkg + "/" + parsedName.name,
        )

        if (parsedName.isGeneric) {
            if (!this.genericInstances[genericName]) {
                this.genericInstances[genericName] = {
                    definedType: def,
                    genericName: parseMangledName(genericName),
                    instances: new Map(),
                }
            }

            this.genericInstances[genericName].instances.set(parsedName, type)

            this.defineTypeAndEmit(def, (w, gen) => {})

            return
        }

        this.defineTypeAndEmit(def, (w, gen) => {
            this.writeExport(gen, w, type, parsedName, def)
        })
    }

    writeExport(gen: FileGenerator, w: CodeWriter, type: JSONSchema, parsedName: TypeName, def: DefinedType) {
        const isClass = type.type === "object" && Object.keys(type.patternProperties || {}).length == 0

        this.useFile(path.join(path.dirname(def.definitionFile), "index"), (gen) => {
            gen.preamble.appendLine(`export { ${def.importName} } from "./${path.basename(def.definitionFile, ".tsx")}"`)
        })

        w.append(`export `)

        if (parsedName.isGeneric) {
            gen.addTypeImport("Type", "@psidb/psidb-sdk/client/schema")

            w.append(`function ${def.importName}<`)
            for (let i = 0; i < (parsedName.inParameters?.length || 0); i++) {
                w.append(`T${i} extends Type`)

                if (i < parsedName.inParameters!.length - 1) {
                    w.append(", ")
                }
            }
            w.append(`>(`)
            for (let i = 0; i < (parsedName.inParameters?.length || 0); i++) {
                w.append(`t${i}: T${i}`)

                if (i < parsedName.inParameters!.length - 1) {
                    w.append(", ")
                }
            }
            w.appendLine(`) {`)
            w.indent()
            w.append(`return `)
        }

        if (isClass) {
            w.append(`class ${def.importName} extends `)
        } else if (!parsedName.isGeneric) {
            w.append(`const ${def.importName} = `)
        }

        this.writeType(gen, w, type, parsedName)

        if (isClass) {
            w.appendLine(` {}`)
        }

        if (parsedName.isGeneric) {
            w.appendLine(``)
            w.dedent()
            w.append(`}`)
        } else if (!isClass) {
            w.appendLine(`;`)
        }
    }

    writeRef(gen: FileGenerator, w: CodeWriter, ref: string) {
        if (ref.startsWith("#/$g/")) {
            const index = ref.replace("#/$g/", "")
            w.append("t" + index)
            return
        }

        const name = decodeURIComponent(ref.replace("#/$defs/", ""))
        const typeName = parseMangledName(name)

        this.writeTypeRef(gen, w, typeName)
    }

    writeTypeRef(gen: FileGenerator, w: CodeWriter, typeName: TypeName) {
        const mangled = mangleName(typeName)
        const genericName = typeName.generic().toString()

        if (gen.declarationStack.some((x) => x == genericName)) {
            const ref = gen.addForwardRef(typeName.name)
            w.append(ref)
            return
        }

        if (NAME_MAP[genericName]) {
            gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")

            w.append(NAME_MAP[genericName])
        } else {
            gen.addTypeImport(mangled)

            w.append(typeName.name)
        }

        if (typeName.inParameters?.length || (typeName.pkg == "" && typeName.name == "func")) {
            w.append(`(`)
            for (let i = 0; i < (typeName.inParameters?.length || 0); i++) {
                this.writeTypeRef(gen, w, typeName.inParameters![i])

                if (i < typeName.inParameters!.length - 1) {
                    w.append(", ")
                }
            }
            w.append(")")
        }

        if (typeName.outParameters?.length || (typeName.pkg == "" && typeName.name == "func")) {
            w.append("(")
            for (let i = 0; i < (typeName.outParameters?.length || 0); i++) {
                this.writeTypeRef(gen, w, typeName.outParameters![i])

                if (i < typeName.outParameters!.length - 1) {
                    w.append(", ")
                }
            }
            w.append(")")
        }
    }

    writeType(gen: FileGenerator, w: CodeWriter, type: JSONSchema, name: TypeName | undefined = undefined) {
        if (type.$ref) {
            this.writeRef(gen, w, type.$ref)
            return
        }

        if (type.type === "array") {
            if (type.prefixItems && type.prefixItems.length > 0) {
                gen.addTypeImport("SequenceOf", "@psidb/psidb-sdk/client/schema")

                w.append(`SequenceOf(`)
                for (let i = 0; i < type.prefixItems.length; i++) {
                    this.writeType(gen, w, type.prefixItems[i] as JSONSchema)

                    if (i < type.prefixItems.length - 1) {
                        w.append(", ")
                    }
                }
                w.append(`)`)
            } else {
                gen.addTypeImport("ArrayOf", "@psidb/psidb-sdk/client/schema")

                w.append(`ArrayOf(`)
                if (type.items) {
                    this.writeType(gen, w, type.items as JSONSchema)
                }
                w.append(`)`)
            }
            return
        } else if (type.type === "object") {
            const patterns = Object.entries(type.patternProperties || {})

            if (patterns.length > 0) {
                if (Object.keys(type.properties || {}).length > 0) {
                    throw new Error("Cannot have both properties and patternProperties")
                }

                if (patterns.length > 1) {
                    throw new Error("Cannot have more than one patternProperty")
                }

                if (patterns[0][0] !== ".*") {
                    throw new Error("patternProperty must be .*")
                }

                gen.addTypeImport("MapOf", "@psidb/psidb-sdk/client/schema")
                gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")

                w.append(`MapOf(PrimitiveTypes.String, `)
                this.writeType(gen, w, patterns[0][1] as JSONSchema)
                w.append(`)`)
            } else {
                gen.addTypeImport("makeSchema", "@psidb/psidb-sdk/client/schema")

                w.append(`makeSchema(`)
                w.append(JSON.stringify(type.$id || ""))
                w.appendLine(`, {`)
                w.useIndent((w) => {
                    for (const [fieldName, field] of Object.entries(type.properties ?? {})) {
                        w.appendIdentifier(fieldName)
                        w.append(": ")
                        this.writeType(gen, w, field as JSONSchema)
                        w.appendLine(",")
                    }
                })
                w.append(`})`)
            }
        } else if (type.type === "interface") {
            gen.addTypeImport("makeInterface", "@psidb/psidb-sdk/client/iface")

            w.appendLine(`makeInterface({`)
            w.useIndent((w) => {
                w.appendIdentifier("name")
                w.append(": ")
                w.append(JSON.stringify(type.$id || ""))
                w.appendLine(",")

                if (Object.entries(type.properties ?? {}).length > 0) {
                    w.appendIdentifier("methods")
                    w.appendLine(": {")
                    w.useIndent((w) => {
                        for (const [fieldName, field] of Object.entries(type.properties ?? {})) {
                            w.appendIdentifier(fieldName)
                            w.append(": ")
                            this.writeType(gen, w, field as JSONSchema)
                            w.appendLine(",")
                        }
                    })
                    w.appendLine("},")
                }
            })
            w.append(`})`)
        } else if (type.type === "function") {
            if (!type.properties) {
                throw new Error("Function field must have properties")
            }

            gen.addTypeImport("defineFunction", "@psidb/psidb-sdk/client/schema")

            w.append(`defineFunction(`)
            if (type.properties?.request_type) {
                this.writeType(gen, w, type.properties?.request_type as JSONSchema)
            }
            w.append(")(")
            if (type.properties?.response_type) {
                this.writeType(gen, w, type.properties?.response_type as JSONSchema)
            }
            w.append(")")
        } else if (type.type === "string") {
            gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")
            w.append(`PrimitiveTypes.String`)
        } else if (type.type === "integer") {
            gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")
            w.append(`PrimitiveTypes.Integer`)
        } else if (type.type === "number") {
            gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")
            w.append(`PrimitiveTypes.Float64`)
        } else if (type.type === "boolean") {
            gen.addTypeImport("PrimitiveTypes", "@psidb/psidb-sdk/client/schema")
            w.append(`PrimitiveTypes.Boolean`)
        }
    }

    emit() {
        this.resolveGenerics()

        return super.emit()
    }

    private resolveGenerics() {
        for (const instance of Object.values(this.genericInstances)) {
            if (instance.instances.size == 0) {
                continue
            }

            this.defineTypeAndEmit(instance.definedType, (w, gen) => {
                const entries= Array.from(instance.instances.entries())

                if (entries.length == 1) {
                    const [name, schema] = entries[0]

                    this.writeExport(gen, w, schema, name, instance.definedType)
                } else {
                    const [nameA, schemaA] = entries[0]
                    const [nameB, schemaB] = entries[1]

                    const genericSchema = this.generify(nameA, schemaA, nameB, schemaB)

                    this.writeExport(gen, w, genericSchema, nameA, instance.definedType)
                }
            })
        }
    }

    generify(rootNameA: TypeName, rootSchemaA: JSONSchema, rootNameB: TypeName, rootSchemaB: JSONSchema): JSONSchema {
        const base = JSON.parse(JSON.stringify({
            ...rootSchemaA,
            ...rootSchemaB,
        }))

        const findCandidatePaths = (name: TypeName, root: TypeName, onHit: (i: number[]) => void) => {
            if (name.equals(root)) {
                onHit([])
                return
            }

            const candidates = [
                ...(root.inParameters || []),
                ...(root.outParameters || []),
            ]

            const index = candidates.findIndex((i) => i.equals(name))

            if (index >= 0) {
                onHit([index])
                return
            }

            for (let i = 0; i < candidates.length; i++) {
                findCandidatePaths(name, candidates[i], (j) => {
                    onHit([i, ...j])
                })
            }
        }

        const unifyNames = (typeA: TypeName, schemaA: JSONSchema, typeB: TypeName, schemaB: JSONSchema): JSONSchema7Definition => {
            let inIndexA: number[][] = []
            let inIndexB: number[][] = []

            findCandidatePaths(typeA, rootNameA, (i) => {
                inIndexA.push(i)
            })

            findCandidatePaths(typeB, rootNameB, (i) => {
                inIndexB.push(i)
            })

            if (inIndexA.length == 0 || inIndexB.length == 0) {
                throw new Error("Could not unify names")
            }

            const setA = new Set(inIndexA.map((i) => i.join(",")))
            const setB = new Set(inIndexB.map((i) => i.join(",")))
            const intersection = new Set([...setA].filter(x => setB.has(x)))

            if (intersection.size == 0) {
                throw new Error("Could not unify names")
            }

            const indexes = Array.from(intersection.values()).map((i) => i.split(",").map((j) => parseInt(j)))

            return {
                $ref: "#/$g/" + indexes[0].join("/")
            } as JSONSchema7Definition
        }

        const unifySchemas = (a: JSONSchema, b: JSONSchema): JSONSchema7Definition => {
            if (!a) {
                return b as JSONSchema7Definition
            }

            if (!b) {
                return a as JSONSchema7Definition
            }

            if (a == b) {
                return a as JSONSchema7Definition
            }

            if (((a.$id && b.$id) && (a.$id == b.$id)) || ((a.$ref && b.$ref) && (a.$ref == b.$ref))) {
                return a as JSONSchema7Definition
            }

            const anameStr = (a.$id || a.$ref || "").replace("#/$defs/", "")
            const bnameStr = (b.$id || b.$ref || "").replace("#/$defs/", "")

            const aname = parseMangledName(anameStr)
            const bname = parseMangledName(bnameStr)

            if (aname.equals(bname)) {
                return a as JSONSchema7Definition
            }

            return unifyNames(aname, a, bname, b)
        }

        if (base.type === "object") {
            const propNames = new Set(Object.keys(rootSchemaA.properties || {}).concat(Object.keys(rootSchemaB.properties || {})))

            base.properties = {}

            for (const key of propNames) {
                const propA = (rootSchemaA.properties?.[key] || {}) as JSONSchema
                const propB = (rootSchemaB.properties?.[key] || {}) as JSONSchema

                base.properties[key] = unifySchemas(propA, propB)
            }

            if (base.patternProperties && Object.keys(base.patternProperties).length > 0) {
                const patterns = new Set(Object.keys(rootSchemaA.patternProperties || {}).concat(Object.keys(rootSchemaB.patternProperties || {})))

                base.patternProperties = {}

                for (const key of patterns) {
                    const propA = (rootSchemaA.patternProperties?.[key] || {}) as JSONSchema
                    const propB = (rootSchemaB.patternProperties?.[key] || {}) as JSONSchema

                    base.patternProperties[key] = unifySchemas(propA, propB)
                }

            }
        } else if (base.type === "array") {
           base.items = unifySchemas(rootSchemaA.items as JSONSchema, rootSchemaB.items as JSONSchema)
        }

        return base
    }
}

async function main() {
    const spec = await fetchSpec()
    const generator = new TypeGenerator(spec, {
        outputPath: "./src/__generated__",
        importPrefix: "@psidb/psidb-sdk/types",
    })

    generator.addSchema(spec)
    generator.emit()
}

main().catch(console.log);
