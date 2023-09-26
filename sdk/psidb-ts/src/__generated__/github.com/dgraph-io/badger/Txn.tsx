import { makeSchema } from "@psidb/psidb-sdk/client/schema";


export class Txn extends makeSchema("github.com/dgraph-io/badger/Txn", {
}) {}
