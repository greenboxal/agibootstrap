const fs = require("fs");
const ts = require("typescript");
const {
    FunctionType,
    SchemaGenerator,
    createFormatter,
    createParser,
    createProgram,
    ObjectType,
    ObjectProperty,
    getKey,
    DefinitionType,
} = require("ts-json-schema-generator");

class FunctionParser {
    constructor(childNodeParser) {
        this.childNodeParser = childNodeParser
    }

    supportsNode(node) {
        console.log('supportsNode', ts.SyntaxKind[node.kind],  node.name && node.name.escapedText)
        return node.kind === ts.SyntaxKind.FunctionType;
    }
    createType(node, context) {
        console.log('createType', node, context)
        const parameterTypes = node.parameters.map((parameter) => {
            return this.childNodeParser.createType(parameter, context);
        });

        const namedArguments = new ObjectType(
            `object-${getKey(node, context)}`,
            [],
            parameterTypes.map((parameterType, index) => {
                // If it's missing a questionToken but has an initializer we can consider the property as not required
                const required = node.parameters[index].questionToken ? false : !node.parameters[index].initializer;

                return new ObjectProperty(node.parameters[index].name.getText(), parameterType, required);
            }),
            false
        );
        return new DefinitionType(this.getTypeName(node, context), namedArguments);
    }

    getTypeName(node) {
        if (ts.isArrowFunction(node) || ts.isFunctionExpression(node)) {
            const parent = node.parent;
            if (ts.isVariableDeclaration(parent)) {
                return `<typeof ${parent.name.getText()}>`;
            }
        }
        if (ts.isFunctionDeclaration(node)) {
            return `<typeof ${node.name.getText()}>`;
        }
        throw new Error("Expected to find a name for function but couldn't");
    }
}

class FunctionTypeFormatter {
    supportsType(type) {
        return type instanceof FunctionType;
    }

    getDefinition(props) {
        if (props instanceof FunctionType) {
            console.log(props)
        }
        return {
            type: "object",
            properties: {
                isFunction: {
                    type: "boolean",
                    const: true,
                },
            },
        };
    }

    getChildren(type) {
        return []
    }
}

class FunctionNodeTypeFormatter {
    supportsType(type) {
        return type instanceof DefinitionType && type.getType() instanceof FunctionType;
    }

    getDefinition(props) {
        if (props instanceof FunctionType) {
            console.log(props)
        }
        return {
            type: "object",
            properties: {
                isFunction: {
                    type: "boolean",
                    const: true,
                },
            },
        };
    }

    getChildren(type) {
        return []
    }
}

const config = {
    path: "./src/index.ts",
    tsconfig: "./tsconfig.json",
    type: "*",
};

const OUTPUT_PATH = "./dist/bundle.d.ts.json";

const program = createProgram(config);
const parser = createParser(program, config, (chainNodeParser) => {
    //chainNodeParser.addNodeParser(new FunctionParser(chainNodeParser));
})
const formatter = createFormatter(config, (fmt, _circularReferenceTypeFormatter) => {
    fmt.addTypeFormatter(new FunctionTypeFormatter());
});
const schema = new SchemaGenerator(program, parser, formatter, config).createSchema(config.type);
const schemaString = JSON.stringify(schema, null, 2);
fs.writeFile(OUTPUT_PATH, schemaString, (err) => {
    if (err) throw err;
});

