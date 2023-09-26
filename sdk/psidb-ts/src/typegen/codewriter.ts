import Ramda from "ramda";
import fs from "fs";
import * as path from "path";

export class ImportEntry {
    constructor(public from: string, public name: string) {}
}

export class DefinedType {
    constructor(
        public name: string,
        public importName: string,
        public importFile: string,
        public definitionFile: string = importFile,
    ) {
    }
}

export interface ImportResolver {
    resolveImport(name: string): DefinedType | undefined
}

export class FileGenerator {
    public preamble: CodeWriter = new CodeWriter()
    public footer: CodeWriter = new CodeWriter()

    public declarationStack: string[] = []

    private forwardRefs = new Set<string>()
    private imports = new Set<string>()
    private resolvedImports: Record<string, ImportEntry> = {}
    private types: Record<string, CodeWriter> = {}

    constructor(
        public fileName: string,
        public importPath: string,
        private importResolver: ImportResolver,
    ) {
    }

    addImport(nameOrNames: string | string[], from: string) {
        let names = Array.of(nameOrNames).flat()

        if (from == this.importPath) {
            return
        }

        for (const name of names) {
            this.resolvedImports[name] = new ImportEntry(from, name)
        }
    }

    addTypeImport(nameOrNames: string | string[], from: string = "") {
        let names = Array.of(nameOrNames).flat()

        if (from != "") {
            names = names.map((n) => from + "/" + n)
        }

        for (const name of names) {
            if (this.imports.has(name)) {
                continue
            }

            if (name == this.importPath) {
                continue
            }

            this.imports.add(name)
        }
    }

    addForwardRef(name: string) {
        const ref = `_F[${JSON.stringify(name)}]`

        if (this.forwardRefs.has(name)) {
            return ref
        }

        this.forwardRefs.add(name)

        this.footer.appendLine(`${ref} = ${name};`)

        return ref
    }

    getWriter(name: string) {
        if (!this.types[name]) {
            this.types[name] = new CodeWriter()
        }

        return this.types[name]
    }

    useWriter(name: string, cb: (writer: CodeWriter) => void) {
        this.declarationStack.push(name)

        try {
            cb(this.getWriter(name))
        } finally {
            this.declarationStack.pop()
        }
    }

    private resolveImports() {
        for (const name of this.imports) {
            const type = this.importResolver.resolveImport(name)

            if (!type) {
                console.warn(new Error(`Type '${name}' not found`))
                continue
            }

            this.addImport(type.importName, type.importFile)
        }
    }

    emit(name: string): string | NodeJS.ArrayBufferView {
        this.resolveImports()
        this.resolveRefs()

        const grouped = Ramda.groupBy(([k, v]) => v.from, Object.entries(this.resolvedImports))

        const imports = Object.entries(grouped).map(([from, v]) => {
            if (!v) {
                return ""
            }

            return `import { ${v.map(([k, v]) => v.name).join(", ")} } from "${from}";`
        }).join("\n")

        const types = Object.entries(this.types).map(
            ([name, writer]) => writer.toString()
        ).join("\n\n")

        return imports + "\n\n" + this.preamble.toString() + "\n" + types + this.footer.toString()
    }

    resolveRefs() {
        if (this.forwardRefs.size == 0) {
            return
        }

        this.preamble.appendLine("const _F = {} as any")
    }
}

export type PackageGeneratorOptions = {
    importPrefix: string,
    outputPath: string,
}

export class PackageGenerator implements ImportResolver {
    protected files: Record<string, FileGenerator> = {}
    protected types: Record<string, DefinedType> = {}

    constructor(public options: PackageGeneratorOptions) {
    }


    defineTypeImport(nameOrNames: string[], from: string) {
        const names = Array.of(nameOrNames).flat()

        for (const name of names) {
            const fullName = from + "/" + name

            this.types[fullName] = new DefinedType(fullName, name, from)
        }
    }

    useFile(name: string, cb: (writer: FileGenerator) => void) {
        if (!this.files[name]) {
            this.files[name] = new FileGenerator(name, path.join(this.options.importPrefix, name), this)
        }

        cb(this.files[name])
    }

    useTypeWriter(fileName: string, typeName: string, cb: (writer: CodeWriter, gen: FileGenerator) => void) {
        this.useFile(fileName, (file) => {
            file.useWriter(typeName, (writer) => {
                cb(writer, file)
            })
        })
    }

    defineTypeAndEmit(def: DefinedType, cb: (writer: CodeWriter, gen: FileGenerator) => void) {
        this.types[def.name] = def

        this.useTypeWriter(def.definitionFile, def.name, cb)
    }

    emit() {
        for (const [name, file] of Object.entries(this.files)) {
            const resolved = path.join(this.options.outputPath, name + ".tsx")

            fs.mkdirSync(path.dirname(resolved), { recursive: true })
            fs.writeFileSync(resolved, file.emit(name))
        }
    }

    resolveImport(name: string): DefinedType | undefined {
        return this.types[name]
    }
}

export class CodeWriter {
    private readonly lines: string[] = []
    private tabLevel: number = 0

    constructor(tabLevel: number = 0) {
        this.tabLevel = tabLevel
        this.lines=[""]
    }

    public indent(count: number = 1) {
        this.tabLevel += count
    }

    public dedent(count: number = 1) {
        this.tabLevel -= count
    }

    public useIndent(cb: (w: CodeWriter) => void, count: number = 1) {
        this.indent(count)

        try {
            cb(this)
        } finally {
            this.dedent(count)
        }
    }

    public appendIdentifier(str: string) {
        if (/[^a-zA-Z0-9_]/.test(str)) {
            this.append(JSON.stringify(str))
        } else {
            this.append(str)
        }
    }

    public append(str: string) {
        const lines = str.split("\n")
        const prefix =  " ".repeat(this.tabLevel * 4)

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i]

            if (i > 0) {
                this.lines.push("")
            }

            if (line == "") {
                continue
            }

            if (this.lines[this.lines.length - 1] == "") {
                this.lines[this.lines.length - 1] = prefix
            }

            this.lines[this.lines.length - 1] += line
        }
    }

    public appendLine(line: string) {
        this.append(line + "\n")
    }

    public toString() {
        return this.lines.join("\n")
    }
}
