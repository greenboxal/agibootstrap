import { makeInterface } from "@psidb/psidb-sdk/client/iface";
import { PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { error } from "@psidb/psidb-sdk/types//error";
import { Iterator } from "@psidb/psidb-sdk/types/agib.platform/stdlib/iterators/Iterator";
import { JournalEntry } from "@psidb/psidb-sdk/types/psidb/JournalEntry";


export const Journal = makeInterface({
    name: "psidb/Journal",
    methods: {
        Close: PrimitiveTypes.Func()(error),
        GetHead: PrimitiveTypes.Func()(PrimitiveTypes.UnsignedInteger, error),
        Iterate: PrimitiveTypes.Func(PrimitiveTypes.UnsignedInteger, PrimitiveTypes.Integer)(Iterator(JournalEntry)),
        Read: PrimitiveTypes.Func(PrimitiveTypes.UnsignedInteger, PrimitiveTypes.Pointer(JournalEntry))(PrimitiveTypes.Pointer(JournalEntry)(error)),
        Write: PrimitiveTypes.Func(PrimitiveTypes.Pointer(JournalEntry))(error),
    },
});
