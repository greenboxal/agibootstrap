import {ISchemaInstance, makeSchema, PrimitiveTypes, SchemaClass, SchemaDefinition, SchemaInstance, TypeSymbol} from "./schema";

export function Schema<T extends {
    new(...args: any[]): any,
}>() {
    return (target: T, other) => {
        return class extends target {

        }
    }
}

export class NodeBase<TSchema extends SchemaDefinition> {
    private _valid: boolean = false;

    declare state: ISchemaInstance<any, TSchema>;

    protected onUpdate() {

    }

    invalidate() {
        this._valid = false
    }

    update() {
        if (this._valid) {
            return
        }

        this.onUpdate()

        this._valid = true
    }
}

export function makeNode<TSchema extends SchemaDefinition>(schema: SchemaClass<any, TSchema>) {
    return class extends NodeBase<TSchema> {
        declare state: InstanceType<typeof schema>;

        get schema() { return schema }


        constructor() {
            super()

            this.state = new schema()
        }
    }
}

export const TestNodeSchema = makeSchema({
    [TypeSymbol]: "psidb.test.TestNode",
    name: PrimitiveTypes.String,
})

export class TestNode extends makeNode(TestNodeSchema) {

    onUpdate() {

    }
}

const a = new TestNode()
a