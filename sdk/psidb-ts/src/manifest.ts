import {Schema as JsonSchema} from "jsonschema";

export type ModuleManifest = {
    name: string;
    entrypoint: string;
    schema: JsonSchema;
}

export type PackageManifest = {
    name: string;
    modules: ModuleManifest[];
    files: PackageFileManifest[];
}

export type PackageFileManifest = {
    name: string;
    hash: string;
}

export class ModuleManifestBuilder {
    name: string = "";
    entrypoint: string = "";
    schema = new SchemaBuilder()

    withName(name: string): ModuleManifestBuilder {
        this.name = name;
        return this;
    }

    withEntrypoint(entrypoint: string): ModuleManifestBuilder {
        this.entrypoint = entrypoint;
        return this;
    }

    withSchemaDefinition(schema: JsonSchema): ModuleManifestBuilder {
        this.schema.withSchemaDefinition(schema)
        return this;
    }

    build(): ModuleManifest {
        return {
            name: this.name,
            entrypoint: this.entrypoint,
            schema: this.schema.build(),
        }
    }
}

export class SchemaBuilder {
    private readonly definitions: Record<string, JsonSchema> = {}

    withSchemaDefinition(schema: JsonSchema): SchemaBuilder {
        schema = {...schema}

        if (!schema.$id) {
            throw new Error("Schema must have an $id");
        }

        const defs = Object.keys(schema.definitions || {})

        if (defs.length > 0) {
            schema.definitions = this.importSchemaDefinitions(schema, defs)
        }

        this.definitions[schema.$id] = schema;

        return this;
    }

    private importSchemaDefinitions(from: JsonSchema, names: string[]) {
        const refs: Record<string, JsonSchema> = {}

        for (const name of names) {
            refs[name] = {
                "$ref": "#/definitions/" + name,
            }

            if (this.definitions[name]) {
                continue;
            }

            let def = (from.definitions || {})[name]

            if (!def) {
                throw new Error(`Schema definition ${name} not found`)
            }

            if (!def.$id) {
                def = {...def, $id: name}
            }

            this.withSchemaDefinition(def)

            if (def.$id != name) {
                this.definitions[name] = { $ref: "#/definitions/" + def.$id }
            }
        }

        return refs
    }

    build(): JsonSchema {
        return {
            definitions: { ...this.definitions },
        }
    }
}

export class PackageManifestBuilder {
    name: string = "";
    modules: ModuleManifest[] = [];
    files: PackageFileManifest[] = [];

    withName(name: string): PackageManifestBuilder {
        this.name = name;
        return this;
    }

    withModule(module: ModuleManifest | ModuleManifestBuilder): PackageManifestBuilder {
        if (module instanceof ModuleManifestBuilder) {
            module = module.build();
        }

        this.modules.push(module);
        return this;
    }

    withFile(name: string, hash: string): PackageManifestBuilder {
        this.files.push({name, hash});
        return this;
    }


    build(): PackageManifest {
        return {
            name: this.name,
            modules: [...this.modules],
            files: [...this.files],
        }
    }
}
