import { makeSchema, PrimitiveTypes } from "@psidb/psidb-sdk/client/schema";
import { BalanceNode } from "@psidb/psidb-sdk/types/psidb.jukebox/BalanceNode";
import { PitchBendNode } from "@psidb/psidb-sdk/types/psidb.jukebox/PitchBendNode";
import { PlayNoteNode } from "@psidb/psidb-sdk/types/psidb.jukebox/PlayNoteNode";
import { SetBPMNode } from "@psidb/psidb-sdk/types/psidb.jukebox/SetBPMNode";


export class Command extends makeSchema("psidb.jukebox/Command", {
    Balance: BalanceNode,
    PitchBend: PitchBendNode,
    PlayNote: PlayNoteNode,
    SetBPM: SetBPMNode,
    Timecode: PrimitiveTypes.Float64,
    Volume: PlayNoteNode,
}) {}
