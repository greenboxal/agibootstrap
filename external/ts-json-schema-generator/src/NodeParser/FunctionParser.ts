import ts from "typescript";
import { NodeParser } from "../NodeParser";
import { Context } from "../NodeParser";
import { SubNodeParser } from "../SubNodeParser";
import { ObjectProperty, ObjectType } from "../Type/ObjectType";
import { getKey } from "../Utils/nodeKey";
import { DefinitionType } from "../Type/DefinitionType";
import {FunctionParameter, FunctionType} from "../Type/FunctionType";
import {symbolAtNode} from "../Utils/symbolAtNode";
import {isNodeHidden} from "../Utils/isHidden";
import {BaseType} from "../Type/BaseType";

/**
 * This function parser supports both `FunctionDeclaration` & `ArrowFunction` nodes.
 * This parser will only parse the input parameters.
 * TODO: Parse `ReturnType` of the function?
 */
export class FunctionParser implements SubNodeParser {
    constructor(
        protected typeChecker: ts.TypeChecker,
        protected childNodeParser: NodeParser,
    ) {}

    public supportsNode(node: ts.ArrowFunction | ts.FunctionDeclaration | ts.FunctionExpression): boolean {
        if (node.kind === ts.SyntaxKind.FunctionDeclaration) {
            // Functions needs a name for us to include it in the json schema
            return Boolean(node.name);
        }
        // We can figure out the name of arrow functions if their parent is a variable declaration
        return (
            (node.kind === ts.SyntaxKind.ArrowFunction || node.kind === ts.SyntaxKind.FunctionExpression) &&
            ts.isVariableDeclaration(node.parent)
        );
    }
    public createType(node: ts.FunctionDeclaration | ts.ArrowFunction, context: Context): FunctionType {
        this.pushParameters(node, context);

        return new FunctionType(
            this.getTypeName(node, context),
            [],
            this.getParameters(node, context),
            this.getAdditionalParameters(node, context),
            this.childNodeParser.createType(node.type!, context)!
        );
    }

    public getTypeName(node: ts.FunctionDeclaration | ts.ArrowFunction, context: Context): string {
        const fullName = node.name?.getText() || `function-${node.getFullStart()}`;

        const argumentIds = context.getArguments().map((arg) => arg!.getId());

        return argumentIds.length ? `${fullName}<${argumentIds.join(",")}>` : fullName;
    }

    private getParameters(node: ts.FunctionDeclaration | ts.ArrowFunction, context: Context): FunctionParameter[] {
        return node.parameters.filter(ts.isParameter).reduce((result: FunctionParameter[], parameterNode) => {
            const parameterSymbol = symbolAtNode(parameterNode);
            if (!parameterSymbol) return result;
            if (isNodeHidden(parameterNode)) {
                return result;
            }
            const newContext = new Context(parameterNode);
            this.pushParameters(node, newContext);
//            newContext.isParameter = true;
            const objectParameter: FunctionParameter = new FunctionParameter(
                parameterSymbol.getName(),
                this.childNodeParser.createType(parameterNode.type!, newContext)!,
                !parameterNode.questionToken
            );

            result.push(objectParameter);
            return result;
        }, []);
    }
    private getAdditionalParameters(node: ts.FunctionDeclaration | ts.ArrowFunction, context: Context): BaseType | false {
        const indexSignature = node.parameters.find(ts.isIndexSignatureDeclaration);
        if (!indexSignature) {
            return false;
        }

        return this.childNodeParser.createType(indexSignature.type!, context)!;
    }

    private pushParameters(node: ts.FunctionDeclaration | ts.ArrowFunction, context: Context) {
        if (node.typeParameters && node.typeParameters.length) {
            node.typeParameters.forEach((typeParam) => {
                const nameSymbol = this.typeChecker.getSymbolAtLocation(typeParam.name)!;
                context.pushParameter(nameSymbol.name);

                let type;

                if (typeParam.default) {
                    type = this.childNodeParser.createType(typeParam.default, context);
                }

                context.setDefault(nameSymbol.name, type!);
            });
        }
    }
}
