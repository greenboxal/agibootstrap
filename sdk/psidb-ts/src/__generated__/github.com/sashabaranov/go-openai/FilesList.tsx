import { makeSchema, ArrayOf } from "@psidb/psidb-sdk/client/schema";
import { File } from "@psidb/psidb-sdk/types/github.com/sashabaranov/go-openai/File";


export class FilesList extends makeSchema("github.com/sashabaranov/go-openai/FilesList", {
    data: ArrayOf(File),
}) {}
