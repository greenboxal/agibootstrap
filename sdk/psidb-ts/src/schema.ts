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

import {Schema as JsonSchema} from "jsonschema";
import * as TJS from "typescript-json-schema";
import fs from "fs";

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

    compileSchema2(): JsonSchema {
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

        return schema as JsonSchema
    }

    compileSchema(): JsonSchema {
        const schema = this.generator.createSchema("ModuleInterface")
        const schemaString = JSON.stringify(schema)
        const parsedSchema = JSON.parse(schemaString) as JsonSchema

        if (!parsedSchema.$id) {
            parsedSchema.$id = this.config.schemaId
        }

        return parsedSchema
    }
}
