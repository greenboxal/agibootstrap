import crypto from "crypto";
import { resolve, relative } from "path";
import fs from "fs";
import {command, run, string, optional, option, positional} from 'cmd-ts';
import {PackageManifestBuilder, ModuleManifestBuilder} from "./manifest";
import {SchemaCompiler} from "./schema";

type AssetManifest = {
    files: Record<string, string>,
    entrypoints: string[],
}

const hashInputFile = (filename: string): string => {
    const hash = crypto.createHash('sha256');
    const buffer = fs.readFileSync(filename);
    hash.update(buffer);
    return hash.digest('hex');
}

const app = command({
    name: 'psidb-module-compiler',

    args: {
        source: positional({ type: string, displayName: 'entrypoint' }),
        context: option({ type: optional(string), long: 'context', short: "C", defaultValue: () => "." }),
        output: option({ type: string, long: 'output', short: 'o' }),
        build: option({ type: string, long: 'build', short: 'b' }),
        entrypoint: option({ type: string, long: 'entrypoint', short: 'e', defaultValue: () => "main" }),
        packageName: option({ type: string, long: 'pkg-name', short: 'n' }),
        tsconfig: option({ type: string, long: 'ts-config', defaultValue: () => "./tsconfig.json" }),
    },

    handler: (args) => {
        if (!args.context) {
            args.context = "."
        }

        args.context = (args.context)

        const schemaCompiler = new SchemaCompiler(args.context, {
            schemaId: args.packageName,
            path: resolve(args.context, args.source),
            tsconfig: resolve(args.context, args.tsconfig),
            type: "*",
            jsDoc: "extended",
            expose: "all",
            topRef: false,
            discriminatorType: "json-schema",

            sortProps: true,
            strictTuples: false,
            skipTypeCheck: false,
            encodeRefs: true,
            minify: false,
            extraTags: [],
            additionalProperties: false,
        });

        const schema = schemaCompiler.compileSchema()

        const builder = new PackageManifestBuilder()
        builder.withName(args.packageName)

        const assetManifest = JSON.parse(fs.readFileSync(args.build + "/asset-manifest.json", "utf-8")) as AssetManifest

        for (const entrypoint of assetManifest.entrypoints) {
            const fileName = resolve(args.build, entrypoint)
            const hash = hashInputFile(fileName)
            const modName = assetManifest.files[fileName]

            builder.withFile(relative(args.output, fileName), hash)

            const modBuilder = new ModuleManifestBuilder()

            modBuilder.withName(modName)
            modBuilder.withEntrypoint(entrypoint)

            if (entrypoint == args.entrypoint) {
                modBuilder.withName(args.packageName)
                modBuilder.withSchemaDefinition(schema)
            }

            builder.withModule(modBuilder.build())
        }

        const manifest = builder.build()
        const serializedManifest = JSON.stringify(manifest, null, 2)

        fs.mkdirSync(args.output, {recursive: true})
        fs.writeFileSync(args.output + "/manifest.psidb-module.json", serializedManifest)
    },
});

run(app, process.argv.slice(2)).catch(console.log);
