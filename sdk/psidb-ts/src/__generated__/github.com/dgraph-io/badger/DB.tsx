import { makeSchema } from "@psidb/psidb-sdk/client/schema";


export class DB extends makeSchema("github.com/dgraph-io/badger/DB", {
}) {}
