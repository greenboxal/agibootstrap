import { Schema as JsonSchema } from "jsonschema";
export type ModuleManifest = {
    name: string;
    files: ModuleManifestFile[];
    schema: JsonSchema;
};
export type ModuleManifestFile = {
    name: string;
    hash: string;
};
export declare class ManifestBuilder {
    name: string;
    files: ModuleManifestFile[];
    schema: JsonSchema;
    withName(name: string): ManifestBuilder;
    withFile(name: string, hash: string): ManifestBuilder;
    withSchemaDefinition(schema: JsonSchema): ManifestBuilder;
    private importSchemaDefinitions;
    build(): ModuleManifest;
}
