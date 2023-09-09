import {JSONSchema7 as JSONSchema} from "json-schema";
import {Schema} from "./api";

export type ModuleManifest = {
    name: string;
    entrypoint: string;
    schema: Schema;
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
    schema: Partial<Schema> = {}

    withName(name: string): ModuleManifestBuilder {
        this.name = name;
        return this;
    }

    withEntrypoint(entrypoint: string): ModuleManifestBuilder {
        this.entrypoint = entrypoint;
        return this;
    }

    withSchemaDefinition(schema: Schema): ModuleManifestBuilder {
        this.schema = schema
        return this;
    }

    build(): ModuleManifest {
        return {
            name: this.name,
            entrypoint: this.entrypoint,
            schema: this.schema as Schema,
        }
    }
}

export class SchemaBuilder {
    private readonly definitions: Record<string, JSONSchema> = {}

    withSchemaDefinition(schema: JSONSchema): SchemaBuilder {
        schema = {...schema}

        if (!schema.$id) {
            throw new Error("Schema must have an $id");
        }

        const defs = Object.keys(schema.$defs || {})

        if (defs.length > 0) {
            schema.$defs = this.importSchemaDefinitions(schema, defs)
        }

        this.definitions[schema.$id] = schema;

        return this;
    }

    private importSchemaDefinitions(from: JSONSchema, names: string[]) {
        const refs: Record<string, JSONSchema> = {}

        for (const name of names) {
            refs[name] = {
                "$ref": "#/$defs/" + name,
            }

            if (this.definitions[name]) {
                continue;
            }

            let def = (from.$defs || {})[name]

            if (!def) {
                throw new Error(`Schema definition ${name} not found`)
            }

            if (typeof def !== "object") {
                def = {}
            } else {
                def = {...def}
            }

            if (!def.$id) {
                def = {...def, $id: name}
            }

            this.withSchemaDefinition(def)

            if (def.$id != name) {
                this.definitions[name] = { $ref: "#/$defs/" + def.$id }
            }
        }

        return refs
    }

    build(): JSONSchema {
        return {
            $defs: { ...this.definitions },
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
