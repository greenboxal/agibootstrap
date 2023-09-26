import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { RawMessage } from "@psidb/psidb-sdk/types/encoding/json/RawMessage";
import { uint8 } from "@psidb/psidb-sdk/types//uint8";


export class VirtualNode extends makeSchema("psidb.vm/VirtualNode", {
    data: RawMessage(uint8),
    uuid: PrimitiveTypes.String,
}) {}
