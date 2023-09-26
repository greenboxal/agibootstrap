import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { Position } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/utils/sparsing/Position";
import { TokenKind } from "@psidb/psidb-sdk/types/github.com/greenboxal/agibootstrap/psidb/utils/sparsing/TokenKind";


export const IToken = makeInterface({
    name: "github.com/greenboxal/agibootstrap/psidb/utils/sparsing/IToken",
    methods: {
        GetEnd: PrimitiveTypes.Func()(Position),
        GetIndex: PrimitiveTypes.Func()(PrimitiveTypes.Integer),
        GetKind: PrimitiveTypes.Func()(TokenKind),
        GetStart: PrimitiveTypes.Func()(Position),
        GetText: PrimitiveTypes.Func()(PrimitiveTypes.String),
        GetValue: PrimitiveTypes.Func()(PrimitiveTypes.String),
    },
});
