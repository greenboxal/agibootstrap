# aigbootstrap

Self bootstrapped general purpose coder.

## Genesis

This codebase had its inception as a single ChatGPT 4 prompt, in pure markdown, brainstorming about it design.

Then ChatGPT 3.5-4k generated the first few lines of almost working code.
It read each `.go` file, searched for `// TODO:` comments, asked `gpt-3.5-turbo` (4k) to complete them, and wrote the result back to the file.

Then we (the original author and the AI) made it automatically generate a commit message, commit, and push. In an infinite loop.

## License

AGPL. See [LICENSE.md](LICENSE.md) file.