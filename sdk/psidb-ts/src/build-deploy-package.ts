import { resolve, relative } from "path";
import fs from "fs";
import {command, run, string, optional, option, positional} from 'cmd-ts';
import {PackageManifest} from "./manifest";

type PackageBundle = {
    name: string;
    manifest: PackageManifest;
    files: Record<string, string>;
}

const app = command({
    name: 'psidb-build-deploy-package',

    args: {
        manifest: positional({ type: string, displayName: 'Manifest' }),
        dist: positional({ type: string, displayName: 'Dist Directory' }),
    },

    handler: async (args) => {
        const manifest = JSON.parse(fs.readFileSync(args.manifest).toString()) as PackageManifest
        const files: Record<string, string> = {}

        for (const file of manifest.files) {
            files[file.name] = fs.readFileSync(resolve(args.dist, file.name)).toString()
        }

        const lastSlash = manifest.name.lastIndexOf('/')

        if (lastSlash === -1) {
            throw new Error("Invalid package name")
        }

        const packageBundle: PackageBundle = {
            name: manifest.name.substring(lastSlash + 1),
            manifest: manifest,
            files: files,
        }

        console.log(JSON.stringify(packageBundle))
    },
});

run(app, process.argv.slice(2)).catch(console.log);
