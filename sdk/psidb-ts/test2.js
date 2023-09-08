const fs = require("fs");
const ts = require("typescript");
const TJS = require("typescript-json-schema");

const settings = {
    required: true,
};

const compilerOptions = {
    strictNullChecks: true,
};

const basePath = "./src";
const OUTPUT_PATH = "./dist/bundle.d.ts.json";

const program = TJS.getProgramFromFiles(
    ["./src/index.ts"],
    compilerOptions,
    basePath
);

const generator = TJS.buildGenerator(program, settings);

const symbols = generator.getMainFileSymbols(program);
const schema = generator.getSchemaForSymbols(symbols);
const schemaString = JSON.stringify(schema, null, 2);

fs.writeFile(OUTPUT_PATH, schemaString, (err) => {
    if (err) throw err;
});