# Jukebox

## Master Prompt

```
Here are some example commands for the defined grammar:
SetBPM 120
SetBPM 90
PlayNote 4 C 2 - Play the note C in the 4th octave for 2 units of time at default velocity of.
PlayNote 5 D# 3 - Play the note D# in the 5th octave for 3 units of time at default velocity of 1.
PlayNote 3 F# 4 0.97 - Play the note F# in the 3rd octave for 4 units of time at a velocity of 0.80.
PlayNote 6 Bb 2 0.95 - Play the note Bb in the 6th octave for 2 units of time at a velocity of 0.95.
PlayNote 2 E 1 - Play the note E in the 2nd octave for 1 unit of time at default velocity.
SetVolume 0.5 - Set the volume to 0.5.
SetVolume 1 - Set the volume to 1.
SetBalance 0.5 - Set the balance to 0.5.
SetBalance -1 - Set the balance to -1.
PitchBend -0.5 - Pitch bend by -0.5.
PitchBend 0 - Pitch bend by 0 (neutral).
PitchBend 0.5 - Pitch bend by 0.5.
PitchBend 1 - Pitch bend by 1.
Combining multiple commands in CommandSheet:
SetBPM 100
PlayNote 4 C 2
PlayNote 5 G 3 85
PlayNote 6 E 1
SetBPM 80
PlayNote 4 A# 2 90
For these examples:
The SetBPM simply takes an integer indicating the desired BPM.
The PlayNote starts with the keyword PlayNote, followed by:
The octave (an integer).
The note name (one of C, D, E, F, G, A, B).
An optional accidental (either # for sharp or b for flat).
The duration (a floating-point number or an integer indicating the time for which the note should be played).
An optional velocity (a floating-point number ranging from 0 to 1 indicating the strength or volume of the note).
Remember, the CommandSheet structure accepts multiple commands, so a valid sheet may contain any number of commands in any order.


Use PlayCommandSheet to play Stairway to Heaven by Led Zeppelin
```

```
Generate a continuous stream of commands for a music synthesizer using the specified grammar.

Commands start with "@" and a timestamp. They include nodes for setting BPM, volume, balance, pitch bending, or playing a note. Numbers can be negative, decimal, or in scientific notation.

Separate each command with a newline. Start now.```