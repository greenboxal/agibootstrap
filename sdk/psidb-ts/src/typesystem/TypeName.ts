import * as path from "path";

export class TypeName {
    public pkg: string = ""
    public name: string = ""
    public inParameters: TypeName[] = []
    public outParameters: TypeName[] = []

    get isGeneric(): boolean {
        return !(this.pkg == "" && this.name == "func") && this.inParameters.length > 0
    }

    constructor(props?: Partial<TypeName>){
        if (props) {
            Object.assign(this, props)
        }
    }

    static fromString(name: string): TypeName {
        return parseMangledName(name)
    }

    withInParameters(inParameters: TypeName[]): TypeName {
        return new TypeName({
            pkg: this.pkg,
            name: this.name,
            inParameters: inParameters,
            outParameters: this.outParameters,
        })
    }

    withOutParameters(outParameters: TypeName[]): TypeName {
        return new TypeName({
            pkg: this.pkg,
            name: this.name,
            inParameters: this.inParameters,
            outParameters: outParameters,
        })
    }

    generic(): TypeName {
        return new TypeName({
            pkg: this.pkg,
            name: this.name,
        })
    }

    equals(other?: TypeName): boolean {
        if (!other) {
            return false
        }

        if (this === other) {
            return true
        }

        if (this.pkg !== other.pkg) {
            return false
        }

        if (this.name !== other.name) {
            return false
        }

        if (this.inParameters.length !== other.inParameters.length) {
            return false
        }

        if (this.outParameters.length !== other.outParameters.length) {
            return false
        }

        for (let i = 0; i < this.inParameters.length; i++) {
            if (!this.inParameters[i].equals(other.inParameters[i])) {
                return false
            }
        }

        for (let i = 0; i < this.outParameters.length; i++) {
            if (!this.outParameters[i].equals(other.outParameters[i])) {
                return false
            }
        }

        return true
    }

    toString() {
        return mangleName(this)
    }

    pickNested(path: number[]): TypeName {
        if (path.length == 0) {
            return this
        }

        const [head, ...tail] = path

        if (head < 0) {
            throw new Error("Invalid path")
        }

        if (head < this.inParameters.length) {
            return this.inParameters[head].pickNested(tail)
        }

        if (head < this.inParameters.length + this.outParameters.length) {
            return this.outParameters[head - this.inParameters.length].pickNested(tail)
        }

        throw new Error("Invalid path")
    }
}

function _parseFullyQualifiedName(mangledName: string, pos: number): [string, string, number] {
    let fullName = ""
    let newPos = pos

    while (mangledName.length - newPos > 0) {
        if (mangledName[newPos] === "(" || mangledName[newPos] === ")" || mangledName[newPos] === ";") {
            break
        }

        fullName += mangledName[newPos]
        newPos++
    }

    const parts = fullName.split("/")
    const pkg = parts.slice(0, parts.length - 1).join("/")
    const name = parts[parts.length - 1]

    return [pkg, name, newPos]
}

function _parseArgumentList(mangledName: string, pos: number): [TypeName[], number] {
    let newPos = pos
    const result: TypeName[] = []

    while (mangledName.length - newPos > 0) {
        if (mangledName[newPos] === ";") {
            newPos++
            continue
        } else if (mangledName[newPos] === ")") {
            break
        }

        const [tn, postTnPos] = _parseMangledName(mangledName, newPos)

        result.push(new TypeName(tn))
        newPos = postTnPos
    }


    return [result, newPos]
}

function _parseMangledName(mangledName: string, pos: number): [Partial<TypeName>, number] {
    const tn: Partial<TypeName> = {
        pkg: "",
        name: "",
    }

    let [pkg, name, postFqnPos] = _parseFullyQualifiedName(mangledName, pos);

    tn.pkg = pkg
    tn.name = name

    if (mangledName.length - postFqnPos <= 0) {
        return [tn, postFqnPos]
    }

    const nextChar = mangledName[postFqnPos]

    if (nextChar !== "(") {
        return [tn, postFqnPos]
    }

    const [inArgs, postInArgsPos] = _parseArgumentList(mangledName, postFqnPos + 1)

    if (mangledName.length - postInArgsPos <= 0) {
        throw new Error("Unexpected end of mangled name, expected ')' or ';'")
    }

    const nextChar2 = mangledName[postInArgsPos]

    if (nextChar2 !== ")") {
        throw new Error("Unexpected end of mangled name, expected ')' or ';'")
    }

    tn.inParameters = inArgs

    if (mangledName.length - postInArgsPos <= 0) {
        return [tn, postInArgsPos]
    }

    const [outArgs, postOutArgsPos] = _parseArgumentList(mangledName, postInArgsPos + 1)

    tn.outParameters = outArgs

    return [tn, postOutArgsPos]
}

export function parseMangledName(mangledName: string): TypeName {
    const [typeName, _index] = _parseMangledName(mangledName, 0);
    return new TypeName(typeName);
}

export function mangleName(n: TypeName): string {
    const base = path.join(n.pkg, n.name);
    let args = "";

    if (n.inParameters?.length || n.outParameters?.length) {
        const inParams = n.inParameters?.map(arg => mangleName(arg)).join(";") || "";
        const outParams = n.outParameters?.map(arg => mangleName(arg)).join(";") || "";
        args = `(${inParams})${outParams}`;
    }

    return `${base}${args}`;
}

