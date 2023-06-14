# aigbootstrap

Self bootstrapped general purpose coder AGI (the "G" is joke?).

    Note: I wanted to have a README, but we need to finish writing the code vector searching first.
    
    That way we can also auto generate the README.

## Genesis

This codebase had its inception as a single ChatGPT 4 prompt, in pure markdown, brainstorming about it design.

Then ChatGPT 3.5-4k generated the first few lines of almost working code.
It read each `.go` file, searched for `// TODO:` comments, asked `gpt-3.5-turbo` (4k) to complete them, and wrote the result back to the file.

Then we (the original author and the AI) made it automatically generate a commit message, commit, and push. In an infinite loop.

## AI Safety

    Note (wip): this is a straight out copy of stable diffusion 2.1 model card, with some edits.

The model is intended for research purposes only. Possible research areas and tasks include

Safe deployment of models which have the potential to generate harmful content.
Probing and understanding the limitations and biases of generative models.
Generation of artworks and use in design and other artistic processes.
Applications in educational or creative tools.
Research on generative models.
Excluded uses are described below.


### Misuse, Malicious Use, and Out-of-Scope Use

The model should not be used to intentionally create or disseminate code that create hostile or completely unsupervised code with no checks and balances.

### Out-of-Scope Use

The model was not trained to be factual or true representations of people or events, and therefore using the model to generate such content is out-of-scope for the abilities of this model.


### Misuse and Malicious Use

Using the model to generate content that is cruel to individuals is a misuse of this model. This includes, but is not limited to:

Generating demeaning, dehumanizing, or otherwise harmful representations of people or their environments, cultures, religions, etc.
Intentionally promoting or propagating discriminatory content or harmful stereotypes.
Impersonating individuals without their consent.
Sexual content without consent of the people who might see it.
Mis- and disinformation
Representations of egregious violence and gore
Sharing of copyrighted or licensed material in violation of its terms of use.
Sharing content that is an alteration of copyrighted or licensed material in violation of its terms of use.

## License

AGPL. See [LICENSE.md](LICENSE.md) file.
